package services

import (
	"context"
	"fmt"
	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
	"net/url"
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
	for _, v := range result {
		v.ShortURL, err = url.JoinPath(serv.c.BaseShortURL, v.ShortURL)
		if err != nil {
			return nil, fmt.Errorf("url.JoinPath error %w", err)
		}
	}
	return result, nil
}
