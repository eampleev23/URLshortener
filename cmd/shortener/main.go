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

/*
Задание по треку «Сервис сокращения URL»

Добавьте поддержку gzip в ваш сервис. Научите его:
Принимать запросы в сжатом формате (с HTTP-заголовком Content-Encoding).
Отдавать сжатый ответ клиенту, который поддерживает обработку сжатых ответов (с HTTP-заголовком Accept-Encoding).
Функция сжатия должна работать для контента с типами application/json и text/html.
Вспомните middleware из урока про HTTP-сервер, это может вам помочь.
*/

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	c := config.NewConfig()
	c.SetValues()
	s := store.NewStore(c)
	s.ReadStoreFromFile()
	h := handlers.NewHandlers(s, c)

	if err := logger.Initialize("info"); err != nil {
		return err
	}
	logger.Log.Info("Running server", zap.String("address", c.GetValueByIndex("runaddr")))

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(compression.GzipMiddleware)
	r.Post("/", h.CreateShortLink)
	r.Get("/{id}", h.UseShortLink)
	r.Post("/api/shorten", h.JSONHandler)

	err := http.ListenAndServe(c.GetValueByIndex("runaddr"), r)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
