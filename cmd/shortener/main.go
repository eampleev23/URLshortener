package main

import (
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
	s := store.NewStore()
	h := handlers.NewHandlers(s)

	if err := logger.Initialize("info"); err != nil {
		return err
	}
	logger.Log.Info("Running server", zap.String("address", ":8080"))

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Post("/", h.CreateShortLink)
	r.Get("/{id}", h.UseShortLink)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
