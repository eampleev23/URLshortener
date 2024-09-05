package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/eampleev23/URLshortener/internal/compression"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/eampleev23/URLshortener/internal/services"

	myauth "github.com/eampleev23/URLshortener/internal/auth"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/handlers"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

/*
Iter18 - Добавьте к основным экспортированным методам и переменным (хендлерам, публичным структурам и интерфейсам)
в вашем проекте документацию в формате godoc.
Добавьте примеры работы с эндпоинтами практического трека в формате example_test.go.
Покрытие вашего кода тестами на данный момент должно быть не менее 40%. Тест.
*/

func run() error {
	c, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize a new config: %w", err)
	}

	myLog, err := logger.NewZapLogger(c.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize a new logger: %w", err)
	}

	au, err := myauth.Initialize(c.SecretKey, c.TokenEXP, myLog)
	if err != nil {
		return fmt.Errorf("failed to initialize a new authorizer: %w", err)
	}

	s, err := store.NewStorage(c, myLog)
	if err != nil {
		return fmt.Errorf("failed to initialize a new store: %w", err)
	}

	if len(c.DBDSN) != 0 {
		// Отложенно закрываем соединение с бд.
		defer func() {
			if err := s.Close(); err != nil {
				myLog.ZL.Info("store failed to properly close the DB connection")
			}
		}()
	}

	serv := services.NewServices(s, c, myLog, *au)
	h := handlers.NewHandlers(s, c, myLog, *au, serv)

	myLog.ZL.Info("Running server", zap.String("address", c.RanAddr))
	r := chi.NewRouter()
	r.Use(myLog.RequestLogger)
	r.Use(compression.GzipMiddleware)
	r.Use(au.Auth)
	r.Mount("/debug", middleware.Profiler())
	r.Post("/", h.CreateShortURL)
	r.Get("/ping", h.PingDBHandler)
	r.Get("/{id}", h.UseShortLink)
	r.Post("/api/shorten", h.JSONHandler)
	r.Post("/api/shorten/batch", h.JSONHandlerBatch)
	r.Get("/api/user/urls", h.GetURLsByUserID)
	r.Delete("/api/user/urls", h.DeleteURLS)
	err = http.ListenAndServe(c.RanAddr, r)
	if err != nil {
		return fmt.Errorf("ошибка ListenAndServe: %w", err)
	}
	return nil
}
