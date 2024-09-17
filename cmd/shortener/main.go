package main

import (
	"context"
	"fmt"
	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/compression"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/handlers"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/services"
	"github.com/eampleev23/URLshortener/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

const (
	domainName string = "shortener.ru"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

/*
Iter24
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

	if c.UseHTTPS {
		// конструируем менеджер TLS-сертификатов
		manager := &autocert.Manager{
			// директория для хранения сертификатов
			Cache: autocert.DirCache("cache-dir"),
			// функция, принимающая Terms of Service издателя сертификатов
			Prompt: autocert.AcceptTOS,
			// перечень доменов, для которых будут поддерживаться сертификаты
			HostPolicy: autocert.HostWhitelist(domainName, "www."+domainName),
		}
		server.TLSConfig = manager.TLSConfig()
		err = server.ListenAndServeTLS("", "")
		if err != nil {
			return fmt.Errorf("ошибка ListenAndServe: %w", err)
		}
		return nil
	}

	err = gracefullyShutdown(server, myLog, c, s)
	if err != nil {
		return fmt.Errorf("gracefully shutdown: %w", err)
	}
	myLog.ZL.Info("Server Shutdown gracefully")
	return nil
}

func gracefullyShutdown(
	server *http.Server,
	myLog *logger.ZapLog,
	c *config.Config,
	s store.Store) error {
	// Заводим канал для получения сигнала о gracefull shotdown сервиса
	allConnsClosed := make(chan struct{})

	// Заводим канал для перенаправления прерываний
	// поскольку нужно отловить всего одно прерывание,
	// ёмкости 1 для канала будет достаточно
	sigint := make(chan os.Signal, 1)

	// Регистрируем перенаправление прерываний
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// запускаем горутину обработки пойманных прерываний
	go func() {
		// читаем из канала прерываний
		// поскольку нужно прочитать только одно прерывание,
		// можно обойтись без цикла
		<-sigint
		// получили сигнал os.Interrupt, запускаем процедуру graceful shutdown
		if err := server.Shutdown(context.Background()); err != nil {
			// Ошибки закрытия listener.
			myLog.ZL.Error("HTTP server Shutdown", zap.Error(err))
		}
		// Сообщаем основному потоку, что все сетевые соединения обработаны и закрыты.
		close(allConnsClosed)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		// Ошибки старта или остановки Listener.
		//myLog.ZL.Error("HTTP server ListenAndServe", zap.Error(err))
		return fmt.Errorf("HTTP server ListenAndServe: %w", err)
	}
	// Ждём завершения процедуры graceful shutdown.
	<-allConnsClosed
	// получили оповещение о завершении
	// здесь можно освобождать ресурсы перед выходом,
	// например закрыть соединение с базой данных,
	// закрыть открытые файлы

	// Закрываем соединение с бд (было выше через defer, переделал по прерыванию)
	if len(c.DBDSN) != 0 {
		if err := s.Close(); err != nil {
			myLog.ZL.Info("store failed to properly close the DB connection")
		}
	}
	return nil
}
