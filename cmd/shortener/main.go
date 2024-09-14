package main

import (
	"context"
	"fmt"
	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/compression"
	"github.com/eampleev23/URLshortener/internal/services"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/handlers"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

/*
Iter 23.
*/

func run() error {

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

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

	// конструируем сервер
	server := &http.Server{
		Addr:    c.RanAddr,
		Handler: r,
	}

	// Если используем HTTPS.
	if c.UseHTTPS {
		// конструируем менеджер TLS-сертификатов
		manager := &autocert.Manager{
			// директория для хранения сертификатов
			Cache: autocert.DirCache("cache-dir"),
			// функция, принимающая Terms of Service издателя сертификатов
			Prompt: autocert.AcceptTOS,
			// перечень доменов, для которых будут поддерживаться сертификаты
			HostPolicy: autocert.HostWhitelist("shortener.ru", "www.shortener.ru"),
		}
		server.TLSConfig = manager.TLSConfig()
	}

	// Реализуем gracefully shutdown ниже.
	// Получаем контекст, в котором планируем слушать 3 сигнала и также получаем стоп-функцию.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	// Отложенно вызывем стоп-функцию чтобы не забыть.
	defer stop()
	// Запускаем сервер в горутине.
	go func(srv *http.Server) error {
		err := srv.ListenAndServe()
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		return nil
	}(server)
	// Ждем завершающий сигнал.
	<-ctx.Done()
	// Возвращаем возможные ошибки.
	if ctx.Err() != nil {
		return fmt.Errorf("gracefully shotdowned: %w", ctx.Err())
	}
	return nil
}
