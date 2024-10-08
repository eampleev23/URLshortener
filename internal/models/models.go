// Package models - структуры данных, используемые в приложении.
package models

import (
	"github.com/eampleev23/URLshortener/internal/store"
)

// BatchItemReq описывает один элемент в запросе пользователя на получение партии коротких ссылок.
type BatchItemReq struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchItemRes описывает один элемент в ответе пользователю на получение партии коротких ссылок.
type BatchItemRes struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// RequestAddShortURL описывает запрос пользователя на получение короткой ссылки.
type RequestAddShortURL struct {
	OriginalURL string `json:"url"`
}

// ResponseAddShortURL описывает запрос пользователя на получение короткой ссылки.
type ResponseAddShortURL struct {
	ShortURL string `json:"result"`
}

// ResponseGetOwnerURL описывает элемент ответа пользователю на получение всех его ссылок.
type ResponseGetOwnerURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// ResponseGlobalStats описывает элемент ответа пользователю на получение всех его ссылок.
type ResponseGlobalStats struct {
	URLs  int64 `json:"urls"`
	Users int64 `json:"users"`
}

// GetResponseGetOwnerURLs - конвертирует в необходимый формат урля пользователя.
func GetResponseGetOwnerURLs(source []store.LinksCouple) (result []ResponseGetOwnerURL, err error) {
	result = make([]ResponseGetOwnerURL, 0, len(source))
	for _, v := range source {
		result = append(result, ResponseGetOwnerURL{
			ShortURL:    v.ShortURL,
			OriginalURL: v.OriginalURL,
		})
	}

	return result, nil
}
