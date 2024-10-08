package store

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
)

// FileStore - хранилище в файле.
type FileStore struct {
	Producer *Producer
	Consumer *Consumer
	ms       *MemoryStore
}

// Producer - писатель данных.
type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

// Consumer - читатель данных.
type Consumer struct {
	file *os.File
	// заменяем Reader на Scanner
	scanner *bufio.Scanner
}

// NewFileStore - конструктор.
func NewFileStore(c *config.Config, l *logger.ZapLog) (*FileStore, error) {
	var perm os.FileMode = 0600
	file, err := os.OpenFile(c.SFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a store by file: %w", err)
	}
	pr := Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}
	co, err := NewConsumer(c.SFilePath)
	if err != nil {
		return nil, fmt.Errorf("error creating NewConsumer: %w", err)
	}
	ms, err := NewMemoryStore(c, l)
	if err != nil {
		return nil, fmt.Errorf("error create memory store inside filestore: %w", err)
	}
	fs := &FileStore{
		Producer: &pr,
		Consumer: co,
		ms:       ms,
	}
	err = fs.ReadStoreFromFile(c)
	if err != nil {
		return nil, fmt.Errorf("ReadStoreFromFile error: %w", err)
	}
	return fs, nil
}

// SetShortURL - хорош.
func (fs *FileStore) SetShortURL(ctx context.Context, originalURL string, ownerID int) (newShortURL string, err error) {
	newShortURL, err = fs.ms.SetShortURL(ctx, originalURL, ownerID)
	if err != nil {
		return "", fmt.Errorf("error set in memory store in file store: %w", err)
	}
	linksCouple := LinksCouple{UUID: "1", ShortURL: newShortURL, OriginalURL: originalURL}
	err = fs.Producer.WriteLinksCouple(&linksCouple)
	if err != nil {
		return "", fmt.Errorf("error write links couple in file store %w", err)
	}
	return newShortURL, nil
}

// GetOriginalURLByShort - хорош.
func (fs *FileStore) GetOriginalURLByShort(ctx context.Context, shortURL string) (originalURL string, err error) {
	originalURL, err = fs.ms.GetOriginalURLByShort(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("error GetOriginalURLByShort in file store %w", err)
	}
	return originalURL, nil
}

// GetShortURLByOriginal - хм.
func (fs *FileStore) GetShortURLByOriginal(ctx context.Context, originalURL string) (shortURL string, err error) {
	shortURL, err = fs.ms.GetShortURLByOriginal(ctx, originalURL)
	if err != nil {
		return "", fmt.Errorf("error GetShortURLByOriginal in file store %w", err)
	}
	return shortURL, nil
}

// GetURLsByOwnerID - хм.
func (fs *FileStore) GetURLsByOwnerID(ctx context.Context, ownerID int) ([]LinksCouple, error) {
	return nil, errors.New("file store doesn't use this method")
}

// PingDB проверяет подключение к базе данных.
func (fs *FileStore) PingDB(ctx context.Context, tiimeLimit time.Duration) (err error) {
	return nil
}

// Close закрывает соединение с базой данных.
func (fs *FileStore) Close() (err error) {
	return nil
}

// WriteLinksCouple - хм.
func (p *Producer) WriteLinksCouple(linksCouple *LinksCouple) error {
	data, err := json.Marshal(&linksCouple)
	if err != nil {
		return fmt.Errorf("ошибка после json.Marshal(&linksCouple) writeLinksCouple %w", err)
	}
	// записываем пару ссылок в буфер
	if _, err := p.writer.Write(data); err != nil {
		return fmt.Errorf("ошибка при записи в буфер p.writer.Write(data) %w", err)
	}
	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("ошибка при добавлении переноса строки %w", err)
	}
	// записываем буфер в файл
	err = p.writer.Flush()
	if err != nil {
		return fmt.Errorf("failed in case call p.writer.Flush():  %w", err)
	}
	return nil
}

// NewConsumer - конструктор читателя.
func NewConsumer(filename string) (*Consumer, error) {
	var perm os.FileMode = 0600
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, perm)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a NewConsumer: %w", err)
	}

	return &Consumer{
		file: file,
		// создаём новый scanner
		scanner: bufio.NewScanner(file),
	}, nil
}

// ReadLinksCouple - прочитать строку.
func (c *Consumer) ReadLinksCouple() (*LinksCouple, error) {
	// одиночное сканирование до следующей строки
	if !c.scanner.Scan() {
		return nil, fmt.Errorf("failed to read a new string from file by c.scanner.Scan: %w", c.scanner.Err())
	}
	// читаем данные из scanner
	data := c.scanner.Bytes()

	linksCouple := LinksCouple{}
	err := json.Unmarshal(data, &linksCouple)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal in case to read file: %w", err)
	}

	return &linksCouple, nil
}

// Close - закрыть поток.
func (c *Consumer) Close() error {
	err := c.file.Close()
	if err != nil {
		return fmt.Errorf("failed to close consumer in case read store from file: %w", err)
	}
	return nil
}

// LineCounter - считает количество строк в файле при инициализации стора.
func LineCounter(r io.Reader) (int, error) {
	bufSize := 32768
	buf := make([]byte, bufSize)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, fmt.Errorf("failed to count lines in file storage: %w", err)
		}
	}
}

// ReadStoreFromFile - прочитать.
func (fs *FileStore) ReadStoreFromFile(c *config.Config) error {
	var perm os.FileMode = 0600
	// открываем файл чтобы посчитать количество строк
	file, err := os.OpenFile(c.SFilePath, os.O_RDONLY|os.O_CREATE, perm)

	if err != nil {
		return fmt.Errorf("error oppening or creating file: %w", err)
	}

	countLines, err := LineCounter(file)
	if err != nil {
		return fmt.Errorf("error counting lines in the file: %w", err)
	}

	if countLines > 0 {
		// добавляем каждую существующую строку в стор
		fc, err := NewConsumer(c.SFilePath)
		if err != nil {
			return fmt.Errorf("NewConsumer error: %w", err)
		}
		for i := 0; i < countLines; i++ {
			linksCouple, err := fc.ReadLinksCouple()
			if err != nil {
				return fmt.Errorf("fc.ReadLinksCouple %v error: %w", linksCouple, err)
			}
			fs.ms.s[linksCouple.ShortURL] = *linksCouple
		}
	}
	return nil
}

// DeleteURLS - удалить.
func (fs *FileStore) DeleteURLS(ctx context.Context, deleteItems []DeleteURLItem) (err error) {
	err = fs.ms.DeleteURLS(ctx, deleteItems)
	if err != nil {
		return fmt.Errorf("memory store delete URL %w", err)
	}
	return nil
}

// GetLinksCoupleByShortURL - получить строку по короткой ссылке.
func (fs *FileStore) GetLinksCoupleByShortURL(ctx context.Context, shortURL string) (lc LinksCouple, err error) {
	originalURL, err := fs.ms.GetOriginalURLByShort(ctx, shortURL)
	if err != nil {
		return LinksCouple{}, fmt.Errorf("error GetOriginalURLByShort in file store %w", err)
	}
	return LinksCouple{UUID: "1", ShortURL: shortURL, OriginalURL: originalURL, OwnerID: 1, DeletedFlag: false}, nil
}

// GetURLsCount - делает запрос суммы сохраненных урлов в сервисе.
func (fs *FileStore) GetURLsCount(ctx context.Context) (count int64, err error) {
	count, err = fs.ms.GetURLsCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("error GetURLsCount in file store %w", err)
	}
	return count, nil
}

// GetUsersCount - делает запрос общего количества пользователей в сервисе.
func (fs *FileStore) GetUsersCount(ctx context.Context) (count int64, err error) {
	count, err = fs.ms.GetUsersCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("error GetUsersCount in file store %w", err)
	}
	return count, nil
}
