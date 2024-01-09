package handlers

import (
	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Handlers struct {
	s  store.Store
	c  *config.Config
	l  *logger.ZapLog
	au myauth.Authorizer
}

func NewHandlers(s store.Store, c *config.Config, l *logger.ZapLog, au myauth.Authorizer) *Handlers {
	return &Handlers{
		s:  s,
		c:  c,
		l:  l,
		au: au,
	}
}
