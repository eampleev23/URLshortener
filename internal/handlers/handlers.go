// Package handlers - эндпоинты сервиса.
package handlers

import (
	"fmt"
	"net/http"

	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/services"
	"github.com/eampleev23/URLshortener/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var keyUserIDCtx myauth.Key = myauth.KeyUserIDCtx

// Handlers - класс хэндлеров.
type Handlers struct {
	serv *services.Services
	s    store.Store
	c    *config.Config
	l    *logger.ZapLog
	au   myauth.Authorizer
}

// NewHandlers - конструктор.
func NewHandlers(
	s store.Store,
	c *config.Config,
	l *logger.ZapLog,
	au myauth.Authorizer,
	serv *services.Services) *Handlers {
	handlers := &Handlers{
		s:    s,
		c:    c,
		l:    l,
		au:   au,
		serv: serv,
	}
	return handlers
}

// GetUserID - метод хэндлеров для получения ид текущего пользователя.
func (h *Handlers) GetUserID(r *http.Request) (userID int, isAuth bool, err error) {
	userIDCtx, ok := r.Context().Value(keyUserIDCtx).(int)
	if !ok {
		return 0, false, fmt.Errorf("userIDCtx is not set: %w", err)
	}
	if userIDCtx != 0 {
		return userIDCtx, false, nil
	}
	cookie, err := r.Cookie("token")
	if err != nil {
		return 0, false, fmt.Errorf("token not set in cookie: %w", err)
	}
	userID, err = h.au.GetUserID(cookie.Value)
	if err != nil {
		return 0, false, fmt.Errorf("au.GetUserID error: %w", err)
	}
	return userID, true, nil
}
