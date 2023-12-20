package store

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

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
	s  map[string]LinksCouple
	fp *Producer
	l  *logger.ZapLog
	c  *config.Config
}

var ErrConflict = errors.New("data conflict")

func NewStore(c *config.Config, l *logger.ZapLog) (*Store, error) {
	if len(c.DBDSN) != 0 {
		db, err := sql.Open("pgx", c.DBDSN) //nolint:goconst // не понятно зачем константа
		if err != nil {
			l.ZL.Info("failed to open a connection to the DB in case to create store")
			return nil, fmt.Errorf("%w", errors.New("new store sql.open failed"))
		}
		// Проверяем через контекст из-за специфики работы sql.Open.
		// Устанавливаем таймаут 3 секудны на запрос.
		var limitTimeQuery = 20 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), limitTimeQuery)
		defer cancel()
		err = db.PingContext(ctx)
		if err != nil {
			l.ZL.Info("PingContext not nil in case to create store db NewStore")
			return nil, fmt.Errorf("new store pingcontext not nil in case to create store db %w", err)
		}
		// Отложенно закрываем соединение.
		defer func() {
			if err := db.Close(); err != nil {
				l.ZL.Info("new store failed to properly close the DB connection")
			}
		}()

		err = QueryCreateTableLinksCouples(ctx, db)
		if err != nil {
			return nil, fmt.Errorf("error in case to create table links_couples %w", err)
		}
	}
	if c.SFilePath != "" {
		var perm os.FileMode = 0600
		file, err := os.OpenFile(c.SFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize a store by file: %w", err)
		}
		return &Store{
			s:  make(map[string]LinksCouple),
			fp: &Producer{file: file, writer: bufio.NewWriter(file)},
			l:  l,
			c:  c,
		}, nil
	}
	return &Store{
		s:  make(map[string]LinksCouple),
		fp: nil,
		l:  l,
		c:  c,
	}, nil
}

func (s *Store) SetShortURL(longURL string) (string, error) {
	// Сюда приходит короткая ссылка без проверки на коллизии
	newShortLink := generatelinks.GenerateShortURL()

	// Если такой короткой ссылки еще нет в базе, значит можем спокойно записывать
	if _, ok := s.s[newShortLink]; !ok {
		// Создаем структуру по заданию и в нее записываем значение
		linksCouple := LinksCouple{UUID: "1", ShortURL: newShortLink, OriginalURL: longURL}
		// Заносим эту структуру в стор
		s.s[newShortLink] = linksCouple
		if len(s.c.DBDSN) != 0 {
			// добавляем в бд
			// Создаем подключение
			db, err := sql.Open("pgx", s.c.DBDSN)
			if err != nil {
				return "", fmt.Errorf("%w", errors.New("sql.open failed in case to create store"))
			}

			// Проверяем через контекст из-за специфики работы sql.Open.
			// Устанавливаем таймаут 3 секудны на запрос.
			var limitTimeQuery = 20 * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), limitTimeQuery)
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

			err = InsertLinksCouple(ctx, db, linksCouple)

			if err != nil {
				// проверяем, что ошибка сигнализирует о потенциальном нарушении целостности данных
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
					err = ErrConflict
				}
				return "", fmt.Errorf("failed to insert linkscouple in db %w", err)
			}
		}

		if s.c.SFilePath != "" {
			// Также записываем в файл
			err := s.fp.WriteLinksCouple(&linksCouple)
			if err != nil {
				delete(s.s, newShortLink)
				return "", fmt.Errorf("failed to write a new couple links in file %w", err)
			}
		}
		return newShortLink, nil
	}
	// Иначе у нас произошла коллизия
	return "", errors.New("a collision occurred")
}

func (s *Store) GetLongLinkByShort(shortURL string) (string, error) {
	if len(s.c.DBDSN) != 0 { //nolint:dupl // разные данные получаем
		// Создаем подключение
		db, err := sql.Open("pgx", s.c.DBDSN)
		if err != nil {
			return "", fmt.Errorf("%w", errors.New("sql.open failed in case to create store"))
		}
		// Проверяем через контекст из-за специфики работы sql.Open.
		// Устанавливаем таймаут 3 секудны на запрос.
		var limitTimeQuery = 20 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), limitTimeQuery)
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

func (s *Store) GetShortLinkByLong(originalURL string) (string, error) {
	if len(s.c.DBDSN) != 0 { //nolint:dupl // разные данные получаем
		// Создаем подключение
		db, err := sql.Open("pgx", s.c.DBDSN)
		if err != nil {
			return "", fmt.Errorf("%w", errors.New("sql.open failed in case to create store"))
		}
		// Проверяем через контекст из-за специфики работы sql.Open.
		// Устанавливаем таймаут 3 секудны на запрос.
		var limitTimeQuery = 20 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), limitTimeQuery)
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

		shortURL, err := GetShortURLByOriginalURL(ctx, db, originalURL)
		if err != nil {
			return "", fmt.Errorf("failed to get original URL %w", err)
		}
		return shortURL, nil
	}
	return "no match", nil
}

func (s *Store) ReadStoreFromFile(c *config.Config) {
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
