package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"go.uber.org/zap"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/generatelinks"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DBStore - класс хранилища.
type DBStore struct {
	dbConn *sql.DB
	c      *config.Config
	l      *logger.ZapLog
}

// NewDBStore - конструктор дб стора.
func NewDBStore(c *config.Config, l *logger.ZapLog) (*DBStore, error) {
	db, err := sql.Open("pgx", c.DBDSN)
	if err != nil {
		return &DBStore{}, fmt.Errorf("%w", errors.New("sql.open failed in case to create store"))
	}
	if err := runMigrations(c.DBDSN); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	return &DBStore{
		dbConn: db,
		c:      c,
		l:      l,
	}, nil
}

// SetShortURL вставляет в бд новую строку или возвращает специфическую ошибку в случае конфликта.
func (ds DBStore) SetShortURL(ctx context.Context, originalURL string, ownerID int) (newShortURL string, err error) {
	newShortURL, err = ds.InsertURL(
		ctx,
		LinksCouple{
			ShortURL:    generatelinks.GenerateShortURL(),
			OriginalURL: originalURL, OwnerID: ownerID,
		})
	if err != nil {
		return "", fmt.Errorf("error InsertURL: %w", err)
	}
	ds.l.ZL.Debug("Успешно добавили новую ссылку", zap.String("newShortURL", newShortURL))
	ds.l.ZL.Debug("ID пользователя", zap.Int("ownerID", ownerID))
	return newShortURL, nil
}

// InsertURL занимается непосредственно запросом вставки строки в бд.
func (ds DBStore) InsertURL(ctx context.Context, linksCouple LinksCouple) (shortURL string, err error) {
	_, err = ds.dbConn.ExecContext(ctx, `INSERT INTO links_couples(uuid, short_url, original_url, owner_id)
VALUES (DEFAULT, $1, $2, $3)`, linksCouple.ShortURL, linksCouple.OriginalURL, linksCouple.OwnerID)
	// Проверяем, что ошибка сигнализирует о потенциальном нарушении целостности данных
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		err = ErrConflict
	}
	return linksCouple.ShortURL, err
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

// GetOriginalURLByShort - получить оригинальную ссылку по короткой.
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

// GetShortURLByOriginal - получить короткую ссылку по оригинальной.
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

// PingDB - пингануть бд.
func (ds DBStore) PingDB(ctx context.Context, timeLimit time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeLimit)
	defer cancel()
	err := ds.dbConn.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("db doesn't ping %w", err)
	}
	return nil
}

// Close - закрыть соединение с бд.
func (ds DBStore) Close() error {
	if err := ds.dbConn.Close(); err != nil {
		return fmt.Errorf("failed to properly close the DB connection %w", err)
	}
	return nil
}

// GetURLsByOwnerID - получить урлы пользователя по его ид.
func (ds DBStore) GetURLsByOwnerID(ctx context.Context, ownerID int) ([]LinksCouple, error) {
	rows, err := ds.dbConn.QueryContext(ctx, "SELECT uuid, short_url, original_url, owner_id, is_deleted"+
		" FROM links_couples WHERE owner_id = $1", ownerID)
	if err != nil {
		return nil, fmt.Errorf("not get links for owner by ownerid %w", err)
	}
	// обязательно закрываем перед возвратом функции
	// Отложенно закрываем соединение с бд.
	defer func() {
		if err := rows.Close(); err != nil {
			ds.l.ZL.Info("error defer rows.Close() in GetURLsByOwnerID")
		}
	}()
	// Готовим переменную для чтения результата
	var linksCouples []LinksCouple
	for rows.Next() {
		var v LinksCouple
		err = rows.Scan(&v.UUID, &v.ShortURL, &v.OriginalURL, &v.OwnerID, &v.DeletedFlag)
		if err != nil {
			return nil, fmt.Errorf("error rows.Scan in GetURLsByOwnerID: %w", err)
		}
		linksCouples = append(linksCouples, v)
	}
	// проверяем на ошибки
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err in GetURLsByOwnerID: %w", err)
	}
	return linksCouples, nil
}

func updateLinksCouplesStatement(deleteItems []DeleteURLItem) string {
	valueParts := ""
	count := len(deleteItems)
	for i := 0; i < count; i++ {
		if i == count-1 {
			valueParts += fmt.Sprintf("('%s', %t, %d)",
				deleteItems[i].ShortURL,
				deleteItems[i].DeleteFlag,
				deleteItems[i].OwnerID)
		} else {
			valueParts += fmt.Sprintf("('%s', %t, %d), ",
				deleteItems[i].ShortURL,
				deleteItems[i].DeleteFlag,
				deleteItems[i].OwnerID)
		}
	}
	stmtResult := `UPDATE links_couples SET is_deleted = tmp.is_deleted FROM (VALUES ` + valueParts +
		`) as tmp (short_url, is_deleted, owner_id) WHERE links_couples.short_url=tmp.short_url
					AND links_couples.owner_id=tmp.owner_id;`
	return stmtResult
}

// DeleteURLS проставляет тем урлам флаг удаления, которые пользователь решает удалить.
func (ds DBStore) DeleteURLS(ctx context.Context, deleteItems []DeleteURLItem) (err error) {
	log.Println("зашли в DeleteURLS")
	// Запускаем транзакцию.
	tx, err := ds.dbConn.BeginTx(ctx, nil)

	// Обрабатываем ошибку.
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %w", err)
	}

	// Отложенно откатываем транзакцию если err != nil.
	defer func() {
		if err := tx.Rollback(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				ds.l.ZL.Info("Failed to rollback the transaction: ", zap.Error(err))
			}
		}
	}()

	// Подготавливаем запрос.
	stmt := updateLinksCouplesStatement(deleteItems)
	if _, err := tx.Exec(stmt); err != nil {
		return fmt.Errorf("failed to update a batch with: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit the transaction: %w", err)
	}
	return nil
}

// GetLinksCoupleByShortURL - получить всю строку по короткой ссылке.
func (ds DBStore) GetLinksCoupleByShortURL(ctx context.Context, shortURL string) (lc LinksCouple, err error) {
	row := ds.dbConn.QueryRowContext(ctx,
		`SELECT uuid, short_url, original_url, owner_id, is_deleted 
FROM links_couples WHERE short_url = $1 LIMIT 1`, shortURL,
	)
	err = row.Scan(&lc.UUID, &lc.ShortURL, &lc.OriginalURL, &lc.OwnerID, &lc.DeletedFlag) // Разбираем результат
	if err != nil {
		return LinksCouple{}, fmt.Errorf("faild to get links couple by row %w", err)
	}
	return lc, nil
}

// GetURLsCount - делает запрос суммы сохраненных урлов в сервисе.
func (ds DBStore) GetURLsCount(ctx context.Context) (count int64, err error) {
	ds.l.ZL.Debug("DBStore / GetURLsCount has started..")
	row := ds.dbConn.QueryRowContext(ctx,
		`SELECT count(*) FROM links_couples;`)
	err = row.Scan(&count) // Разбираем результат
	if err != nil {
		return count, fmt.Errorf("faild to get count of entries %w", err)
	}
	ds.l.ZL.Debug("DBStore / GetURLsCount / Got count of the entries", zap.Int64("count", count))
	return count, nil
}

// GetUsersCount - делает запрос общего количества пользователей в сервисе.
func (ds DBStore) GetUsersCount(ctx context.Context) (count int64, err error) {
	ds.l.ZL.Debug("DBStore / GetUsersCount has started..")
	row := ds.dbConn.QueryRowContext(ctx,
		`SELECT COUNT(*) AS frequency
	FROM links_couples
	GROUP BY owner_id;`)
	err = row.Scan(&count) // Разбираем результат
	if err != nil {
		return count, fmt.Errorf("faild to get count of uniq users %w", err)
	}
	ds.l.ZL.Debug("DBStore / GetUsersCount / Got count of users", zap.Int64("count", count))
	return count, nil
}
