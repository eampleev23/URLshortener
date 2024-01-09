package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eampleev23/URLshortener/internal/models"
	"github.com/eampleev23/URLshortener/internal/store"
	"go.uber.org/zap"
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

	shortURL, err := h.s.SetShortURL(r.Context(), req.LongURL, 12)
	if err != nil {
		h.l.ZL.Info("Ошибка создания новой ссылки: ", zap.Error(err))
		// Если такая ссылка уже есть в базе, возвращаем шорт для нее
		if errors.Is(err, store.ErrConflict) {
			// пытаемся получить ссылку для оригинального урл, который уже есть в базе
			shortURL, err = h.s.GetShortURLByOriginal(r.Context(), req.LongURL)
			if err != nil {
				h.l.ZL.Error("ошибка получения существующей короткой ссылки при конфликте", zap.Error(err))
			}
			shortLinkWithPrefix := h.c.BaseShortURL + "/" + shortURL
			resp := models.ResponseAddShortURL{ShortURL: shortLinkWithPrefix}
			enc := json.NewEncoder(w)
			w.Header().Set("content-type", "application/json")
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
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := enc.Encode(resp); err != nil {
		h.l.ZL.Info("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
