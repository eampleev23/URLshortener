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

/*
Добавьте в код сервера новый эндпоинт POST /api/shorten, который будет принимать в теле запроса JSON-объект
{"url":"<some_url>"} и возвращать в ответ объект {"result":"<short_url>"}.
Запрос может иметь такой вид:

POST http://localhost:8080/api/shorten HTTP/1.1
Host: localhost:8080
Content-Type: application/json
{
  "url": "https://practicum.yandex.ru"
}

Ответ может быть таким:

HTTP/1.1 201 OK
Content-Type: application/json
Content-Length: 30
{
 "result": "http://localhost:8080/EwHXdJfB"
}

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
	r.Post("/api/shorten", h.JSONHandler)

	err := http.ListenAndServe(c.GetValueByIndex("runaddr"), r)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
