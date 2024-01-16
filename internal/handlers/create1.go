package handlers

import (
	"errors"
	"github.com/eampleev23/URLshortener/internal/store"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"net/url"
)

func (h *Handlers) CreateShortLink1(w http.ResponseWriter, r *http.Request) {
	log.Println("here")
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
		userIDCtx, ok := r.Context().Value(keyUserIDCtx).(int)
		if !ok {
			h.l.ZL.Info("Error getting if set new cookie user id ctx")
			return
		}
		h.l.ZL.Info("Значение userIDCtx!!!", zap.Int("userIDCtx", userIDCtx))
		for shortLink == "" {
			if userIDCtx != 0 {
				shortLink, err = h.s.SetShortURL(r.Context(), longLink, userIDCtx)
			} else {
				cookie, _ := r.Cookie("token")
				userID, _ := h.au.GetUserID(cookie.Value)
				shortLink, err = h.s.SetShortURL(r.Context(), longLink, userID)
			}
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
					shortLinkWithPrefix, err := url.JoinPath(h.c.BaseShortURL, shortLink)
					if err != nil {
						h.l.ZL.Info("error url.JoinPath", zap.String("shortlink", shortLink))
					}
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
			h.l.ZL.Info("внизу цикла оказались")
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
