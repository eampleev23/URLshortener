package handlers

import (
	"fmt"
	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/services"
	"github.com/eampleev23/URLshortener/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/http"
)

var keyUserIDCtx myauth.Key = myauth.KeyUserIDCtx

type Handlers struct {
	s          store.Store
	c          *config.Config
	l          *logger.ZapLog
	au         myauth.Authorizer
	serv       *services.Services
	deleteChan chan store.DeleteURLItem
}

func NewHandlers(s store.Store, c *config.Config, l *logger.ZapLog, au myauth.Authorizer, serv *services.Services) *Handlers {
	handlers := &Handlers{
		s:          s,
		c:          c,
		l:          l,
		au:         au,
		serv:       serv,
		deleteChan: make(chan store.DeleteURLItem, 1024), // установим каналу буфер в 1024 сообщения
	}
	// запустим горутину с фоновым удалением урлов
	go handlers.flushRequests()
	return handlers
}

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
