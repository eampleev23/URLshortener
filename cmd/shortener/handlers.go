package main

import (
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
)

func createShortLink(res http.ResponseWriter, req *http.Request) {
	// Сначала необходимо убедиться, что запрос корректный (в теле должна быть строка как text/plain в виде урл
	// а вернуть должен код 201 и сокращенный урл как text/plain

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusBadRequest)
	} else {

		// заводим строковую переменную чтобы хранить в ней значение длинной ссылки
		var longLink string
		if b, err := io.ReadAll(req.Body); err == nil {
			longLink = string(b)
		}
		// Генерируем короткую ссылку для переданной длинной
		shortLink, err := generateUniqShortLink()
		if err != nil {
			log.Fatal(err)
		}

		// Записываем в бд // В качестве индекса мапы используем короткую ссылку чтобы можно было быстро найти
		linksCouples[shortLink] = longLink

		// Устанавливаем статус 201
		res.WriteHeader(201)

		// Устаннавливаем тип контента text/plain
		res.Header().Set("content-type", "text/plain")

		// В качестве ответа возвращаем сокращенный урл с именем домена
		//shortLinkWithPrefix := "http://localhost:8080/" + shortLink
		shortLinkWithPrefix := baseShortURL + "/" + shortLink
		res.Write([]byte(shortLinkWithPrefix))

	}

}

func useShortLink(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusBadRequest)
	} else {
		res.Header().Add("Location", linksCouples[chi.URLParam(req, "id")])
		// добавляю для эксперимента
		res.WriteHeader(http.StatusTemporaryRedirect)
		// Объявляем строковую переменную, в которой будем хранить урл (ожидаем, что это короткий урл из базы)
		shortLink := chi.URLParam(req, "id")

		// В строковой переменной резалт храним либо longLink если есть совпадение в базе, либо "no match" если в базе такой ссылки нет
		result := searchShortLink(shortLink)

		// Если совпадений в бд нет, то ставим статус код бэд реквест
		if result == "no match" {
			res.WriteHeader(http.StatusBadRequest)
		} else { // иначе это успех, есть совпадение, ставим 307 и в заголовок ответа локейшн отправляем длинную ссылку
			res.WriteHeader(http.StatusTemporaryRedirect)
		}

	}

}
