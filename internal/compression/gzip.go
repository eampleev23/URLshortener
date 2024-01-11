package compression

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/eampleev23/URLshortener/internal/logger"
	"go.uber.org/zap"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	result, err := c.zw.Write(p)
	if err != nil {
		return result, fmt.Errorf("failed to write by compressWriter: %w", err)
	}
	return result, nil
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	err := c.zw.Close()
	if err != nil {
		return fmt.Errorf("failed to close gzip.writer: %w", err)
	}
	return nil
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a newCompressReader: %w", err)
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	result, err := c.zr.Read(p)
	return result, err //nolint:wrapcheck // устаревший способ обработки ошибки внутри Read
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("failed to close a *compressReader: %w", err)
	}
	return nil
}

var keyLogger logger.Key = logger.KeyLoggerCtx

func GzipMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Получаем логгер из контекста запроса
		logger, ok := r.Context().Value(keyLogger).(*logger.ZapLog)
		if !ok {
			log.Printf("Error getting logger")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		// и на соответствие контэнт тайпа
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		contentType := r.Header.Get("Content-Type")
		supportContentType := strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html")

		if supportsGzip && supportContentType {
			// оборачиваем оригинальный http.ResponseWriter новым, с поддержкой сжатия
			cw := newCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware.
			defer func() {
				err := cw.Close()
				if err != nil {
					logger.ZL.Info("middleware failed by compresswriter", zap.Error(err))
				}
			}()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer func() {
				err := cr.Close()
				if err != nil {
					logger.ZL.Info("middleware failed by compressreader", zap.Error(err))
				}
			}()
		}

		// передаём управление хендлеру
		next.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(fn)
}
