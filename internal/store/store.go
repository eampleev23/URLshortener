package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/eampleev23/URLshortener/internal/datagen"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
)

type Store interface {
	// SetShortURL добавляет новое значение в стор.
	SetShortURL(ctx context.Context, originalURL string, ownerID int) (shortURL string, err error)
	// GetOriginalURLByShort возвращает оригинальную ссылку по короткой
	GetOriginalURLByShort(ctx context.Context, shortURL string) (originalURL string, err error)
	// GetShortURLByOriginal возвращает короткую ссылку по длинной если такая есть
	GetShortURLByOriginal(ctx context.Context, originalURL string) (shortURL string, err error)
	// PingDB проверяет подключение к базе данных
	PingDB(ctx context.Context, tiimeLimit time.Duration) (err error)
	// Close закрывает соединение с базой данных
	Close() (err error)
	// GetURLsByOwnerID возвращает ссылки по ID пользователя с использованием авторизации.
	GetURLsByOwnerID(ctx context.Context, ownerID int) ([]LinksCouple, error)
	// DeleteURLS проставляет флаг удаления всем переданным shortURL у которых ownerID совпадает с id отправителя запроса
	DeleteURLS(ctx context.Context, deleteItems []DeleteURLItem) (err error)
	// GetLinksCoupleByShortURL возвращает LinksCouple со всеми полями по shortURL
	GetLinksCoupleByShortURL(ctx context.Context, shortURL string) (lc LinksCouple, err error)
}

type LinksCouple struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	OwnerID     int    `json:"owner_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

type DeleteURLItem struct {
	ShortURL   string
	DeleteFlag bool
	OwnerID    int
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
		err = datagen.GenerateData(context.Background(), c, l)
		if err != nil {
			return nil, fmt.Errorf("error data generation: %w", err)
		}
		l.ZL.Info("Use DB store..")
		return s, nil
	case len(c.SFilePath) != 0:
		s, err := NewFileStore(c, l)
		if err != nil {
			return nil, fmt.Errorf("error creating new file store: %w", err)
		}
		l.ZL.Info("Use File Store..")
		return s, nil
	default:
		s, err := NewMemoryStore(c, l)
		if err != nil {
			return nil, fmt.Errorf("error create memory store: %w", err)
		}
		l.ZL.Info("Use Memory Store..")
		return s, nil
	}
}
