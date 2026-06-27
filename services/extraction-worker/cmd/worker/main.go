package main

import (
	"context"
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
	"github.com/therealagt/ContractManagementTool/libs/common/vertex"
	"github.com/therealagt/ContractManagementTool/services/extraction-worker/internal/config"
	"github.com/therealagt/ContractManagementTool/services/extraction-worker/internal/handler"
	"github.com/therealagt/ContractManagementTool/services/extraction-worker/internal/pipeline"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"database/sql"
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

	loader, loaderCleanup := mustLoader(ctx, settings)
	defer loaderCleanup()

	extractor, err := vertex.NewExtractor(ctx, settings.GCPProjectID, settings.GCPRegion, settings.GeminiModel, settings.PromptVersion)
	if err != nil {
		log.Fatalf("vertex: %v", err)
	}

	p := pipeline.New(db, loader, extractor)
	h := handler.New(p)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Post("/pubsub/extraction", h.ServeHTTP)

	addr := envOr("PORT", "8080")
	server := &http.Server{Addr: ":" + addr, Handler: r, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		log.Printf("extraction worker listening on :%s", addr)
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
	if settings.IsSQLite() {
		driver = "sqlite"
	}
	db, err := sql.Open(driver, settings.DatabaseDSN())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func mustLoader(ctx context.Context, settings *config.Settings) (pipeline.PDFLoader, func()) {
	if settings.GCSStagingBucket != "" {
		c, err := gcs.NewClient(ctx, settings.GCSStagingBucket)
		if err != nil {
			log.Fatalf("gcs: %v", err)
		}
		return c, func() { _ = c.Close() }
	}
	local, err := gcs.NewLocalClient(".local-gcs", "local-staging")
	if err != nil {
		log.Fatalf("local gcs: %v", err)
	}
	return local, func() {}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
