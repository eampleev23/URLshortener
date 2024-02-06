package services

import (
	"context"
	"fmt"
	"net/url"

	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
)

type Services struct {
	s  store.Store
	c  *config.Config
	l  *logger.ZapLog
	au myauth.Authorizer
}

func NewServices(s store.Store, c *config.Config, l *logger.ZapLog, au myauth.Authorizer) *Services {
	return &Services{
		s:  s,
		c:  c,
		l:  l,
		au: au,
	}
}

func (serv *Services) GetURLsByOwnerID(ctx context.Context, userID int) ([]store.LinksCouple, error) {
	result, err := serv.s.GetURLsByOwnerID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetURLsByOwnerID error %w", err)
	}
	for i := 0; i < len(result); i++ {
		result[i].ShortURL, err = url.JoinPath(serv.c.BaseShortURL, result[i].ShortURL)
		if err != nil {
			return nil, fmt.Errorf("url.JoinPath error %w", err)
		}
	}
	return result, nil
}
