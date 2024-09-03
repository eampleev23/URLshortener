package handlers

import (
	"fmt"
	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/services"
	"github.com/eampleev23/URLshortener/internal/store"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// ExampleHandlers_CreateShortURL - пример использования хэндлера для создания короткой ссылки
// и одновременно тест проверки допустимого времени генерации
func ExampleHandlers_CreateShortURL() {
	l, err := logger.NewZapLogger("info")
	if err != nil {
		fmt.Println("Failed..")
		return
	}
	c, err := config.NewConfig()
	if err != nil {
		fmt.Println("Failed..")
		return
	}
	au, err := myauth.Initialize(c.SecretKey, c.TokenEXP, l)
	if err != nil {
		fmt.Println("Failed..")
		return
	}
	s, err := store.NewMemoryStore(c, l)
	if err != nil {
		fmt.Println("Failed..")
		return
	}
	serv := services.NewServices(s, c, l, *au)
	h := NewHandlers(s, c, l, *au, serv)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://test.com"))
	w := httptest.NewRecorder()

	start := time.Now()
	h.CreateShortURL(w, r)
	duration := time.Since(start)
	if duration > time.Minute {
		fmt.Printf("Duration GenerateShortURL > time.Minute, the value is %s (OMG!!!)\n", duration)
	} else {
		fmt.Println("Ok")
	}

	// Output:
	// Ok
}
