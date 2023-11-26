package store

import "math/rand"

type Store struct {
	s map[string]string
}

func NewStore() *Store {
	return &Store{
		s: make(map[string]string),
	}
}

func (s *Store) SetShortURL(longURL string) (string, error) {
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
	if _, ok := s.s[strResult]; !ok {
		s.s[strResult] = longURL
		return strResult, nil
	}
	return "", nil

}

func (s *Store) GetLongLinkByShort(shortURL string) (string, error) {
	if c, ok := s.s[shortURL]; ok {
		return c, nil
	}
	return "no match", nil
}

func (s *Store) SetDataForMyTests() error {
	s.s["shortlink"] = "longlink"
	return nil
}
