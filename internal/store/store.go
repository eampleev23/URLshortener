package store

import (
	"bufio"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/generatelinks"
	"github.com/eampleev23/URLshortener/internal/logger"
	"go.uber.org/zap"
	"log"
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
	l  *logger.ZapLog
}

func NewStore(c *config.Config, l *logger.ZapLog) *Store {
	var perm os.FileMode = 0600
	file, err := os.OpenFile(c.GetValueByIndex("sfilepath"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		log.Printf("Error open file: %s", err)
	}
	return &Store{
		s:  make(map[string]LinksCouple),
		fp: &Producer{file: file, writer: bufio.NewWriter(file)},
		l:  l,
	}
}

func (s *Store) SetShortURL(longURL string) (string, error) {
	strResult, err := generatelinks.GenerateShortURL()
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	if _, ok := s.s[strResult]; !ok {
		linksCouple := LinksCouple{UUID: "1", ShortURL: strResult, OriginalURL: longURL}
		s.s[strResult] = linksCouple
		err := s.fp.WriteLinksCouple(&linksCouple)
		//if err != nil {
		//	return "", err
		//}
		if err != nil {
			s.l.ZL.Info("Ошибка при записи новой пары ссылок", zap.Error(err))
			//return "", err
		}

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

func (s *Store) ReadStoreFromFile(c *config.Config) {
	// открываем файл чтобы посчитать количество строк
	file, err := os.OpenFile(c.GetValueByIndex("sfilepath"), os.O_RDONLY|os.O_CREATE, 0600)

	if err != nil {
		log.Printf("%s", err)
	}

	if err != nil {
		log.Printf("Error open file: %s", err)
	}

	countLines, err := LineCounter(file)
	if err != nil {
		log.Printf("%s", err)
	}

	if countLines > 0 {
		// добавляем каждую существующую строку в стор
		fc, err := NewConsumer(c.GetValueByIndex("sfilepath"))
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
	}
}
