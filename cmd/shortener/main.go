package main

import (
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"log"
	"net/http"
)

var linksCouples = map[string]string{
	"shortlink": "longlink",
}

func run(appConfig AppConfig) error {
	if err := logger.Initialize(appConfig.flagLogLevel); err != nil {
		return err
	}

	logger.Log.Info("Running server", zap.String("address", appConfig.flagRunAddr))
	r := chi.NewRouter()
	r.Post("/", createShortLink)
	r.Get("/{id}", useShortLink)
	return http.ListenAndServe(appConfig.flagRunAddr, r)
}

func main() {
	// обрабатываем флаги
	appConfig := getAppConfig()
	if err := run(appConfig); err != nil {
		log.Fatal(err)
	}
}
