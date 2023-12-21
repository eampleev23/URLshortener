package store

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"os"

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
	ctx    context.Context
	useDB  bool
	useF   bool
	useM   bool
}

var ErrConflict = errors.New("data conflict")

func NewStore(c *config.Config, l *logger.ZapLog) (*Store, error) {
	if len(c.DBDSN) != 0 {
		// Используем только базу данных и не используем файл и озу
		// Создаем подключение один раз и дальше всегда тольоко его используем
		dbConn, err := sql.Open("pgx", c.DBDSN) //nolint:goconst // не понятно зачем константа
		if err != nil {
			return nil, fmt.Errorf("error create db connection %w", err)
		}

		// Отложенно закрываем соединение. Если закрыть, то перестает работать. Переносить в main?
		//defer func() {
		//	if err := dbConn.Close(); err != nil {
		//		l.ZL.Info("new store failed to properly close the DB connection")
		//	}
		//}()

		ctx := context.Background()
		err = QueryCreateTableLinksCouples(ctx, dbConn)
		if err != nil {
			return nil, fmt.Errorf("error create table %w", err)
		}
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
			s:      nil,
			fp:     &Producer{file: file, writer: bufio.NewWriter(file)},
			l:      l,
			c:      c,
			DBConn: nil,
			useDB:  false,
			useF:   true,
			useM:   false,
		}
		store.readStoreFromFile(c)
		return store, nil
	}
	// иначе используем ОЗУ
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

// SetShortURL генерирует короткую ссылку без коллизий, но это не точно
func (s *Store) SetShortURL(longURL string) (string, error) {
	// Сюда приходит короткая ссылка без проверки на коллизии
	newShortLink := generatelinks.GenerateShortURL()
	linksCouple := LinksCouple{ShortURL: newShortLink, OriginalURL: longURL}
	switch {
	case s.useDB:
		err := InsertLinksCouple(s.ctx, s.DBConn, linksCouple)
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
		break
	default:
		log.Printf("default")
		if _, ok := s.s[newShortLink]; !ok {
			log.Printf("Зашли в условие, что нет коллизии")
			s.s[newShortLink] = linksCouple
			return newShortLink, nil
		} else {
			// Иначе у нас произошла коллизия
			log.Printf("Зашли в условие, что есть коллизия")
			return "", errors.New("a collision occurred")
		}
		break
	}
	log.Println("s.useM=", s.useM)
	log.Println("s.useDB=", s.useDB)
	log.Println("s.useF=", s.useF)
	log.Printf("Зашли ниже всех кейсов")
	return "", errors.New("a collision occurred")
}

func (s *Store) GetLongLinkByShort(ctxR context.Context, shortURL string) (string, error) {
	if s.useDB { //nolint:dupl // разные данные получаем
		// Создаем подключение
		db, err := sql.Open("pgx", s.c.DBDSN)
		if err != nil {
			return "", fmt.Errorf("%w", errors.New("sql.open failed in case to create store"))
		}
		// Проверяем через контекст из-за специфики работы sql.Open.
		// Устанавливаем таймаут 3 секудны на запрос.
		ctx, cancel := context.WithTimeout(ctxR, s.c.TLimitQuery)
		defer cancel()
		err = db.PingContext(ctx)
		if err != nil {
			s.l.ZL.Info("PingContext not nil in case to create store db")
			return "", fmt.Errorf("pingcontext not nil in case to insert entry %w", err)
		}
		// Отложенно закрываем соединение.
		defer func() {
			if err := db.Close(); err != nil {
				s.l.ZL.Info("failed to properly close the DB connection")
			}
		}()

		originalURL, err := GetOriginalURLByShortURL(ctx, db, shortURL)
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
	if s.useDB { //nolint:dupl // разные данные получаем
		ctx, cancel := context.WithTimeout(ctxR, s.c.TLimitQuery)
		defer cancel()
		shortURL, err := GetShortURLByOriginalURL(ctx, s.DBConn, originalURL)
		if err != nil {
			return "", fmt.Errorf("failed to get original URL %w", err)
		}
		return shortURL, nil
	}
	return "no match", nil
}

func (s *Store) readStoreFromFile(c *config.Config) {
	var perm os.FileMode = 0600
	// открываем файл чтобы посчитать количество строк
	file, err := os.OpenFile(c.SFilePath, os.O_RDONLY|os.O_CREATE, perm)

	if err != nil {
		log.Printf("%s", err)
	}

	if err != nil {
		log.Printf("Error open file: %s", err)
	}

	countLines, err := LineCounter(file)
	if err != nil {
		log.Printf("%s", err)
	}

	if countLines > 0 {
		// добавляем каждую существующую строку в стор
		fc, err := NewConsumer(c.SFilePath)
		if err != nil {
			log.Printf("%s", err)
		}
		for i := 0; i < countLines; i++ {
			linksCouple, err := fc.ReadLinksCouple()
			if err != nil {
				log.Printf("%s", err)
			}
			fmt.Println("linksCouple=", linksCouple)
			s.s[linksCouple.ShortURL] = *linksCouple
		}
	}
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
