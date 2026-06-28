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
	"github.com/therealagt/ContractManagementTool/libs/common/migrate"
	"github.com/therealagt/ContractManagementTool/services/archive-worker/internal/config"
	"github.com/therealagt/ContractManagementTool/services/archive-worker/internal/handler"
	"github.com/therealagt/ContractManagementTool/services/archive-worker/internal/pipeline"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func main() {
	settings := config.Load()
	ctx := context.Background()

	db, err := openDB(settings)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	if err := migrate.Run(ctx, db, settings.IsSQLite()); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	staging, archive, cleanup := mustStorage(ctx, settings)
	defer cleanup()

	p := pipeline.New(db, staging, archive, settings.RetentionYears)
	h := handler.New(p)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Post("/pubsub/archive", h.ServeHTTP)

	addr := envOr("PORT", "8080")
	server := &http.Server{Addr: ":" + addr, Handler: r, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		log.Printf("archive worker listening on :%s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func openDB(settings *config.Settings) (*sql.DB, error) {
	driver := "pgx"
	dsn := settings.DatabaseDSN()
	if settings.IsSQLite() {
		driver = "sqlite"
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func mustStorage(ctx context.Context, settings *config.Settings) (pipeline.StagingReader, pipeline.ArchiveWriter, func()) {
	if settings.GCSStagingBucket != "" && settings.GCSArchiveBucket != "" && settings.GCPProjectID != "" {
		staging, err := gcs.NewClient(ctx, settings.GCSStagingBucket)
		if err != nil {
			log.Fatalf("staging gcs: %v", err)
		}
		archive, err := gcs.NewClient(ctx, settings.GCSArchiveBucket)
		if err != nil {
			_ = staging.Close()
			log.Fatalf("archive gcs: %v", err)
		}
		return staging, archive, func() {
			_ = staging.Close()
			_ = archive.Close()
		}
	}

	root := ".local-gcs"
	staging, err := gcs.NewLocalClient(root, "local-staging")
	if err != nil {
		log.Fatalf("local staging: %v", err)
	}
	archive, err := gcs.NewLocalClient(root, "local-archive")
	if err != nil {
		log.Fatalf("local archive: %v", err)
	}
	log.Printf("using local GCS storage (set GCS_*_BUCKET for GCP)")
	return staging, archive, func() {}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
