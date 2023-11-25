package mymiddlewares

import (
	"log"
	"net/http"
)

func MyMiddlewareTest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("my middleware")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
