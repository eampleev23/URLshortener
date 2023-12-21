package handlers

import (
	"encoding/json"
	"errors"
	"github.com/eampleev23/URLshortener/internal/models"
	"github.com/eampleev23/URLshortener/internal/store"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func (h *Handlers) JSONHandler(w http.ResponseWriter, r *http.Request) {
	// Создаем модель для парсинга в нее запроса
	var req models.RequestAddShortURL
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.l.ZL.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.s.SetShortURL(req.LongURL)
	if err != nil {
		h.l.ZL.Info("Ошибка создания новой ссылки: ", zap.Error(err))
		// Если такая ссылка уже есть в базе, возвращаем шорт для нее
		if errors.Is(err, store.ErrConflict) {
			// пытаемся получить ссылку для оригинального урл, который уже есть в базе
			shortURL, err = h.s.GetShortLinkByLong(r.Context(), req.LongURL)
			if err != nil {
				log.Printf("ошибка получения существующей короткой ссылки при конфликте%v", err)
			}
			shortLinkWithPrefix := h.c.BaseShortURL + "/" + shortURL
			resp := models.ResponseAddShortURL{ShortURL: shortLinkWithPrefix}
			enc := json.NewEncoder(w)
			w.WriteHeader(http.StatusConflict)
			if err := enc.Encode(resp); err != nil {
				h.l.ZL.Info("error encoding response", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}

	}
	// Здесь получается такой ссылки нет в базе, создаем новую
	shortLinkWithPrefix := h.c.BaseShortURL + "/" + shortURL
	resp := models.ResponseAddShortURL{ShortURL: shortLinkWithPrefix}
	enc := json.NewEncoder(w)
	w.WriteHeader(http.StatusCreated)
	if err := enc.Encode(resp); err != nil {
		h.l.ZL.Info("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	return
}
