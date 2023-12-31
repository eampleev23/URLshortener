package store

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/generatelinks"
	"github.com/eampleev23/URLshortener/internal/logger"
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
	c  *config.Config
}

func NewStore(c *config.Config, l *logger.ZapLog) (*Store, error) {
	if c.SFilePath != "" {
		var perm os.FileMode = 0600
		file, err := os.OpenFile(c.SFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize a store by file: %w", err)
		}
		return &Store{
			s:  make(map[string]LinksCouple),
			fp: &Producer{file: file, writer: bufio.NewWriter(file)},
			l:  l,
			c:  c,
		}, nil
	}

	return &Store{
		s:  make(map[string]LinksCouple),
		fp: nil,
		l:  l,
		c:  c,
	}, nil
}

func (s *Store) SetShortURL(longURL string) (string, error) {
	// Сюда приходит короткая ссылка без проверки на коллизии
	newShortLink := generatelinks.GenerateShortURL()

	// Если такой короткой ссылки еще нет в базе, значит можем спокойно записывать
	if _, ok := s.s[newShortLink]; !ok {
		// Создаем структуру по заданию и в нее записываем значение
		linksCouple := LinksCouple{UUID: "1", ShortURL: newShortLink, OriginalURL: longURL}
		// Заносим эту структуру в стор
		s.s[newShortLink] = linksCouple
		if s.c.SFilePath != "" {
			// Также записываем в файл
			err := s.fp.WriteLinksCouple(&linksCouple)
			if err != nil {
				delete(s.s, newShortLink)
				return "", fmt.Errorf("failed to write a new couple links in file %w", err)
			}
		}
		return newShortLink, nil
	}
	// Иначе у нас произошла коллизия
	return "", errors.New("a collision occurred")
}

func (s *Store) GetLongLinkByShort(shortURL string) (string, error) {
	if c, ok := s.s[shortURL]; ok {
		return c.OriginalURL, nil
	}
	return "no match", nil
}

func (s *Store) ReadStoreFromFile(c *config.Config) {
	var perm os.FileMode = 0600
	// открываем файл чтобы посчитать количество строк
	file, err := os.OpenFile(c.SFilePath, os.O_RDONLY|os.O_CREATE, perm)

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
		fc, err := NewConsumer(c.SFilePath)
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
