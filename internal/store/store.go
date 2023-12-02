package store

import (
	"bufio"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/config"
	"log"
	"math/rand"
	"os"
)

type LinksCouple struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Store struct {
	s  map[string]LinksCouple
	fp *Producer
	c  *config.Config
}

func NewStore(c *config.Config) *Store {
	file, err := os.OpenFile("../../tmp/"+c.GetValueByIndex("sfilepath"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Error open file: %s", err)
	}
	return &Store{
		s:  make(map[string]LinksCouple),
		fp: &Producer{file: file, writer: bufio.NewWriter(file)},
	}
}

func (s *Store) SetShortURL(longURL string) (string, error) {

	strResult, err := generateShortURL()
	if err != nil {
		return "", err
	}

	if _, ok := s.s[strResult]; !ok {
		linksCouple := LinksCouple{UUID: "1", ShortURL: strResult, OriginalURL: longURL}
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

// Вспомогательная функция для генерации коротких ссылок
func generateShortURL() (string, error) {
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
func (s *Store) ReadStoreFromFile(c *config.Config) {
	// открываем файл чтобы посчитать количество строк
	file, err := os.OpenFile("../../tmp/"+c.GetValueByIndex("sfilepath"), os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		log.Printf("Error open file: %s", err)
	}

	countLines, err := LineCounter(file)
	if err != nil {
		fmt.Println("!!err:", err)
	}

	file.Close()
	if countLines > 0 {
		// добавляем каждую существующую строку в стор
		fc, err := NewConsumer("../../tmp/" + c.GetValueByIndex("sfilepath"))
		if err != nil {
			log.Printf("%s", err)
		}
		for i := 0; i < countLines; i++ {
			linksCouple, err := fc.ReadLinksCouple()
			if err != nil {
				log.Printf("%s", err)
			}
			fmt.Println("linksCouple=", linksCouple)
			s.s[linksCouple.ShortURL] = *linksCouple
		}
		fc.Close()
	}
}
