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

var linksCouples = map[string]string{}

func createShortLink(res http.ResponseWriter, req *http.Request) {
	log.Printf("запустили createShortLink")
	// Сначала необходимо убедиться, что запрос корректный (в теле должна быть строка как text/plain в виде урл
	// а вернуть должен код 201 и сокращенный урл как text/plain

	// заводим строкову переменную чтобы хранить в ней значение длинной ссылки
	var longLink string
	if b, err := io.ReadAll(req.Body); err == nil {
		longLink = string(b)
	}
	log.Printf("longLink = %s", longLink)

	// Генерируем короткую ссылку для переданной длинной
	shortLink := generateUniqShortLink()
	log.Printf("shortLink = %s", shortLink)

	// Записываем в бд // В качестве индекса мапы используем короткую ссылку чтобы можно было быстро найти
	linksCouples[shortLink] = longLink
	log.Printf("linksCouples = %s", linksCouples)

	// Устанавливаем статус 201
	res.WriteHeader(201)

	// Устаннавливаем тип контента text/plain
	res.Header().Set("content-type", "text/plain")

	// В качестве ответа возвращаем сокращенный урл с именем домена
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
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXY" +
		"abcdefghijklmnopqrstuvwxy" +
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
