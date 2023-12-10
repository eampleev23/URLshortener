package main

import (
	"github.com/eampleev23/URLshortener/internal/compression"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/handlers"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	c, err := config.NewConfig()
	if err != nil {
		return err
	}
	s := store.NewStore(c)
	s.ReadStoreFromFile(c)
	myLog, _ := logger.NewZapLogger("info")
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
