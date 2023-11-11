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
	"shortlink": "longlink",
}

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
	log.Printf("Запустили useShortLink")

	// Объявляем строковую переменную, в которой будем хранить урл (ожидаем, что это короткий урл из базы)
	var shortLink string
	shortLink = chi.URLParam(req, "id")
	log.Printf("shortLink = %s", shortLink)

	// В строковой переменной резалт храним либо longLink если есть совпадение в базе, либо "no match" если в базе такой ссылки нет
	result := searchShortLink(shortLink)
	log.Printf("result =`%s`, shortlik = `%s`, linksCouples = `%s`", result, shortLink, linksCouples)

	// Если совпадений в бд нет, то ставим статус код бэд реквест
	if result == "no match" {
		res.WriteHeader(http.StatusBadRequest)
	} else { // иначе это успех, есть совпадение, ставим 307 и в заголовок ответа локейшн отправляем длинную ссылку
		res.WriteHeader(307)
		res.Header().Add("location", linksCouples[shortLink])
		log.Printf("Header location `%s`", res.Header().Get("location"))
	}
	res.Write([]byte(""))
}

// Вспомогательная функция для поиска совпадений по урлам в базе
func searchShortLink(shortLink string) string {
	log.Printf("start searchShortLink..")
	log.Printf("shortLink get is %s", shortLink)
	log.Printf("linksCouples[%s] = %s", shortLink, linksCouples[shortLink])
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
