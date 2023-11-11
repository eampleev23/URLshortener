package main

import (
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var linksCouples = map[string]string{
	"EwHXdJfB": "https://practicum.yandex.ru/",
	"AwHXzJsN": "ampleev.com",
	"pzfkq5d":  "test.com",
}

func createShortLink(res http.ResponseWriter, req *http.Request) {
	log.Printf("запустили createShortLink")
	var longLink string
	if b, err := io.ReadAll(req.Body); err == nil {
		longLink = string(b)
	}
	log.Printf("longLink = %s", longLink)
	shortLink := generateUniqShortLink()
	log.Printf("shortLink = %s", shortLink)
	linksCouples[shortLink] = longLink
	log.Printf("After party: %s", linksCouples[shortLink])
	log.Printf("linksCouples = %s", linksCouples)

	// установить статус 201
	res.WriteHeader(201)

	// установить тип контента text/plain
	res.Header().Set("content-type", "text/plain")

	// добавить в качестве ответа сокращенный урл
	shortLinkWithPrefix := "http://localhost:8080/" + shortLink
	res.Write([]byte(shortLinkWithPrefix))
}

func useShortLink(res http.ResponseWriter, req *http.Request) {
	// Дальше проверяем есть ли такой урл в базе
	log.Printf("Запустили useShortLink")
	var shortLink string
	shortLink = chi.URLParam(req, "id")
	log.Printf("shortLink = %s", shortLink)
	result := searchShortLink(shortLink)
	log.Printf("result =`%s`, shortlik = `%s`, linksCouples = `%s`", result, shortLink, linksCouples)
	if result == "no match" {
		res.WriteHeader(http.StatusBadRequest)
	} else {
		res.WriteHeader(307)
		res.Header().Set("Location", linksCouples[shortLink])
	}
}

// Вспомогательная функция для поиска совпадений по урлам в базе
func searchShortLink(shortLink string) string {
	log.Printf("start searchShortLink..")
	log.Printf("shortLink get is %s", shortLink)
	log.Printf("linksCouples[%s] = %s", shortLink, linksCouples[shortLink])
	log.Printf("linksCouples[%s] = %s", "EwHXdJfB", linksCouples["EwHXdJfB"])
	if c, ok := linksCouples[shortLink]; ok {
		return c
	}
	return "no match"
}

// Вспомогательная функция для генерации уникальной короткой ссылки
func generateUniqShortLink() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
		"abcdefghijklmnopqrstuvwxyzåäö" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String() // Например "ExcbsVQs"
	return str
}

func main() {
	r := chi.NewRouter()
	r.Post("/", createShortLink)
	r.Get("/{id}", useShortLink)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
