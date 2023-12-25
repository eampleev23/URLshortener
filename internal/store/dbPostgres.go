package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func InsertLinksCouple(ctx context.Context, db *sql.DB, linksCouple LinksCouple) error {
	_, err := db.ExecContext(ctx, `INSERT INTO links_couples(uuid, short_url, original_url)
VALUES (DEFAULT, $1, $2)`, linksCouple.ShortURL, linksCouple.OriginalURL)
	if err != nil {
		return fmt.Errorf("faild to insert entry in links_couples %w", err)
	}
	return nil
}

func GetOriginalURLByShortURL(ctx context.Context, db *sql.DB, shortURL string) (string, error) {
	row := db.QueryRowContext(ctx,
		`SELECT original_url FROM links_couples WHERE short_url = $1 LIMIT 1`, shortURL,
	)
	// Готовим переменную для чтения результата
	var originalURL string

	err := row.Scan(&originalURL) // Разбираем результат
	if err != nil {
		return "", fmt.Errorf("faild to get originalURL %w", err)
	}
	return originalURL, nil
}

func GetShortURLByOriginalURL(ctx context.Context, db *sql.DB, originalURL string) (string, error) {
	row := db.QueryRowContext(ctx,
		"SELECT short_url "+
			"FROM links_couples WHERE original_url = $1 LIMIT 1", originalURL,
	)
	// Готовим переменную для чтения результата
	var shortURL string

	err := row.Scan(&shortURL) // Разбираем результат
	if err != nil {
		return "", fmt.Errorf("faild to get originalURL %w", err)
	}
	return shortURL, nil
}
