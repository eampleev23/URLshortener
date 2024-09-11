package main

import (
	"fmt"
	"golang.org/x/crypto/acme/autocert"
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
Iter21
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
		// конструируем сервер с поддержкой TLS
		server := &http.Server{
			Addr:    c.RanAddr,
			Handler: r,
			// для TLS-конфигурации используем менеджер сертификатов
			TLSConfig: manager.TLSConfig(),
		}
		err = server.ListenAndServeTLS("", "")
		if err != nil {
			return fmt.Errorf("ошибка ListenAndServe: %w", err)
		}
		return nil
	}
	err = http.ListenAndServe(c.RanAddr, r)
	if err != nil {
		return fmt.Errorf("ошибка ListenAndServe: %w", err)
	}
	return nil
}
