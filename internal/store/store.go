package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/eampleev23/URLshortener/internal/datagen"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
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
	// Close закрывает соединение с базой данных
	Close() (err error)
	// GetURLsByOwnerID возвращает ссылки по ID пользователя с использованием авторизации.
	GetURLsByOwnerID(ctx context.Context, db *sql.DB, ownerID int64) ([]LinksCouple, error)
}

// ErrConflict ошибка, которую используем для сигнала о нарушении целостности данных.
var ErrConflict = errors.New("data conflict")

func NewStorage(c *config.Config, l *logger.ZapLog) (Store, error) {
	switch {
	case len(c.DBDSN) != 0:
		// используем в качестве хранилища только базу данных
		s, err := NewDBStore(c, l)
		if err != nil {
			return nil, fmt.Errorf("error creating new db store: %w", err)
		}
		err = s.createTable()
		if err != nil {
			return nil, fmt.Errorf("error create table: %w", err)
		}
		err = datagen.GenerateData(context.Background(), c, l)
		if err != nil {
			return nil, fmt.Errorf("error data generation: %w", err)
		}
		return s, nil

	case len(c.SFilePath) != 0:
		l.ZL.Info("Using File Store..")
		s, err := NewFileStore(c, l)
		if err != nil {
			return nil, fmt.Errorf("error creating new file store: %w", err)
		}
		return s, nil
	default:
		l.ZL.Info("Using Memory Store..")
		s, err := NewMemoryStore(c, l)
		if err != nil {
			return nil, fmt.Errorf("error create memory store: %w", err)
		}
		return s, nil
	}
}

type LinksCouple struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	OwnerID     int    `json:"owner_id"`
}
