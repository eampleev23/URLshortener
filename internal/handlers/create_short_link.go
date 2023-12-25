package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/eampleev23/URLshortener/internal/store"
	"go.uber.org/zap"
)

func (h *Handlers) CreateShortLink(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var longLink string
		if b, err := io.ReadAll(r.Body); err == nil {
			longLink = string(b)
		}
		// Генерируем и сразу сохраняем короткую ссылку для переданной длинной
		shortLink := ""
		var err error
		var numberOfAttempts int8 = 0
		var limitTry int8 = 10
		for shortLink == "" {
			shortLink, err = h.s.SetShortURL(longLink)
			if err != nil {
				// здесь делаем проверку на конфликт
				if errors.Is(err, store.ErrConflict) {
					// пытаемся получить ссылку для оригинального урл, который уже есть в базе
					w.WriteHeader(http.StatusConflict)
					w.Header().Set("content-type", "text/plain")
					shortLink, err = h.s.GetShortLinkByLong(r.Context(), longLink)
					if err != nil {
						h.l.ZL.Info("error GetShortLinkByLong", zap.Error(err))
					}
					shortLinkWithPrefix := h.c.BaseShortURL + "/" + shortLink
					_, err = w.Write([]byte(shortLinkWithPrefix))
					if err != nil {
						h.l.ZL.Info("Ошибка в хэндлере CreateShortLink при вызове w.Write", zap.Error(err))
						w.WriteHeader(http.StatusInternalServerError)
					}
					return
				}
				numberOfAttempts++
				if numberOfAttempts > limitTry {
					// Попробовали, хватит
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		// Устанавливаем статус 201
		w.WriteHeader(http.StatusCreated)
		// Устаннавливаем тип контента text/plain
		w.Header().Set("content-type", "text/plain")
		shortLinkWithPrefix, err := url.JoinPath(h.c.BaseShortURL, shortLink)
		if err != nil {
			h.l.ZL.Info("error url.joinpath in handler", zap.Error(err))
		}
		_, err = w.Write([]byte(shortLinkWithPrefix))
		if err != nil {
			h.l.ZL.Info("Ошибка в хэндлере CreateShortLink при вызове w.Write", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
