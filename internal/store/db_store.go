package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net/url"
	"time"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/generatelinks"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
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

// SetShortURL вставляет в бд новую строку или возвращает специфическую ошибку в случае конфликта.
func (ds DBStore) SetShortURL(ctx context.Context, originalURL string, ownerID int) (newShortURL string, err error) {
	newShortURL, err = ds.InsertURL(ctx, LinksCouple{ShortURL: generatelinks.GenerateShortURL(), OriginalURL: originalURL, OwnerID: ownerID})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		err = ErrConflict
		return "", fmt.Errorf("conflict: %w", err)
	}
	if err != nil {
		return "", fmt.Errorf("error InsertURL: %w", err)
	}
	ds.l.ZL.Info("Успешно добавили новую ссылку", zap.String("newShortURL", newShortURL))
	ds.l.ZL.Info("ID пользователя", zap.Int("ownerID", ownerID))
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
		return fmt.Errorf("failed to properly close the DB connection %w", err)
	}
	return nil
}

// CreateTable создает таблицу если ее нет при инициализации стора. Без этого не работают тесты 11 инкремента.
func (ds DBStore) createTable() error {
	ctx := context.Background()
	defer ctx.Done()
	_, err := ds.dbConn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS links_couples (
        "uuid" int GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
        "short_url" VARCHAR(250) NOT NULL DEFAULT '',
        "original_url"  VARCHAR(1000) NOT NULL DEFAULT '',
        "owner_id"  INT NOT NULL DEFAULT 0
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

// InsertURL занимается непосредственно запросом вставки строки в бд.
func (ds DBStore) InsertURL(ctx context.Context, linksCouple LinksCouple) (shortURL string, err error) {
	_, err = ds.dbConn.ExecContext(ctx, `INSERT INTO links_couples(uuid, short_url, original_url, owner_id)
VALUES (DEFAULT, $1, $2, $3)`, linksCouple.ShortURL, linksCouple.OriginalURL, linksCouple.OwnerID)
	if err != nil {
		return "", fmt.Errorf("faild to insert entry in links_couples %w", err)
	}
	return linksCouple.ShortURL, nil
}

func (ds DBStore) GetURLsByOwnerID(ctx context.Context, ownerID int) ([]LinksCouple, error) {
	rows, err := ds.dbConn.QueryContext(ctx, "SELECT * FROM links_couples WHERE owner_id = $1", ownerID)
	fmt.Println("ownerID", ownerID)
	if err != nil {
		return nil, fmt.Errorf("error get links for owner by ownerid %w", err)
	}
	// обязательно закрываем перед возвратом функции
	defer rows.Close()
	// Готовим переменную для чтения результата
	var linksCouples []LinksCouple
	for rows.Next() {
		var v LinksCouple
		err = rows.Scan(&v.UUID, &v.ShortURL, &v.OriginalURL, &v.OwnerID)
		if err != nil {
			return nil, err
		}
		v.ShortURL, err = url.JoinPath(ds.c.BaseShortURL, v.ShortURL)
		if err != nil {
			return nil, fmt.Errorf("error url.JoinPath: %w", err)
		}
		fmt.Println(v)
		linksCouples = append(linksCouples, v)
	}
	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return linksCouples, nil
}
