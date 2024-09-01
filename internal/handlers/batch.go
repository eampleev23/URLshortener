package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/eampleev23/URLshortener/internal/models"
	"go.uber.org/zap"
)

/*
Пример запроса:
[
    {
        "correlation_id": "test",
        "original_url": "testqwedqwe.com"
    }
]

// сейчас работает не через батчинг, но батчинг реализовывал в другом инкременте (возможно в проекте, при необходимости
найду) Также работает только если передавать один урл, с группой сбоит.. можно разобраться при необходимости
*/

func (h *Handlers) JSONHandlerBatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json") //nolint:goconst // не понятно зачем константа
	var req []models.BatchItemReq

	// Декодер работает потоково, кажется это правильнее + короче, чем анмаршал.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.l.ZL.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Получается реквест мы получили корректный, теперь начинаем готовить ответ
	// Перебираем каждый элемент в запросе
	res := make([]models.BatchItemRes, 0)
	for i := range req {
		defaultValueOwnerID := 12
		log.Println("req[i].OriginalURL=", req[i].OriginalURL)
		shortURL, err := h.s.SetShortURL(r.Context(), req[i].OriginalURL, defaultValueOwnerID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resURL, err := url.JoinPath(h.c.BaseShortURL, shortURL)
		if err != nil {
			h.l.ZL.Info("error url.joinpath", zap.Error(err))
		}
		res = append(res, models.BatchItemRes{
			CorrelationID: req[i].CorrelationID,
			ShortURL:      resURL,
		})
	}
	enc := json.NewEncoder(w)
	w.WriteHeader(http.StatusCreated)
	if err := enc.Encode(res); err != nil {
		h.l.ZL.Info("error encoding response in batch handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
