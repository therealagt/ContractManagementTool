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

	"github.com/therealagt/ContractManagementTool/libs/common/migrate"
	"github.com/therealagt/ContractManagementTool/services/weekly-report/internal/config"
	"github.com/therealagt/ContractManagementTool/services/weekly-report/internal/handler"
	"github.com/therealagt/ContractManagementTool/services/weekly-report/internal/pipeline"

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

	mailer := pipeline.NewMailerFromConfig(
		settings.SMTPHost,
		settings.SMTPPort,
		settings.SMTPUser,
		settings.SMTPPassword,
		settings.EmailFrom,
	)
	p := pipeline.New(db, mailer, settings.ReviewSLADays, settings.Environment, settings.Recipients())
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
		log.Printf("weekly report service listening on :%s", addr)
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

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
