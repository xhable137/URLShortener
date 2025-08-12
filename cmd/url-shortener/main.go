package main

import (
	"URLShortener/internal/config"
	"URLShortener/internal/http-server/handlers/redirect"
	"URLShortener/internal/http-server/handlers/url/save"
	"URLShortener/internal/lib/logger/handlers/slogpretty"
	"URLShortener/internal/lib/storage"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal string = "local"
	envDev   string = "dev"
	envProd  string = "prod"
)

func main() {
	cfg := config.MustLoad()

	var log = setupLogger(cfg.Env)

	log.Info("Starting URL Shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// Initialize PostgreSQL storage
	postgresStorage, err := storage.NewPostgresStorage(cfg)
	if err != nil {
		log.Error("Failed to initialize PostgreSQL storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer postgresStorage.Close()

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.NewPostgresStorage(log, postgresStorage))
	})

	router.Get("/{alias}", redirect.New(log, postgresStorage))

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started", slog.String("address", srv.Addr))
	<-done
	log.Info("shutting down server", slog.String("address", srv.Addr))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown server", slog.String("error", err.Error()))
		return
	}
	log.Info("server stopped", slog.String("address", srv.Addr))
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
