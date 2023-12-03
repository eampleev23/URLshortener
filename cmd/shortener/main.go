// изменения для коммита
package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

var linksCouples = map[string]string{
	"shortlink": "longlink",
}

func run(appConfig AppConfig) error {
	log.Printf("running server on %s", appConfig.flagRunAddr)
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
