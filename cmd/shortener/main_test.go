package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateShortLink(t *testing.T) {
	// описываем ожидаемое тело ответа при успешном запросе

	testCases := []struct {
		method       string
		expectedCode int
		contentType  string
	}{
		{method: http.MethodGet, expectedCode: http.StatusBadRequest, contentType: ""},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest, contentType: ""},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest, contentType: ""},
		{method: http.MethodPost, expectedCode: 201, contentType: "text/plain"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", nil)
			w := httptest.NewRecorder()

			// Вызовем хэндлер как обычную функцию без запуска сервера
			createShortLink(w, r)
			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")

			// Проверим корректность полученного ответа если мы его ожидаем
			// // Пока закомментировал т.к. функция генерит уникальный и это по идее норм
			//if tc.expectedBody != "" {
			//	assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			//}
		})
	}
}

func TestUseShortLink(t *testing.T) {

	testCases := []struct {
		method       string
		expectedCode int
		expectedURL  string
	}{
		{method: http.MethodGet, expectedCode: 307, expectedURL: "http://localhost:8080/shortlink"},
		{method: http.MethodPost, expectedCode: 400, expectedURL: "http://localhost:8080/shortlink"},
		{method: http.MethodPut, expectedCode: 400, expectedURL: "http://localhost:8080/shortlink"},
		{method: http.MethodDelete, expectedCode: 400, expectedURL: "http://localhost:8080/shortlink"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://localhost:8080/shortlink", nil)
			w := httptest.NewRecorder()

			// вызываем хэндлер как обычную функцию без запуска сервера
			useShortLink(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedURL, r.URL.String(), "Урл не совпадает с ожидаемым")
			//assert.Equal(t, tc.expectedLocation, w.Header().Get("Location"), "Заголовок Location не совпадает с ожидаемым")

		})
	}
}
