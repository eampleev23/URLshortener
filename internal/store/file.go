package store

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
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

type Consumer struct {
	file *os.File
	// заменяем Reader на Scanner
	scanner *bufio.Scanner
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
