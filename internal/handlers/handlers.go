package handlers

import (
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Handlers struct {
	s *store.Storage
	c *config.Config
	l *logger.ZapLog
}

func NewHandlers(s *store.Storage, c *config.Config, l *logger.ZapLog) *Handlers {
	return &Handlers{
		s: s,
		c: c,
		l: l,
	}
}
