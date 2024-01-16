package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	myauth "github.com/eampleev23/URLshortener/internal/auth"

	"github.com/go-resty/resty/v2"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/handlers"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
	"github.com/stretchr/testify/assert"
)

var c *config.Config
var l *logger.ZapLog

func TestCreateShortLink(t *testing.T) {
	l, err := logger.NewZapLogger("info")
	if err != nil {
		t.Log(err)
	}
	c, _ = config.NewConfig(l)
	au, err := myauth.Initialize(c.SecretKey, c.TokenEXP, l)
	if err != nil {
		t.Log(err)
	}
	s, err := store.NewMemoryStore(c, l)
	if err != nil {
		t.Log(err)
	}
	h := handlers.NewHandlers(s, c, l, *au)

	testCases := []struct {
		method       string
		expectedCode int
		contentType  string
	}{
		{method: http.MethodGet, expectedCode: http.StatusInternalServerError, contentType: ""},
		{method: http.MethodPut, expectedCode: http.StatusInternalServerError, contentType: ""},
		{method: http.MethodDelete, expectedCode: http.StatusInternalServerError, contentType: ""},
		{method: http.MethodPost, expectedCode: http.StatusInternalServerError, contentType: "text/plain"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", nil)
			w := httptest.NewRecorder()

			// Вызовем хэндлер как обычную функцию без запуска сервера
			h.CreateShortURL(w, r)
			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestUseShortLink(t *testing.T) {
	s, err := store.NewMemoryStore(c, l)
	if err != nil {
		t.Log(err)
	}
	l, err = logger.NewZapLogger("info")
	if err != nil {
		t.Log(err)
	}
	au, err := myauth.Initialize(c.SecretKey, c.TokenEXP, l)
	if err != nil {
		t.Log(err)
	}
	h := handlers.NewHandlers(s, c, l, *au)

	testCases := []struct {
		method       string
		expectedCode int
		expectedURL  string
	}{
		{method: http.MethodGet, expectedCode: 400, expectedURL: "http://localhost:8080/shortlink"},
		{method: http.MethodPost, expectedCode: 400, expectedURL: "http://localhost:8080/shortlink"},
		{method: http.MethodPut, expectedCode: 400, expectedURL: "http://localhost:8080/shortlink"},
		{method: http.MethodDelete, expectedCode: 400, expectedURL: "http://localhost:8080/shortlink"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://localhost:8080/shortlink", nil)
			w := httptest.NewRecorder()

			// вызываем хэндлер как обычную функцию без запуска сервера
			h.UseShortLink(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedURL, r.URL.String(), "Урл не совпадает с ожидаемым")
			// assert.Equal(t, tc.expectedLocation, w.Header().Get("Location"), "Заголовок Location не совпадает с ожидаемым").
		})
	}
}

func TestJSONHandler(t *testing.T) {
	s, err := store.NewMemoryStore(c, l)
	if err != nil {
		t.Log(err)
	}
	l, err = logger.NewZapLogger("info")
	if err != nil {
		t.Log(err)
	}
	au, err := myauth.Initialize(c.SecretKey, c.TokenEXP, l)
	if err != nil {
		t.Log(err)
	}
	h := handlers.NewHandlers(s, c, l, *au)
	handler := http.HandlerFunc(h.JSONHandler)
	srv := httptest.NewServer(handler)
	defer srv.Close()
	// successBody := `{"result": "http://localhost:8080/EwHXdJfB"}`
	testCases := []struct {
		name         string // добавляем название тестов
		method       string
		body         string // добавляем тело запроса в табличные тесты
		expectedCode int
		expectedBody string
	}{
		{
			name:         "method_get",
			method:       http.MethodGet,
			expectedCode: 400,
			expectedBody: "",
		},
		{
			name:         "method_put",
			method:       http.MethodPut,
			expectedCode: 400,
			expectedBody: "",
		},
		{
			name:         "method_delete",
			method:       http.MethodDelete,
			expectedCode: 400,
			expectedBody: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL

			if len(tc.body) > 0 {
				req.SetHeader("Content-Type", "application/json")
				req.SetBody(tc.body)
			}

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			// проверяем корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(resp.Body()))
				// to delete it later
			}
		})
	}
}
