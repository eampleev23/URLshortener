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

func NewProducer(filename string) (*Producer, error) {
	// открываем файл для записи в конец
	/*
		Линтер ругается на 0600, не понятно где взть константу, самому определить?
	*/
	var perm os.FileMode = 0600
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &Producer{file: file}, nil
}

func (p *Producer) WriteLinksCouple(linksCouple *LinksCouple) error {
	data, err := json.Marshal(&linksCouple)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// записываем пару ссылок в буфер
	if _, err := p.writer.Write(data); err != nil {
		return fmt.Errorf("%w", err)
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("%w", err)
	}

	// записываем буфер в файл
	return p.writer.Flush()
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
		return nil, err
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
		return nil, c.scanner.Err()
	}
	// читаем данные из scanner
	data := c.scanner.Bytes()

	linksCouple := LinksCouple{}
	err := json.Unmarshal(data, &linksCouple)
	if err != nil {
		return nil, err
	}

	return &linksCouple, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
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
			return count, err
		}
	}
}
