package main

import (
	"math/rand"
	"strings"
)

// Вспомогательная функция для генерации уникальной короткой ссылки
func generateUniqShortLink() (string, error) {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXY" +
		"abcdefghijklmnopqrstuvwxy" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String() // Например "ExcbsVQs"

	if _, ok := linksCouples[str]; ok {
		return generateUniqShortLink()
	}
	return str, nil
}

// Вспомогательная функция для поиска совпадений по урлам в базе
func searchShortLink(shortLink string) string {

	if c, ok := linksCouples[shortLink]; ok {
		return c
	}
	return "no match"
}
