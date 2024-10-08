// Package generatelinks - пакет, занимающийся генерацией коротких ссылок.
package generatelinks

import (
	"math/rand"
)

// GenerateShortURL Вспомогательная функция для генерации коротких ссылок.
func GenerateShortURL() string {
	// заводим слайс рун возможных для сгенерированной короткой ссылки
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	lenLetterRunes := len(letterRunes)
	// задаем количество символов в сгенерированной ссылке
	numberOfSimbols := 5
	b := make([]rune, numberOfSimbols)

	// генерируем случайный символ последовательно для всей длины
	for i := range b {
		b[i] = letterRunes[rand.Intn(lenLetterRunes)]
	}
	// в результат записываем байты преобразованные в строку
	strResult := string(b)
	return strResult
}
