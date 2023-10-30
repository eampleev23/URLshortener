package main

import "net/http"

func postShortLink(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

		// установить статус 201
		res.WriteHeader(201)

		// установить тип контента text/plain
		res.Header().Set("content-type", "text/plain")

		// добавить в качестве ответа сокращенный урл
		res.Write([]byte("http://localhost:8080/EwHXdJfB"))

	} else {
		// установить статус 201
		res.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func main() {
	http.HandleFunc(`/`, postShortLink)

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
