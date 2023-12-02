package store

import (
	"bufio"
	"math/rand"
	"os"
)

type LinksCouple struct {
	Uuid        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Store struct {
	s  map[string]LinksCouple
	fp *Producer
}

func NewStore() *Store {
	file, _ := os.OpenFile("./tmp/short-url-db.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	return &Store{
		s:  make(map[string]LinksCouple),
		fp: &Producer{file: file, writer: bufio.NewWriter(file)},
	}
}

func (s *Store) SetShortURL(longURL string) (string, error) {
	strResult, err := generateShortUrl()

	if err != nil {
		return "", err
	}

	if _, ok := s.s[strResult]; !ok {
		linksCouple := LinksCouple{Uuid: "1", ShortURL: strResult, OriginalURL: longURL}
		s.s[strResult] = linksCouple
		s.fp.WriteLinksCouple(&linksCouple)
		return strResult, nil
	}

	return "", nil

}

func (s *Store) GetLongLinkByShort(shortURL string) (string, error) {
	if c, ok := s.s[shortURL]; ok {
		return c.OriginalURL, nil
	}
	return "no match", nil
}

func generateShortUrl() (string, error) {
	// заводим слайс рун возможных для сгенерированной короткой ссылки
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	lenLetterRunes := len(letterRunes)
	// делаем из 2 символов
	b := make([]rune, 2)

	// генерируем случайный символ последовательно для всей длины
	for i := range b {
		b[i] = letterRunes[rand.Intn(lenLetterRunes)]
	}
	// в результат записываем байты преобразованные в строку
	strResult := string(b)
	return strResult, nil
}
