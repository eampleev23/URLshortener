package models

// Request описывает запрос пользователя на получение короткой ссылки
type RequestAddShortURL struct {
	LongURL string `json:"url"`
}

// Request описывает запрос пользователя на получение короткой ссылки
type ResponseAddShortURL struct {
	ShortURL string `json:"result"`
}
