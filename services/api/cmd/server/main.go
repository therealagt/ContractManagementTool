package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/therealagt/ContractManagementTool/libs/common/gcs"
	"github.com/therealagt/ContractManagementTool/libs/common/pubsub"
	"github.com/therealagt/ContractManagementTool/services/api/internal/auth"
	"github.com/therealagt/ContractManagementTool/services/api/internal/config"
	"github.com/therealagt/ContractManagementTool/services/api/internal/db"
	"github.com/therealagt/ContractManagementTool/services/api/internal/routes"
	"github.com/therealagt/ContractManagementTool/services/api/internal/services"
)

func main() {
	settings := config.Load()
	if err := settings.ValidateSecurity(); err != nil {
		log.Fatalf("security config: %v", err)
	}

	ctx := context.Background()
	database, err := db.Open(settings)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(ctx, database, settings); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	staging, stagingCleanup := mustStaging(ctx, settings)
	defer stagingCleanup()

	publisher, publisherCleanup := mustPublisher(ctx, settings, settings.PubSubExtractionTopic, "extraction")
	defer publisherCleanup()

	archivePublisher, archiveCleanup := mustArchivePublisher(ctx, settings)
	defer archiveCleanup()

	uploads := services.NewUploadService(database, staging, publisher)
	reviews := services.NewReviewService(database, archivePublisher)
	validator := auth.NewIAPValidator(settings)
	accessLogger := newAccessLoggerFactory(database)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	routes.Mount(r, settings, validator, accessLogger)
	routes.MountContracts(r, settings, validator, uploads, accessLogger)
	routes.MountReview(r, settings, validator, reviews, accessLogger)

	addr := envOr("PORT", "8080")
	server := &http.Server{
		Addr:              ":" + addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func mustStaging(ctx context.Context, settings *config.Settings) (services.StagingStorage, func()) {
	if settings.GCSStagingBucket != "" && settings.GCPProjectID != "" {
		client, err := gcs.NewClient(ctx, settings.GCSStagingBucket)
		if err != nil {
			log.Fatalf("gcs: %v", err)
		}
		return client, func() { _ = client.Close() }
	}
	local, err := gcs.NewLocalClient(".local-gcs", "local-staging")
	if err != nil {
		log.Fatalf("local staging: %v", err)
	}
	log.Printf("using local staging (set GCS_STAGING_BUCKET for GCP)")
	return local, func() {}
}

func mustPublisher(ctx context.Context, settings *config.Settings, topic, label string) (services.ExtractionPublisher, func()) {
	if topic == "" || settings.GCPProjectID == "" {
		log.Printf("pubsub %s publisher disabled (set topic env for pipeline)", label)
		return nil, func() {}
	}
	p, err := pubsub.NewPublisher(ctx, settings.GCPProjectID, topic)
	if err != nil {
		log.Fatalf("pubsub %s: %v", label, err)
	}
	return p, func() { _ = p.Close() }
}

func mustArchivePublisher(ctx context.Context, settings *config.Settings) (services.ArchivePublisher, func()) {
	if settings.PubSubArchiveTopic == "" || settings.GCPProjectID == "" {
		log.Printf("pubsub archive publisher disabled (set PUBSUB_ARCHIVE_TOPIC for archive pipeline)")
		return nil, func() {}
	}
	p, err := pubsub.NewPublisher(ctx, settings.GCPProjectID, settings.PubSubArchiveTopic)
	if err != nil {
		log.Fatalf("pubsub archive: %v", err)
	}
	return p, func() { _ = p.Close() }
}

func newAccessLoggerFactory(database *sql.DB) func(*http.Request) *auth.AccessLogger {
	return func(r *http.Request) *auth.AccessLogger {
		user, _ := auth.UserFromContext(r.Context())
		ip := auth.ClientIP(r)
		return auth.NewAccessLogger(database, user, ip)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
