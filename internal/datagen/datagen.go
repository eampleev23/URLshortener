package datagen

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"go.uber.org/zap"
)

func GenerateData(ctx context.Context, cfg *config.Config, l *logger.ZapLog) error {
	if len(cfg.DBDSN) == 0 {
		return fmt.Errorf("passed DSN is empty")
	}
	if cfg.DatagenEC <= 0 || cfg.DatagenEC > 100_000_000 {
		return fmt.Errorf(
			"expected employees count to be 0 <= count <= 1_000_000, got: %d",
			cfg.DatagenEC,
		)
	}
	db, err := sql.Open("pgx", cfg.DBDSN)
	if err != nil {
		return fmt.Errorf("failed to open a connection to the DB: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			l.ZL.Info("Failed to properly close the DB connection: ", zap.Error(err))
		}
	}()
	if err := createSchema(ctx, db, l); err != nil {
		return fmt.Errorf("failed to create the DB schema: %w", err)
	}
	if err := generateData(ctx, db, cfg.DatagenEC, l); err != nil {
		return fmt.Errorf("failed to generate DB data: %w", err)
	}

	return nil
}

func createSchema(ctx context.Context, db *sql.DB, l *logger.ZapLog) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				l.ZL.Info("Failed create schema: ", zap.Error(err))
			}
		}
	}()
	createSchemaStmts := []string{
		`CREATE TABLE IF NOT EXISTS links_couples (
        "uuid" int GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
        "short_url" VARCHAR(250) NOT NULL DEFAULT '',
        "original_url"  VARCHAR(1000) NOT NULL DEFAULT ''
      );
      CREATE UNIQUE INDEX IF NOT EXISTS links_couples_index_by_original_url_unique
    ON links_couples
        USING btree (original_url);
      `,
	}
	for _, stmt := range createSchemaStmts {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute statement `%s`: %w", stmt, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit the transaction: %w", err)
	}
	return nil
}

func generateData(ctx context.Context, db *sql.DB, entrCount int, l *logger.ZapLog) error {
	// Генерируем данные для таблицы links_couples
	if err := generateLinksCouplesData(ctx, db, generateLinksCouplesDataOpts{
		Count: entrCount,
	}, l); err != nil {
		return fmt.Errorf("failed to generate the links couples data: %w", err)
	}
	return nil
}

type generateLinksCouplesDataOpts struct {
	Count int
}

func generateLinksCouplesData(
	ctx context.Context,
	db *sql.DB,
	opts generateLinksCouplesDataOpts,
	l *logger.ZapLog,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				l.ZL.Info("Failed to rollback the transaction: ", zap.Error(err))
			}
		}
	}()
	// а вот и батчинг
	const batchSize = 10000
	// оптс приходит со значением? каким - entrCount
	for opts.Count-batchSize >= 0 {
		stmt := generateLinksCouplesStatement(batchSize)
		fields := generateLinksCouplesFields(batchSize)
		if _, err := tx.Exec(stmt, fields...); err != nil {
			return fmt.Errorf("failed to insert a batch: %w", err)
		}
		opts.Count -= batchSize
		// just do it.
		l.ZL.Debug("Count: ", zap.Int("count", opts.Count))
		l.ZL.Debug("BatchSize: ", zap.Int("batchSize", opts.Count))
	}
	if opts.Count > 0 {
		stmt := generateLinksCouplesStatement(opts.Count)
		fields := generateLinksCouplesFields(opts.Count)
		if _, err := tx.Exec(stmt, fields...); err != nil {
			return fmt.Errorf("failed to insert a batch with: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit the transaction: %w", err)
	}
	return nil
}
func generateLinksCouplesStatement(count int) string {
	const stmtTmpl = `INSERT INTO links_couples(short_url, original_url)
VALUES %s`

	valuesParts := make([]string, 0, count)
	numColumns := 2
	for i := 0; i < count; i++ {
		valuesParts = append(
			valuesParts,
			fmt.Sprintf(
				"($%d, $%d)",
				i*numColumns+1, i*numColumns+2, //nolint:gomnd // not magik
			),
		)
	}
	return fmt.Sprintf(stmtTmpl, strings.Join(valuesParts, ","))
}
func generateLinksCouplesFields(count int) []any {
	values := make([]any, 0, 2*count) //nolint:gomnd // not magik
	for i := 0; i < count; i++ {
		values = append(
			values,
			gofakeit.FirstName(),
			gofakeit.URL(),
		)
	}
	return values
}
