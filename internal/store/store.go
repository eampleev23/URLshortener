package store

import (
	"fmt"
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
	if _, ok := s.s[str]; !ok {
		s.s[str] = longURL
		return str, nil
	} else {
		return "", fmt.Errorf("The generated shortlink is not uniq, error: %w", ok) //надо разобраться
	}

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
