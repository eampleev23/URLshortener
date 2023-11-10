package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
)

var linksCouples = map[string]string{
	"https://practicum.yandex.ru/": "EwHXdJfB",
	"ampleev.com":                  "AwHXzJsN",
}

func createShortLink(res http.ResponseWriter, req *http.Request) {

	fmt.Println("вызвался createShortLink")

	if req.Method == http.MethodPost {

		fmt.Println("подтвердился метод POST")
		fmt.Println("тип req.Body", reflect.TypeOf(req.Body).Kind())
		fmt.Println("req.Body", req.Body)
		// теперь необходимо создать новую пару длинная короткая ссылка
		// body переформатируем в строку
		var body string
		if b, err := io.ReadAll(req.Body); err == nil {
			body = string(b)
		}
		fmt.Println("тип body после io.ReadAll", reflect.TypeOf(body).Kind())
		fmt.Println("Содержание: ", body)

		body = strings.Trim(body, "url=")

		linksCouples[body] = generateUniqLink()
		fmt.Println("linksCouples:", linksCouples)
		fmt.Println("linksCouples[body]: ", linksCouples[body])

		// установить статус 201
		res.WriteHeader(201)

		// установить тип контента text/plain
		res.Header().Set("content-type", "text/plain")

		// добавить в качестве ответа сокращенный урл
		res.Write([]byte("/"))

	} else {

		// установить статус StatusBadRequest
		res.WriteHeader(http.StatusBadRequest)
	}

}

func useShortLink(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Зашли в хэндлер useShortLink")
	if req.Method == http.MethodGet {
		fmt.Println("Метод GET определился корректно")
		fmt.Println("req.URL: ", req.URL)
		// Дальше проверяем есть ли такой урл в базе, но вычитаем слеш в начале
		//fmt.Println("тип req.Url: ", reflect.TypeOf(req.URL.Query().Get("url")).Kind())
		fmt.Println("chi.URLParam(req, \"id\"): ", chi.URLParam(req, "id"))

		var url string
		url = chi.URLParam(req, "id")
		result := searchUrl(url)
		fmt.Println("result: ", result)
		if result == "" {
			res.WriteHeader(http.StatusBadRequest)
		} else {
			res.WriteHeader(307)
			res.Header().Set("Location", "https://practicum.yandex.ru/")
		}

	} else {

		// установить статус StatusBadRequest
		res.WriteHeader(http.StatusBadRequest)
	}

}

// Вспомогательная функция для поиска совпадений по урлам в базе
func searchUrl(url string) string {
	if c, ok := linksCouples[url]; ok {
		return c
	}
	return ""
}

// Вспомогательная функция для генерации уникальной короткой ссылки
func generateUniqLink() string {
	//rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	all := "abcdefghijklmnopqrstuvwxyz" + digits
	length := 8
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	str := string(buf) // Например "3i[g0|)z"
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
