package handlers

import (
	"errors"
	"github.com/eampleev23/URLshortener/internal/store"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

// CreateShortLink получает в пост запросе урл, который необходимо сократить и возвращает сокращенный в виде text/plain.
func (h *Handlers) CreateShortLink1(w http.ResponseWriter, r *http.Request) {

	// Заводим переменную, в которую запишем длинную ссылку из запроса
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.l.ZL.Info("Error of reading the request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Преобразуем байты в строку и заносим значение в соответствующую переменную
	originalURL := string(reqBody)
	h.l.ZL.Info("Got originalURL", zap.String("originalURL", originalURL))

	// заводим переменную, в которой будем хранить id пользователя
	//var userID int
	userID, err := h.GetUserID(r)
	if err != nil {
		h.l.ZL.Info("Error getting userID", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.l.ZL.Info("Got userID", zap.Int("userID", userID))

	// изначально заносим пустую строку в shortURL
	shortURL := ""
	// это номер попытки в случае коллизии
	var numberOfAttempts int8 = 0
	// это количество попыток в члучае коллизии
	var limitTry int8 = 2

	// запускаем цикл, который будет отрабатывать до тех пор пока в shortURL не занесется не пусая строка
	for shortURL == "" {
		// заносим в стор новое значение
		// видимо надо отрефакторить этот метод стора
		shortURL, err = h.s.SetShortURL(r.Context(), originalURL, userID)

		if err != nil {
			// здесь делаем проверку на конфликт
			if errors.Is(err, store.ErrConflict) {
				// пытаемся получить ссылку для оригинального урл, который уже есть в базе
				w.WriteHeader(http.StatusConflict)
				w.Header().Set("content-type", "text/plain")
				shortURL, err = h.s.GetShortURLByOriginal(r.Context(), originalURL)
				if err != nil {
					h.l.ZL.Info("error GetShortURLByOriginal", zap.Error(err))
				}
				shortLinkWithPrefix, err := url.JoinPath(h.c.BaseShortURL, shortURL)
				if err != nil {
					h.l.ZL.Info("error url.JoinPath", zap.String("shortlink", shortURL))
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

	//// Устанавливаем статус 201
	//w.WriteHeader(http.StatusCreated)
	//// Устаннавливаем тип контента text/plain
	//w.Header().Set("content-type", "text/plain")
	//shortLinkWithPrefix, err := url.JoinPath(h.c.BaseShortURL, shortLink)
	//if err != nil {
	//	h.l.ZL.Info("error url.joinpath in handler", zap.Error(err))
	//}
	//_, err = w.Write([]byte(shortLinkWithPrefix))
	//if err != nil {
	//	h.l.ZL.Info("Ошибка в хэндлере CreateShortLink при вызове w.Write", zap.Error(err))
	//	w.WriteHeader(http.StatusInternalServerError)
	//}
}
