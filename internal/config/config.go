// Package config - конфигурация приложения.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

// Config - класс, хранящий все необходимые для конфигурации приложения поля.
type Config struct {
	RanAddr        string        `json:"server_address"`
	LogLevel       string        `json:"-"`
	BaseShortURL   string        `json:"base_url"`
	SFilePath      string        `json:"file_storage_path"`
	DBDSN          string        `json:"database_dsn"`
	SecretKey      string        `json:"-"`
	DatagenEC      int           `json:"-"`
	UseHTTPS       bool          `json:"enable_https"`
	FileConfigPath string        `json:"-"`
	TLimitQuery    time.Duration `json:"-"`
	TokenEXP       time.Duration `json:"-"`
	TimeDeleteURLs time.Duration `json:"-"`
	TrustedSubnet  string        `json:"trusted_subnet"`
}

// NewConfig - конструктор конфига.
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

// SetValues - метод установки значений конфига из флагов или из переменных окружения.
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
	// принимаем секретный ключ сервера для авторизации
	flag.BoolVar(&c.UseHTTPS, "tls", false, "use https")
	// принимаем секретный ключ сервера для авторизации
	flag.StringVar(&c.FileConfigPath, "c", "", "file config path")
	// доверенная подсеть для запросов статистики
	flag.StringVar(&c.TrustedSubnet, "t", "172.17.0.0/16", "trusted subnet")
	// парсим переданные серверу аргументы в зарегестрированные переменные
	flag.Parse()

	if c.FileConfigPath != "" {
		// Open config json file
		confJsonFile, err := os.Open(c.FileConfigPath)
		if err != nil {
			return fmt.Errorf("fail openning config json file: %w", err)
		}
		defer confJsonFile.Close()
		// read our opened jsonFile as a byte array.
		byteValue, _ := io.ReadAll(confJsonFile)
		json.Unmarshal(byteValue, &c)
	}

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
	if envUseHTTPS := os.Getenv("ENABLE_HTTPS"); envUseHTTPS == "true" {
		c.UseHTTPS = true
	}
	if envConfigPath := os.Getenv("CONFIG"); envConfigPath != "" {
		c.FileConfigPath = envConfigPath
	}
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		c.TrustedSubnet = envTrustedSubnet
	}
	return nil
}
