package config

import (
	"flag"
	"os"
	"time"

	"github.com/eampleev23/URLshortener/internal/logger"
)

type Config struct {
	RanAddr      string
	LogLevel     string
	BaseShortURL string
	SFilePath    string
	DBDSN        string
	TLimitQuery  time.Duration
	DatagenEC    int
}

func NewConfig(l *logger.ZapLog) (*Config, error) {
	config := &Config{TLimitQuery: 20 * time.Second} //nolint:gomnd //nomagik
	err := config.SetValues()
	if err != nil {
		return nil, err
	}
	l.ZL.Info("Config set success..")
	return config, nil
}

func (c *Config) SetValues() error {
	// регистрируем переменную flagRunAddr как аргумент -a со значением по умолчанию :8080
	flag.StringVar(&c.RanAddr, "a", "localhost:8080", "address and port to run server")
	// регистрируем уровень логирования
	flag.StringVar(&c.LogLevel, "l", "info", "logger level")
	// регистрируем переменную baseShortUrl как аргумент -b со значением по умолчанию http://localhost:8080
	flag.StringVar(&c.BaseShortURL, "b", "http://localhost:8080", "base prefix for the shortURL")
	// принимаем путь к файлу где хранить данные
	flag.StringVar(&c.SFilePath, "f", "/tmp/short-url-db.json", "file database")
	// принимаем строку подключения к базе данных
	flag.StringVar(&c.DBDSN, "d", "", "postgres database")
	// количество записей. генерирующихся по умолчанию
	flag.IntVar(&c.DatagenEC, "dg", 1, "entries count for data generation in case to use it")
	// парсим переданные серверу аргументы в зарегестрированные переменные
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		c.RanAddr = envRunAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		c.BaseShortURL = envBaseURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		c.LogLevel = envLogLevel
	}

	if envSFilePath := os.Getenv("FILE_STORAGE_PATH"); envSFilePath != "" {
		c.SFilePath = envSFilePath
	}
	if envDBDSN := os.Getenv("DATABASE_DSN"); envDBDSN != "" {
		c.DBDSN = envDBDSN
	}
	return nil
}
