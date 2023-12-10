package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	c map[string]string
}

func NewConfig() (*Config, error) {
	config := &Config{
		c: make(map[string]string),
	}
	err := config.SetValues()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Config) SetValue(i, v string) error {
	c.c[i] = v
	return nil
}

func (c *Config) GetValueByIndex(i string) string {
	return c.c[i]
}

func (c *Config) SetValues() error {
	// неэкспортированная переменная flagRunAddr содержит адрес и порт для запуска сервера
	var runAddr string // localhost:8888
	// уровень логирования
	var logLevel string // localhost:8888
	// неэкспортированная переменная baseShortURL содержит базовый адрес короткой ссылки
	var baseShortURL string // localhost:8888
	// путь для файла где хранить все данные
	var sFilePath string
	// регистрируем переменную flagRunAddr как аргумент -a со значением по умолчанию :8080
	flag.StringVar(&runAddr, "a", "localhost:8080", "address and port to run server")
	// регистрируем уровень логирования
	flag.StringVar(&logLevel, "l", "info", "logger level")
	// регистрируем переменную baseShortUrl как аргумент -b со значением по умолчанию http://localhost:8080
	flag.StringVar(&baseShortURL, "b", "http://localhost:8080", "base prefix for the shortURL")
	// принимаем путь к файлу где хранить данные
	flag.StringVar(&sFilePath, "f", "./tmp/short-url-db.json", "file database")
	// парсим переданные серверу аргументы в зарегестрированные переменные
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		runAddr = envRunAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		baseShortURL = envBaseURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel = envLogLevel
	}

	if envSFilePath := os.Getenv("FILE_STORAGE_PATH"); envSFilePath != "" {
		sFilePath = envSFilePath
	}

	err := c.SetValue("runaddr", runAddr)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	err = c.SetValue("loglevel", logLevel)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	err = c.SetValue("baseshorturl", baseShortURL)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	err = c.SetValue("sfilepath", sFilePath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
