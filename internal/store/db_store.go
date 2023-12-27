package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/generatelinks"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"time"
)

type DBStore struct {
	dbConn *sql.DB
	c      *config.Config
	l      *logger.ZapLog
}

func NewDBStore(c *config.Config, l *logger.ZapLog) (*DBStore, error) {
	db, err := sql.Open("pgx", c.DBDSN)
	if err != nil {
		return &DBStore{}, fmt.Errorf("%w", errors.New("sql.open failed in case to create store"))
	}
	return &DBStore{
		dbConn: db,
		c:      c,
		l:      l,
	}, nil
}

// SetShortURL вставляет в бд новую строку или возвращает специфическую ошибку в случае конфликта
func (ds DBStore) SetShortURL(ctx context.Context, originalURL string) (newShortURL string, err error) {
	newShortURL, err = ds.InsertURL(ctx, LinksCouple{ShortURL: generatelinks.GenerateShortURL(), OriginalURL: originalURL})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		err = ErrConflict
		return "", fmt.Errorf("conflict: %w", err)
	}
	if err != nil {
		return "", fmt.Errorf("error InsertURL: %w", err)
	}
	return newShortURL, nil
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
func (ds DBStore) Close() error {
	if err := ds.dbConn.Close(); err != nil {
		ds.l.ZL.Info("failed to properly close the DB connection", zap.Error(err))
		return err
	}
	return nil
}

// CreateTable создает таблицу если ее нет при инициализации стора. Без этого не работают тесты 11 инкремента
func (ds DBStore) createTable() error {
	ctx := context.Background()
	defer ctx.Done()
	_, err := ds.dbConn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS links_couples (
        "uuid" int GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
        "short_url" VARCHAR(250) NOT NULL DEFAULT '',
        "original_url"  VARCHAR(1000) NOT NULL DEFAULT ''
      );
      CREATE UNIQUE INDEX IF NOT EXISTS links_couples_index_by_original_url_unique
    ON links_couples
        USING btree (original_url);
      `)
	if err != nil {
		return fmt.Errorf("faild to create table links_couples %w", err)
	}

	return nil
}

// InsertURL занимается непосредственно запросом вставки строки в бд
func (ds DBStore) InsertURL(ctx context.Context, linksCouple LinksCouple) (shortURL string, err error) {
	_, err = ds.dbConn.ExecContext(ctx, `INSERT INTO links_couples(uuid, short_url, original_url)
VALUES (DEFAULT, $1, $2)`, linksCouple.ShortURL, linksCouple.OriginalURL)
	if err != nil {
		return "", fmt.Errorf("faild to insert entry in links_couples %w", err)
	}
	return linksCouple.ShortURL, nil
}
