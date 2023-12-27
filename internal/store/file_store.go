package store

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/generatelinks"
	"github.com/eampleev23/URLshortener/internal/logger"
	"io"
	"log"
	"os"
	"time"
)

type FileStore struct {
	Producer *Producer
	Consumer *Consumer
}
type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}
type Consumer struct {
	file *os.File
	// заменяем Reader на Scanner
	scanner *bufio.Scanner
}

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
	return &FileStore{
		Producer: &pr,
		Consumer: co,
	}, nil
}

func (fs *FileStore) SetShortURL(ctx context.Context, originalURL string) (newShortURL string, err error) {
	newShortURL = generatelinks.GenerateShortURL()
	linksCouple := LinksCouple{UUID: "1", ShortURL: newShortURL, OriginalURL: originalURL}
	err = fs.Producer.WriteLinksCouple(&linksCouple)
	if err != nil {
		return "", fmt.Errorf("error write links couple in file store %w", err)
	}
	return newShortURL, nil
}
func (fs *FileStore) GetOriginalURLByShort(ctx context.Context, shortURL string) (originalURL string, err error) {
	originalURL = ""
	return originalURL, nil
}
func (fs *FileStore) GetShortURLByOriginal(ctx context.Context, originalURL string) (shortURL string, err error) {
	shortURL = ""
	return shortURL, nil
}

// PingDB проверяет подключение к базе данных
func (fs *FileStore) PingDB(ctx context.Context, tiimeLimit time.Duration) (err error) {
	return nil
}

// Close закрывает соединение с базой данных
func (fs *FileStore) Close() (err error) {
	return nil
}

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
func (s *FileStore) ReadStoreFromFile(c *config.Config) {
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
			//s.s[linksCouple.ShortURL] = *linksCouple
		}
	}
}