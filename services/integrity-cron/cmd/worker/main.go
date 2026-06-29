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

	"github.com/therealagt/ContractManagementTool/libs/common/alerts"
	"github.com/therealagt/ContractManagementTool/libs/common/gcs"
	"github.com/therealagt/ContractManagementTool/libs/common/migrate"
	"github.com/therealagt/ContractManagementTool/libs/common/monitoring"
	"github.com/therealagt/ContractManagementTool/services/integrity-cron/internal/config"
	"github.com/therealagt/ContractManagementTool/services/integrity-cron/internal/handler"
	"github.com/therealagt/ContractManagementTool/services/integrity-cron/internal/pipeline"

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

	archive, archiveCleanup := mustArchive(ctx, settings)
	defer archiveCleanup()

	alertRecorder := mustAlerts(ctx, settings, db)
	metrics := mustMetrics(ctx, settings)

	p := pipeline.New(db, archive, alertRecorder, metrics, settings.ReviewSLADays)
	h := handler.New(p)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Post("/run", h.ServeHTTP)

	addr := envOr("PORT", "8080")
	server := &http.Server{Addr: ":" + addr, Handler: r, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		log.Printf("integrity cron listening on :%s", addr)
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

func mustArchive(ctx context.Context, settings *config.Settings) (pipeline.ArchiveReader, func()) {
	if settings.GCSArchiveBucket != "" && settings.GCPProjectID != "" {
		client, err := gcs.NewClient(ctx, settings.GCSArchiveBucket)
		if err != nil {
			log.Fatalf("archive gcs: %v", err)
		}
		return client, func() { _ = client.Close() }
	}
	client, err := gcs.NewLocalClient(".local-gcs", "local-archive")
	if err != nil {
		log.Fatalf("local archive: %v", err)
	}
	log.Printf("using local archive storage (set GCS_ARCHIVE_BUCKET for GCP)")
	return client, func() {}
}

func mustAlerts(ctx context.Context, settings *config.Settings, db *sql.DB) *alerts.Recorder {
	var bq *alerts.BigQuerySink
	if settings.GCPProjectID != "" && settings.BigQueryDataset != "" {
		sink, err := alerts.NewBigQuerySink(ctx, settings.GCPProjectID, settings.BigQueryDataset)
		if err != nil {
			log.Printf("bigquery alerts disabled: %v", err)
		} else {
			bq = sink
		}
	}
	return alerts.NewRecorder(db, bq)
}

func mustMetrics(ctx context.Context, settings *config.Settings) pipeline.MetricPublisher {
	if settings.GCPProjectID == "" {
		return monitoring.NoopPublisher{}
	}
	pub, err := monitoring.NewPublisher(ctx, settings.GCPProjectID)
	if err != nil {
		log.Printf("monitoring disabled: %v", err)
		return monitoring.NoopPublisher{}
	}
	return pub
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
