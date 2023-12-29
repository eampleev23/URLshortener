package logger

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type ZapLog struct {
	ZL *zap.Logger
}

func NewZapLogger(level string) (*ZapLog, error) {
	lg := &ZapLog{ZL: zap.NewNop()}
	var err error
	lg, err = Initialize(level, lg)
	return lg, err
}

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string, zapObj *ZapLog) (*ZapLog, error) {
	// Преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to ParseAtomicLevel by LoggerInitialaze: %w", err)
	}
	// Создаем новую конфигурацию логгера
	cfg := zap.NewProductionConfig()
	// Устанавливаем уровень
	cfg.Level = lvl
	// Создаем логгер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build by config in init logger: %w", err)
	}
	zapObj.ZL = zl
	return zapObj, nil
}

type (
	// Берём структуру для хранения сведений об ответе.
	responseData struct {
		status int
		size   int
	}

	// Добавляем реализацию http.ResponseWriter.
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("failed to write response by original responsewriter: %w", err)
	}
	r.responseData.size += size // захватываем размер
	return size, nil
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func (zl *ZapLog) RequestLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		ctx := context.WithValue(r.Context(), keyLoggerID, zl)
		next.ServeHTTP(&lw, r.WithContext(ctx))
		duration := time.Since(start)
		zl.ZL.Info("got incoming HTTP request",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("content-type", r.Header.Get("content-type")),
			zap.String("duration", shortDur(duration)),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)
	}
	return http.HandlerFunc(fn)
}

// Вспомогательная функция для перевода time.Duration в строку при выводе в лог.
func shortDur(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}
