package handlers

import (
	"encoding/json"
	"github.com/eampleev23/URLshortener/internal/store"
	"go.uber.org/zap"
	"net/http"
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

	userID, _, err := h.GetUserID(r)
	if err != nil {
		h.l.ZL.Info("getting userId", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// здеь нам нужно пробежаться в цикле и напихать запросов в канал
	for _, v := range req {
		// отправим сообщение в очередь на удаление
		h.serv.DeleteChan <- store.DeleteURLItem{
			ShortURL:   v,
			DeleteFlag: true,
			OwnerID:    userID,
		}
	}
}
