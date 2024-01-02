package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/generatelinks"
	"github.com/eampleev23/URLshortener/internal/logger"
	"go.uber.org/zap"
)

type MemoryStore struct {
	s map[string]LinksCouple
	c *config.Config
	l *logger.ZapLog
}

func NewMemoryStore(c *config.Config, l *logger.ZapLog) (*MemoryStore, error) {
	return &MemoryStore{
		s: make(map[string]LinksCouple),
		c: c,
		l: l,
	}, nil
}

func (ms *MemoryStore) SetShortURL(ctx context.Context, originalURL string) (newShortURL string, err error) {
	// Проверяем есть ли уже такой оригинальный урл в базе
	for i, v := range ms.s {
		if v.OriginalURL == originalURL {
			err = ErrConflict
			ms.l.ZL.Info("original url already exists", zap.String("originalURL", originalURL))
			return i, fmt.Errorf("original url %v already exists: %w", originalURL, err)
		}
	}
	newShortURL = generatelinks.GenerateShortURL()
	// Если такой короткой ссылки еще нет в базе, значит можем спокойно записывать
	if _, ok := ms.s[newShortURL]; !ok {
		linksCouple := LinksCouple{UUID: "1", ShortURL: newShortURL, OriginalURL: originalURL}
		ms.s[newShortURL] = linksCouple
		return newShortURL, nil
	}
	// Произошла коллизия
	err = ErrConflict
	ms.l.ZL.Info("There was a collision", zap.String("newShortURL", newShortURL))
	return "", fmt.Errorf("shortURL %v already exists: %w", newShortURL, err)
}
func (ms *MemoryStore) GetOriginalURLByShort(ctx context.Context, shortURL string) (originalURL string, err error) {
	if _, ok := ms.s[shortURL]; ok {
		return ms.s[shortURL].OriginalURL, nil
	}
	return "", fmt.Errorf("no shortURL like this %v: %w", shortURL, err)
}
func (ms *MemoryStore) GetShortURLByOriginal(ctx context.Context, originalURL string) (shortURL string, err error) {
	is := false
	shortURL = ""
	for i, v := range ms.s {
		if v.OriginalURL == originalURL {
			is = true
			shortURL = i
		}
	}
	if is {
		return shortURL, nil
	}
	return shortURL, fmt.Errorf("there is no original URL like this: %v", originalURL)
}

// PingDB проверяет подключение к базе данных.
func (ms *MemoryStore) PingDB(ctx context.Context, tiimeLimit time.Duration) (err error) {
	return nil
}

// Close закрывает соединение с базой данных.
func (ms *MemoryStore) Close() (err error) {
	return nil
}

func (ms *MemoryStore) GetURLsByOwnerID(ctx context.Context, ownerID int64) ([]LinksCouple, error) {
	return nil, errors.New("map store doesn't use this method")
}
