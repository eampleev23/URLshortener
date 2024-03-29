package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	RanAddr        string
	LogLevel       string
	BaseShortURL   string
	SFilePath      string
	DBDSN          string
	SecretKey      string
	DatagenEC      int
	TLimitQuery    time.Duration
	TokenEXP       time.Duration
	TimeDeleteURLs time.Duration
}

func NewConfig() (*Config, error) {
	config := &Config{
		TLimitQuery:    20 * time.Second, //nolint:gomnd //nomagik
		TokenEXP:       time.Hour * 3,    //nolint:gomnd //nomagik
		TimeDeleteURLs: time.Second * 5,  //nolint:gomnd //nomagik
	}
	err := config.SetValues()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Config) SetValues() error {
	// регистрируем переменную flagRunAddr как аргумент -a со значением по умолчанию :8080
	flag.StringVar(&c.RanAddr, "a", "localhost:8080", "address and port to run server")
	// регистрируем уровень логирования
	flag.StringVar(&c.LogLevel, "l", "debug", "logger level")
	// регистрируем переменную baseShortUrl как аргумент -b со значением по умолчанию http://localhost:8080
	flag.StringVar(&c.BaseShortURL, "b", "http://localhost:8080", "base prefix for the shortURL")
	// принимаем путь к файлу где хранить данные
	flag.StringVar(&c.SFilePath, "f", "/tmp/short-url-db.json", "file database")
	// принимаем строку подключения к базе данных
	flag.StringVar(&c.DBDSN, "d", "", "postgres database")
	// количество записей. генерирующихся по умолчанию
	flag.IntVar(&c.DatagenEC, "dg", 1, "entries count for data generation in case to use it")
	// принимаем секретный ключ сервера для авторизации
	flag.StringVar(&c.SecretKey, "s", "e4853f5c4810101e88f1898db21c15d3", "server's secret key for authorization")
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
	if envSecretKey := os.Getenv("SECRET_KEY"); envSecretKey != "" {
		c.SecretKey = envSecretKey
	}
	return nil
}
