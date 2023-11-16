package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateShortLink(t *testing.T) {
	// описываем ожидаемое тело ответа при успешном запросе
	successBody := "http://localhost:8080/" + generateUniqShortLink()

	testCases := []struct {
		method       string
		expectedCode int
		contentType  string
		expectedBody string
	}{
		{method: http.MethodGet, expectedCode: http.StatusBadRequest, contentType: "", expectedBody: ""},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest, contentType: "", expectedBody: ""},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest, contentType: "", expectedBody: ""},
		{method: http.MethodPost, expectedCode: 201, contentType: "text/plain", expectedBody: successBody},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", nil)
			w := httptest.NewRecorder()

			// Вызовем хэндлер как обычную функцию без запуска сервера
			createShortLink(w, r)
			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")

			// Проверим корректность полученного ответа если мы его ожидаем
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}

func TestUseShortLink(t *testing.T) {

	testCases := []struct {
		method           string
		expectedCode     int
		expectedUrl      string
		expectedLocation string
	}{
		{method: http.MethodGet, expectedCode: 307, expectedUrl: "http://localhost:8080/shortlink", expectedLocation: "longlink"},
		{method: http.MethodPost, expectedCode: 400, expectedUrl: "http://localhost:8080/shortlink", expectedLocation: ""},
		{method: http.MethodPut, expectedCode: 400, expectedUrl: "http://localhost:8080/shortlink", expectedLocation: ""},
		{method: http.MethodDelete, expectedCode: 400, expectedUrl: "http://localhost:8080/shortlink", expectedLocation: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://localhost:8080/shortlink", nil)
			w := httptest.NewRecorder()

			// вызываем хэндлер как обычную функцию без запуска сервера
			useShortLink(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedUrl, r.URL.String(), "Урл не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedLocation, w.Header().Get("Location"), "Заголовок Location не совпадает с ожидаемым")

		})
	}
}
