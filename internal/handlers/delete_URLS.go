package handlers

import (
	"context"
	"encoding/json"
	"github.com/eampleev23/URLshortener/internal/store"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

func (h *Handlers) DeleteURLS(w http.ResponseWriter, r *http.Request) {
	h.l.ZL.Debug("DeleteURLS start..")
	// Сначала преобразуем входящие данные в массив моделей.
	var req []string
	// Далее создаем декодер
	dec := json.NewDecoder(r.Body)
	// В случае некорректного запроса, возвращаем соответствующий заголовок
	if err := dec.Decode(&req); err != nil {
		h.l.ZL.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// В случае успешного парсинга в массив моделей, возвращаем статус 202 Accepted
	w.WriteHeader(http.StatusAccepted)
	// Далее передаем в модель данные для обработки.

	userID, err := h.GetUserID(r)
	if err != nil {
		h.l.ZL.Info("getting userId", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// здеь нам нужно пробежаться в цикле и напихать запросов в канал
	for _, v := range req {
		// отправим сообщение в очередь на удаление
		h.deleteChan <- store.DeleteURLItem{
			ShortURL:   v,
			DeleteFlag: true,
			OwnerID:    userID,
		}
	}
}

func (h *Handlers) flushRequests() {
	// будем сохранять сообщения, накопленные за последние 10 секунд
	ticker := time.NewTicker(5 * time.Second)
	var deleteItems []store.DeleteURLItem

	for {
		select {
		case deleteReq := <-h.deleteChan:
			// добавим запрос на удаление в слайс для последующего удаления
			deleteItems = append(deleteItems, deleteReq)
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(deleteItems) == 0 {
				log.Println("delete items len = ", len(deleteItems))
				continue
			}
			log.Println("delete items len = ", len(deleteItems))
			// выполним все пришедшие за 10 секунд запросы за один раз батчингом
			err := h.s.DeleteURLS(context.TODO(), deleteItems)
			if err != nil {
				// не будем стирать сообщения, попробуем отправить их чуть позже
				continue
			}
			// сотрём успешно отосланные сообщения
			deleteItems = nil
		}
	}
}
