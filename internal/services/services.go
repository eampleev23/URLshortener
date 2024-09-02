package services

import (
	"context"
	"fmt"
	"net/url"
	"time"

	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/store"
)

// Services - класс сервисов.
type Services struct {
	DeleteChan chan store.DeleteURLItem
	s          store.Store
	c          *config.Config
	l          *logger.ZapLog
	au         myauth.Authorizer
}

// NewServices - конструтор класса сервисов.
func NewServices(s store.Store, c *config.Config, l *logger.ZapLog, au myauth.Authorizer) *Services {
	services := &Services{
		DeleteChan: make(chan store.DeleteURLItem, 1024), //nolint:gomnd //установим каналу буфер в 1024 сообщения
		s:          s,
		c:          c,
		l:          l,
		au:         au,
	}
	// запустим горутину с фоновым удалением урлов
	go services.FlushRequests()
	return services
}

// GetURLsByOwnerID - ну вы поняли.
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

// FlushRequests - выполнить накопленные запросы на удаление.
func (serv *Services) FlushRequests() {
	// будем сохранять сообщения, накопленные за последние 10 секунд
	ticker := time.NewTicker(serv.c.TimeDeleteURLs)
	var deleteItems []store.DeleteURLItem

	for {
		select {
		case deleteReq := <-serv.DeleteChan:
			// добавим запрос на удаление в слайс для последующего удаления
			deleteItems = append(deleteItems, deleteReq)
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(deleteItems) == 0 {
				continue
			}
			// выполним все пришедшие за 10 секунд запросы за один раз батчингом
			ctx, cancel := context.WithTimeout(context.Background(), serv.c.TLimitQuery)
			err := serv.s.DeleteURLS(ctx, deleteItems)
			if err != nil {
				// не будем стирать сообщения, попробуем отправить их чуть позже
				continue
			}
			// сотрём успешно отосланные сообщения
			deleteItems = nil
			cancel()
		}
	}
}
