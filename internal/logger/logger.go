package logger

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

// Log будет доступен всему коду как синглтон.
// Никакой код навыка, кроме функции InitLogger, не должен модифицировать эту переменную.
// По умолчанию установлен no-op-логер, который не выводит никаких сообщений.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	// Преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// Создаем новую конфигурацию логгера
	cfg := zap.NewProductionConfig()
	// Устанавливаем уровень
	cfg.Level = lvl
	// Создаем логгер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("Запустился мидлвеар")
		Log.Debug("got incoming HTTP request",
			zap.String("method", request.Method),
			zap.String("path", request.URL.Path),
		)
		fmt.Println("Прошли Log.Debug")
	})
	return next
}
