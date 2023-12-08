package models

// RequestAddShortURL описывает запрос пользователя на получение короткой ссылки.
type RequestAddShortURL struct {
	LongURL string `json:"url"`
}

// ResponseAddShortURL описывает запрос пользователя на получение короткой ссылки.
type ResponseAddShortURL struct {
	ShortURL string `json:"result"`
}
