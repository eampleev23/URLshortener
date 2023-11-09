package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostShortLink(t *testing.T) {
	// описываем ожидаемое тело ответа при успешном запросе
	successBody := "/EwHXdJfB"

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
			postShortLink(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")

			// Проверим корректность полученного ответа если мы его ожидаем
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}

func TestGetLongLink(t *testing.T) {

	testCases := []struct {
		method           string
		expectedCode     int
		expectedUrl      string
		expectedLocation string
	}{
		{method: http.MethodGet, expectedCode: 307, expectedUrl: "/EwHXdJfB", expectedLocation: "https://practicum.yandex.ru/"},
		{method: http.MethodPost, expectedCode: 400, expectedUrl: "/EwHXdJfB", expectedLocation: ""},
		{method: http.MethodPut, expectedCode: 400, expectedUrl: "/EwHXdJfB", expectedLocation: ""},
		{method: http.MethodDelete, expectedCode: 400, expectedUrl: "/EwHXdJfB", expectedLocation: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/EwHXdJfB", nil)
			w := httptest.NewRecorder()

			// вызываем хэндлер как обычную функцию без запуска сервера
			getLongLink(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedUrl, r.URL.String(), "Урл не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedLocation, w.Header().Get("Location"), "Заголовок Location не совпадает с ожидаемым")

		})
	}
}
