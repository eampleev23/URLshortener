package store

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func QueryCreateTableLinksCouples(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS links_couples (
        "uuid" int GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
        "short_url" VARCHAR(250) NOT NULL DEFAULT '',
        "original_url"  VARCHAR(1000) NOT NULL DEFAULT ''
      )`)
	if err != nil {
		return fmt.Errorf("faild to create table links_couples %w", err)
	}
	return nil
}
func InsertLinksCouple(ctx context.Context, db *sql.DB, linksCouple LinksCouple) error {
	_, err := db.ExecContext(ctx, `INSERT INTO links_couples(uuid, short_url, original_url)
VALUES (DEFAULT, $1, $2)`, linksCouple.ShortURL, linksCouple.OriginalURL)
	if err != nil {
		return fmt.Errorf("faild to insert entry in links_couples %w", err)
	}
	return nil
}
