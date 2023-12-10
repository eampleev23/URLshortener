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
	"go.uber.org/zap"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	c, err := config.NewConfig()
	myLog, _ := logger.NewZapLogger("info")
	if err != nil {
		myLog.ZL.Info("Ошибка при создании конфига", zap.Error(err))
		return fmt.Errorf("%w", err)
	}

	s, err := store.NewStore(c, myLog)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	if c.GetValueByIndex("sfilepath") != "" {
		s.ReadStoreFromFile(c)
	}

	h := handlers.NewHandlers(s, c, myLog)

	myLog.ZL.Info("Running server", zap.String("address", c.GetValueByIndex("runaddr")))
	r := chi.NewRouter()
	r.Use(myLog.RequestLogger)
	r.Use(compression.GzipMiddleware)
	r.Post("/", h.CreateShortLink)
	r.Get("/{id}", h.UseShortLink)
	r.Post("/api/shorten", h.JSONHandler)

	err = http.ListenAndServe(c.GetValueByIndex("runaddr"), r)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
