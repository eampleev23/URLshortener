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

		// Проверяем была ли установлена новая кука при текущем запросе
		setNewCookie, ok := r.Context().Value(keyAuth).(bool)
		if !ok {
			h.l.ZL.Info("Error getting if set new cookie")
			return
		}
		var aInt int
		aInt = 1
		if !setNewCookie {
			h.l.ZL.Info("не устанавливалась новая, есть старая")
			cookie, _ := r.Cookie("token")
			bInt, _ := h.au.GetUserID(cookie.Value)
			aInt = bInt
		}

		for shortLink == "" {
			shortLink, err = h.s.SetShortURL(r.Context(), longLink, aInt)
			if err != nil {
				// здесь делаем проверку на конфликт
				if errors.Is(err, store.ErrConflict) {
					// пытаемся получить ссылку для оригинального урл, который уже есть в базе
					w.WriteHeader(http.StatusConflict)
					w.Header().Set("content-type", "text/plain")
					shortLink, err = h.s.GetShortURLByOriginal(r.Context(), longLink)
					if err != nil {
						h.l.ZL.Info("error GetShortURLByOriginal", zap.Error(err))
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
