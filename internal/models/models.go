package models

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
	LongURL string `json:"url"`
}

// ResponseAddShortURL описывает запрос пользователя на получение короткой ссылки.
type ResponseAddShortURL struct {
	ShortURL string `json:"result"`
}
