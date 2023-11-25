package store

import (
	"math/rand"
	"strings"
)

type Store struct {
	s map[string]string
}

func NewStore() *Store {
	return &Store{
		s: make(map[string]string),
	}
}

func (s *Store) SetShortURL(longURL string) (string, error) {
	chars := []rune(
		"abcdefghijklmnopqrstuvwxy" +
			"0123456789")
	length := 8
	var b strings.Builder
	//b := make([]byte, length)
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String() // Например "ExcbsVQs"
	s.s[str] = longURL
	return str, nil
}
func (s *Store) GetLongLinkByShort(shortURL string) (string, error) {
	return s.s[shortURL], nil
}
