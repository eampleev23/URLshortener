package store

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/generatelinks"
	"github.com/eampleev23/URLshortener/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type LinksCouple struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Store struct {
	s      map[string]LinksCouple
	fp     *Producer
	l      *logger.ZapLog
	c      *config.Config
	DBConn *sql.DB
	ctx    context.Context //nolint:containedctx // надо разбираться один на один
	useDB  bool
	useF   bool
	useM   bool
}

var ErrConflict = errors.New("data conflict")

func NewStore(c *config.Config, l *logger.ZapLog) (*Store, error) {
	if len(c.DBDSN) != 0 {
		// Используем только базу данных и не используем файл и озу
		// Создаем подключение один раз и дальше всегда тольоко его используем
		dbConn, err := sql.Open("pgx", c.DBDSN)
		if err != nil {
			return nil, fmt.Errorf("error create db connection %w", err)
		}

		// Отложенно закрываем соединение. Если закрыть, то перестает работать. Переносить в main?
		// defer func() {
		//	if err := dbConn.Close(); err != nil {
		//		l.ZL.Info("new store failed to properly close the DB connection")
		//	}
		// }()

		ctx := context.Background()
		return &Store{
			s:      nil,
			fp:     nil,
			l:      l,
			c:      c,
			DBConn: dbConn,
			ctx:    ctx,
			useDB:  true,
			useF:   false,
			useM:   false,
		}, nil
	}
	if c.SFilePath != "" {
		// используем только файл
		var perm os.FileMode = 0600
		file, err := os.OpenFile(c.SFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize a store by file: %w", err)
		}

		store := &Store{
			s:      make(map[string]LinksCouple),
			fp:     &Producer{file: file, writer: bufio.NewWriter(file)},
			l:      l,
			c:      c,
			DBConn: nil,
			useDB:  false,
			useF:   true,
			useM:   false,
		}
		err = store.readStoreFromFile(c)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		return store, nil
	}
	// Иначе используем только ОЗУ.
	return &Store{
		s:      make(map[string]LinksCouple),
		fp:     nil,
		l:      l,
		c:      c,
		DBConn: nil,
		useDB:  false,
		useF:   false,
		useM:   true,
	}, nil
}

// SetShortURL генерирует короткую ссылку без коллизий, но это не точно.
func (s *Store) SetShortURL(ctxR context.Context, longURL string) (string, error) {
	// Сюда приходит короткая ссылка без проверки на коллизии
	newShortLink := generatelinks.GenerateShortURL()
	linksCouple := LinksCouple{ShortURL: newShortLink, OriginalURL: longURL}
	switch {
	case s.useDB:
		err := s.InsertLinksCouple(ctxR, linksCouple)
		if err != nil {
			// проверяем, что ошибка сигнализирует о потенциальном нарушении целостности данных
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				err = ErrConflict
			}
			return "", fmt.Errorf("failed to insert linkscouple in db %w", err)
		}
		return newShortLink, nil
	case s.useF:
		err := s.fp.WriteLinksCouple(&linksCouple)
		if err != nil {
			delete(s.s, newShortLink)
			return "", fmt.Errorf("failed to write a new couple links in file %w", err)
		}
		if _, ok := s.s[newShortLink]; !ok {
			s.s[newShortLink] = linksCouple
			return newShortLink, nil
		} else {
			// Иначе у нас произошла коллизия
			return "", errors.New("a collision occurred")
		}
	default:
		if _, ok := s.s[newShortLink]; !ok {
			s.s[newShortLink] = linksCouple
			return newShortLink, nil
		} else {
			// Иначе у нас произошла коллизия
			return "", errors.New("a collision occurred")
		}
	}
}

func (s *Store) GetLongLinkByShort(ctxR context.Context, shortURL string) (string, error) {
	if s.useDB {
		originalURL, err := s.GetOriginalURLByShortURL(ctxR, shortURL)
		if err != nil {
			return "", fmt.Errorf("failed to get original URL %w", err)
		}
		return originalURL, nil
	}

	if c, ok := s.s[shortURL]; ok {
		return c.OriginalURL, nil
	}
	return "no match", nil
}

func (s *Store) GetShortLinkByLong(ctxR context.Context, originalURL string) (string, error) {
	if s.useDB {
		shortURL, err := s.GetShortURLByOriginalURL(ctxR, originalURL)
		if err != nil {
			return "", fmt.Errorf("failed to get original URL %w", err)
		}
		return shortURL, nil
	}
	return "no match", nil
}

func (s *Store) readStoreFromFile(c *config.Config) error {
	var perm os.FileMode = 0600
	// открываем файл чтобы посчитать количество строк
	file, err := os.OpenFile(c.SFilePath, os.O_RDONLY|os.O_CREATE, perm)
	if err != nil {
		return fmt.Errorf("ошибка при попытке открытия файла при стартовой загрузке данных: %w", err)
	}
	countLines, err := LineCounter(file)
	if err != nil {
		return fmt.Errorf("ошибка при подсчете количества строк в файле при стартовой загрузке данных: %w", err)
	}
	if countLines > 0 {
		// добавляем каждую существующую строку в стор
		fc, err := NewConsumer(c.SFilePath)
		if err != nil {
			return fmt.Errorf("ошибка при создании читателя файла при стартовом получении данных: %w", err)
		}
		for i := 0; i < countLines; i++ {
			linksCouple, err := fc.ReadLinksCouple()
			if err != nil {
				return fmt.Errorf("ошибка при чтении строки файла при стартовом получении данных: %w", err)
			}
			s.s[linksCouple.ShortURL] = *linksCouple
		}
	}
	return nil
}

func (s *Store) PingDB(ctxR context.Context) (bool, error) {
	// Проверяем через контекст из-за специфики работы sql.Open.
	ctx, cancel := context.WithTimeout(ctxR, s.c.TLimitQuery)
	defer cancel()
	err := s.DBConn.PingContext(ctx)
	if err != nil {
		return false, fmt.Errorf("db doesn't ping %w", err)
	}
	return true, nil
}
