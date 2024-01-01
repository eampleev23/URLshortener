package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eampleev23/URLshortener/internal/compression"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/handlers"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	myLog, err := logger.NewZapLogger("info")
	if err != nil {
		return fmt.Errorf("failed to initialize a new logger: %w", err)
	}

	c, err := config.NewConfig(myLog)
	if err != nil {
		return fmt.Errorf("failed to initialize a new config: %w", err)
	}

	s, err := store.NewStorage(c, myLog)
	if err != nil {
		return fmt.Errorf("failed to initialize a new store: %w", err)
	}

	if len(c.DBDSN) != 0 {
		// Отложенно закрываем соединение с бд.
		defer func() {
			if err := s.Close(); err != nil {
				myLog.ZL.Info("new store failed to properly close the DB connection")
			}
		}()
	}

	h := handlers.NewHandlers(s, c, myLog)

	myLog.ZL.Info("Running server", zap.String("address", c.RanAddr))
	r := chi.NewRouter()
	r.Use(myLog.RequestLogger)
	r.Use(compression.GzipMiddleware)
	r.Post("/", h.CreateShortLink)
	r.Get("/ping", h.PingDBHandler)
	r.Get("/{id}", h.UseShortLink)
	r.Post("/api/shorten", h.JSONHandler)
	r.Post("/api/shorten/batch", h.JSONHandlerBatch)
	r.Get("/api/user/urls", h.GetURLsByUserID)

	err = http.ListenAndServe(c.RanAddr, r)
	if err != nil {
		return fmt.Errorf("ошибка ListenAndServe: %w", err)
	}
	return nil
}
