package main

import (
	"flag"
	"os"
)

type AppConfig struct {
	flagRunAddr  string
	flagLogLevel string
	baseShortURL string
}

// неэкспортированная переменная flagRunAddr содержит адрес и порт для запуска сервера
var flagRunAddr string  // localhost:8888
var flagLogLevel string // localhost:8888

// неэкспортированная переменная baseShortURL содержит базовый адрес короткой ссылки
var baseShortURL string // localhost:8888

//Добавьте возможность конфигурировать сервис с помощью аргументов командной строки.
//Создайте конфигурацию или переменные для запуска со следующими флагами:
//Флаг -a отвечает за адрес запуска HTTP-сервера (значение может быть таким: localhost:8888).
//Флаг -b отвечает за базовый адрес результирующего сокращённого URL (значение: адрес сервера перед коротким URL,
//например http://localhost:8000/qsd54gFg).
//Совет: создайте отдельный пакет config, где будет храниться структура с вашей конфигурацией и функция,
//которая будет инициализировать поля этой структуры. По мере усложнения конфигурации вы сможете добавлять
//необходимые поля в вашу структуру и инициализировать их.

// parseFlags обраатывает аргументы командной строки и сохраняет их значения в соответствующих переменных
func getAppConfig() AppConfig {

	// регистрируем переменную flagRunAddr как аргумент -a со значением по умолчанию :8080
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "logger level")

	// регистрируем переменную baseShortUrl как аргумент -b со значением по умолчанию http://localhost:8080
	flag.StringVar(&baseShortURL, "b", "http://localhost:8080", "base prefix for the shortURL")
	// парсим переданные серверу аргументы в зарегестрированные переменные
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		baseShortURL = envBaseURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}

	appConfig := AppConfig{
		flagRunAddr:  flagRunAddr,
		baseShortURL: baseShortURL,
		flagLogLevel: flagLogLevel,
	}
	return appConfig

}
