package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/generatelinks"
	"github.com/eampleev23/URLshortener/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

var ErrConflict = errors.New("data conflict")

type DBStore struct {
	dbConn *sql.DB
}

func NewDBStore(c *config.Config, l *logger.ZapLog) (*DBStore, error) {
	db, err := sql.Open("pgx", c.DBDSN)
	if err != nil {
		return &DBStore{}, fmt.Errorf("%w", errors.New("sql.open failed in case to create store"))
	}
	// Отложенно закрываем соединение.
	defer func() {
		if err := db.Close(); err != nil {
			l.ZL.Info("failed to properly close the DB connection")
		}
	}()
	return &DBStore{dbConn: db}, nil
}

func (ds DBStore) SetShortURL(ctx context.Context, originalURL string) (shortURL string, err error) {
	shortURL = ""
	// Сюда приходит короткая ссылка без проверки на коллизии
	newShortLink := generatelinks.GenerateShortURL()
	// Создаем структуру и в нее записываем значение
	linksCouple := LinksCouple{ShortURL: newShortLink, OriginalURL: originalURL}
	_, err = ds.dbConn.ExecContext(ctx,
		`INSERT INTO links_couples(uuid, short_url, original_url) VALUES (DEFAULT, $1, $2)`,
		linksCouple.ShortURL,
		linksCouple.OriginalURL)
	if err != nil {
		return "", fmt.Errorf("insert error %w", err)
	}
	return shortURL, nil
}
func (ds DBStore) GetOriginalURLByShort(ctx context.Context, shortURL string) (originalURL string, err error) {
	row := ds.dbConn.QueryRowContext(ctx,
		`SELECT original_url FROM links_couples WHERE short_url = $1 LIMIT 1`, shortURL,
	)
	err = row.Scan(&originalURL) // Разбираем результат
	if err != nil {
		return "", fmt.Errorf("faild to get originalURL %w", err)
	}
	return originalURL, nil
}
func (ds DBStore) GetShortURLByOriginal(ctx context.Context, originalURL string) (shortURL string, err error) {
	row := ds.dbConn.QueryRowContext(ctx,
		`SELECT short_url FROM links_couples WHERE original_url = $1 LIMIT 1`, originalURL,
	)
	err = row.Scan(&shortURL) // Разбираем результат
	if err != nil {
		return "", fmt.Errorf("faild to get shortURL %w", err)
	}
	return shortURL, nil
}
func (ds DBStore) PingDB(ctx context.Context, timeLimit time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeLimit)
	defer cancel()
	err := ds.dbConn.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("db doesn't ping %w", err)
	}
	return nil
}
