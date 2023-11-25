package config

import (
	"flag"
	"os"
)

type Config struct {
	c map[string]string
}

func NewConfig() *Config {
	return &Config{
		c: make(map[string]string),
	}
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
	var runAddr string  // localhost:8888
	var logLevel string // localhost:8888

	// неэкспортированная переменная baseShortURL содержит базовый адрес короткой ссылки
	var baseShortURL string // localhost:8888

	// регистрируем переменную flagRunAddr как аргумент -a со значением по умолчанию :8080
	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&logLevel, "l", "info", "logger level")

	// регистрируем переменную baseShortUrl как аргумент -b со значением по умолчанию http://localhost:8080
	flag.StringVar(&baseShortURL, "b", "http://localhost:8080", "base prefix for the shortURL")
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
	c.SetValue("runaddr", runAddr)
	c.SetValue("loglevel", logLevel)
	c.SetValue("baseshorturl", baseShortURL)
	return nil
}
