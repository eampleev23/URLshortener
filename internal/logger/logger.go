package logger

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type ZapLog struct {
	ZL *zap.Logger
}

func NewZapLogger(level string) (*ZapLog, error) {
	lg := &ZapLog{ZL: zap.NewNop()}
	var err error
	lg, err = LoggerInitialize(level, lg)
	return lg, err
}

// LoggerInitialize инициализирует синглтон логера с необходимым уровнем логирования.
func LoggerInitialize(level string, zapObj *ZapLog) (*ZapLog, error) {
	// Преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	// Создаем новую конфигурацию логгера
	cfg := zap.NewProductionConfig()
	// Устанавливаем уровень
	cfg.Level = lvl
	// Создаем логгер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
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
	r.responseData.size += size // захватываем размер
	/*
				Если обернуть ошибку вот так:
				return size, fmt.Errorf("%w", err)
				то падает автотест 7 инкремента

		iteration7_test.go:162:
			11
			        	Error Trace:	/__w/URLshortener/URLshortener/iteration7_test.go:162
			12
			        	            				/__w/URLshortener/URLshortener/suite.go:91
			13
			        	Error:      	Received unexpected error:
			14
			        	            	unexpected EOF
			15
			        	Test:       	TestIteration7/TestJSONHandler/shorten
			16
			        	Messages:   	Ошибка при попытке сделать запрос для сокращения URL
			17
			    iteration7_test.go:180: Оригинальный запрос:
			18

			19
			        POST /api/shorten HTTP/1.1
			20
			        Host: localhost:8080
			21
			        Accept: application/json
			22
			        Content-Type: application/json
			23
			        User-Agent: go-resty/2.7.0 (https://github.com/go-resty/resty)


	*/
	return size, fmt.Errorf("%w", err)
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

		next.ServeHTTP(&lw, r)

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
