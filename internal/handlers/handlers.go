package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/models"
	"github.com/eampleev23/URLshortener/internal/store"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handlers struct {
	s *store.Store
	c *config.Config
	l *logger.ZapLog
}

func NewHandlers(s *store.Store, c *config.Config, l *logger.ZapLog) *Handlers {
	return &Handlers{
		s: s,
		c: c,
		l: l,
	}
}

func (h *Handlers) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	pingDB := false
	if pingDB {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(""))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func (h *Handlers) JSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		var req models.RequestAddShortURL
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			h.l.ZL.Info("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		shortURL := ""
		var err error
		var numberOfAttempts int8 = 0
		var limitTry int8 = 10

		for shortURL == "" {
			shortURL, err = h.s.SetShortURL(req.LongURL)
			if err != nil {
				h.l.ZL.Info("Произошла коллизия: ", zap.Error(err))
				numberOfAttempts++
				if numberOfAttempts > limitTry {
					// Попробовали, хватит
					w.WriteHeader(http.StatusExpectationFailed)
					return
				}
			}
		}
		shortURL = h.c.BaseShortURL + "/" + shortURL
		if err != nil {
			h.l.ZL.Info("cannot set shortURL:", zap.Error(err))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		resp := models.ResponseAddShortURL{ShortURL: shortURL}
		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			h.l.ZL.Info("error encoding response", zap.Error(err))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

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
				h.l.ZL.Info("Произошла коллизия: ", zap.Error(err))
				numberOfAttempts++
				if numberOfAttempts > limitTry {
					// Попробовали, хватит
					w.WriteHeader(http.StatusExpectationFailed)
					return
				}
			}
		}

		// Устанавливаем статус 201
		w.WriteHeader(http.StatusCreated)
		// Устаннавливаем тип контента text/plain
		w.Header().Set("content-type", "text/plain")
		shortLinkWithPrefix := h.c.BaseShortURL + "/" + shortLink
		_, err = w.Write([]byte(shortLinkWithPrefix))
		if err != nil {
			h.l.ZL.Info("Ошибка в хэндлере CreateShortLink при вызове w.Write", zap.Error(err))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (h *Handlers) UseShortLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		loc, err := h.s.GetLongLinkByShort(chi.URLParam(r, "id"))
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Header().Add("Location", loc)
		// добавляю для эксперимента
		w.WriteHeader(http.StatusTemporaryRedirect)

		// Если совпадений в бд нет, то ставим статус код бэд реквест
		if loc == "no match" {
			w.WriteHeader(http.StatusBadRequest)
		} else { // иначе это успех, есть совпадение, ставим 307 и в заголовок ответа локейшн отправляем длинную ссылку
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}
