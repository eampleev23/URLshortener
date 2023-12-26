package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"time"
)

type Store interface {
	// SetShortURL добавляет новое значение в стор.
	SetShortURL(ctx context.Context, originalURL string) (shortURL string, err error)
	// GetOriginalURLByShort возвращает оригинальную ссылку по короткой
	GetOriginalURLByShort(ctx context.Context, shortURL string) (originalURL string, err error)
	// GetShortURLByOriginal возвращает короткую ссылку по длинной если такая есть
	GetShortURLByOriginal(ctx context.Context, originalURL string) (shortURL string, err error)
	// PingDB проверяет подключение к базе данных
	PingDB(ctx context.Context, tiimeLimit time.Duration) (err error)
}
type Storage struct {
	s *Store
	c *config.Config
	l *logger.ZapLog
}

func NewStorage(c *config.Config, l *logger.ZapLog) (*Storage, error) {
	if len(c.DBDSN) != 0 {
		// используем в качестве хранилища только базу данных
		var s Store
		s, err := NewDBStore(c, l)
		if err != nil {
			return nil, fmt.Errorf("error creating new db store: %w", err)
		}
		return &Storage{
			s: &s,
			c: c,
			l: l,
		}, nil
	}
	return nil, errors.New("in the development for now: ")
}

type LinksCouple struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (s *Storage) SetShortURL(ctx context.Context, originalURL string) (shortURL string, err error) {
	shortURL, err = s.SetShortURL(ctx, originalURL)
	if err != nil {
		return "", fmt.Errorf("error set short URL: %w", err)
	}
	return shortURL, nil
}
func (s *Storage) GetOriginalURLByShort(ctx context.Context, shortURL string) (originalURL string, err error) {
	originalURL, err = s.GetOriginalURLByShort(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("error get original URL by short: %w", err)
	}
	return originalURL, nil
}
func (s *Storage) GetShortURLByOriginal(ctx context.Context, originalURL string) (shortURL string, err error) {
	shortURL, err = s.GetShortURLByOriginal(ctx, originalURL)
	if err != nil {
		return "", fmt.Errorf("error get short URL by original: %w", err)
	}
	return shortURL, nil
}
func (s *Storage) PingDB(ctx context.Context, timeLimit time.Duration) error {
	err := s.PingDB(ctx, s.c.TLimitQuery)
	if err != nil {
		return fmt.Errorf("db doesn't ping now %w", err)
	}
	return nil
}
