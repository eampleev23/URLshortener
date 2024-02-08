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

type Services struct {
	DeleteChan chan store.DeleteURLItem
	s          store.Store
	c          *config.Config
	l          *logger.ZapLog
	au         myauth.Authorizer
}

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

	//for i := 0; i < len(result); i++ {
	//	result[i].ShortURL, err = url.JoinPath(serv.c.BaseShortURL, result[i].ShortURL)
	//	if err != nil {
	//		return nil, fmt.Errorf("url.JoinPath error %w", err)
	//	}
	//}
	return result, nil
}

func (serv *Services) FlushRequests() {
	// будем сохранять сообщения, накопленные за последние 10 секунд
	ticker := time.NewTicker(5 * time.Second) //nolint:gomnd //no magik
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
			err := serv.s.DeleteURLS(context.TODO(), deleteItems)
			if err != nil {
				// не будем стирать сообщения, попробуем отправить их чуть позже
				continue
			}
			// сотрём успешно отосланные сообщения
			deleteItems = nil
		}
	}
}
