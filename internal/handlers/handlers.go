package handlers

import (
	"fmt"
	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/http"
)

var keyUserIDCtx myauth.Key = myauth.KeyUserIDCtx

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

func (h *Handlers) GetUserID(r *http.Request) (userID int, err error) {
	userIDCtx, ok := r.Context().Value(keyUserIDCtx).(int)
	if !ok {
		return 0, fmt.Errorf("userIDCtx is not set: %w", err)
	}
	if userIDCtx != 0 {
		return userIDCtx, nil
	}
	cookie, err := r.Cookie("token")
	if err != nil {
		return 0, fmt.Errorf("token not set in cookie: %w", err)
	}
	userID, err = h.au.GetUserID(cookie.Value)
	if err != nil {
		return 0, fmt.Errorf("au.GetUserID error: %w", err)
	}
	return userID, nil
}
