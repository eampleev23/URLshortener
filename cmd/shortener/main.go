// iter6
package main

import (
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
	c := config.NewConfig()
	c.SetValues()
	s := store.NewStore()
	h := handlers.NewHandlers(s, c)

	if err := logger.Initialize("info"); err != nil {
		return err
	}
	logger.Log.Info("Running server", zap.String("address", c.GetValueByIndex("runaddr")))

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Post("/", h.CreateShortLink)
	r.Get("/{id}", h.UseShortLink)

	err := http.ListenAndServe(c.GetValueByIndex("runaddr"), r)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
