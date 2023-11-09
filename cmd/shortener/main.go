package main

import (
	"net/http"
)

func postShortLink(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

		// установить статус 201
		res.WriteHeader(201)

		// установить тип контента text/plain
		res.Header().Set("content-type", "text/plain")

		// добавить в качестве ответа сокращенный урл
		res.Write([]byte("/EwHXdJfB"))

	} else {

		// установить статус StatusBadRequest
		res.WriteHeader(http.StatusBadRequest)
	}

}

func getLongLink(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {

		// Эндпоинт с методом GET и путём /{id}, где id — идентификатор сокращённого URL (например, /EwHXdJfB).
		//В случае успешной обработки запроса сервер возвращает ответ с кодом 307
		//и оригинальным URL в HTTP-заголовке Location.
		res.WriteHeader(307)
		res.Header().Set("Location", "https://practicum.yandex.ru/")
		//fmt.Println(req.URL)
		//fmt.Println(res.Header().Get("Location"))

	} else {

		// установить статус StatusBadRequest
		res.WriteHeader(http.StatusBadRequest)
	}

}

func main() {
	http.HandleFunc(`/`, postShortLink)
	http.HandleFunc(`/EwHXdJfB`, getLongLink)

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
