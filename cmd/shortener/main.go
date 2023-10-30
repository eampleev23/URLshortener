package main

import "net/http"

func mainPage(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Привет!"))
}

func main() {
	http.HandleFunc(`/`, mainPage)

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
