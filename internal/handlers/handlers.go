package handlers

import (
	"encoding/json"
	"github.com/eampleev23/URLshortener/internal/config"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/eampleev23/URLshortener/internal/models"
	"github.com/eampleev23/URLshortener/internal/store"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
)

type Handlers struct {
	s *store.Store
	c *config.Config
}

func NewHandlers(s *store.Store, c *config.Config) *Handlers {
	return &Handlers{
		s: s,
		c: c,
	}
}

func (h *Handlers) JSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(201)
		var req models.RequestAddShortURL
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			logger.Log.Info("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		shortURL, err := h.s.SetShortURL(req.LongURL)
		shortURL = h.c.GetValueByIndex("baseshorturl") + "/" + shortURL
		if err != nil {
			logger.Log.Info("cannot set shortURL:", zap.Error(err))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		resp := models.ResponseAddShortURL{ShortURL: shortURL}
		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			logger.Log.Info("error encoding response", zap.Error(err))
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
		var shortLink string
		shortLink = ""
		for shortLink == "" {
			shortLink, _ = h.s.SetShortURL(longLink)
		}

		// Устанавливаем статус 201
		w.WriteHeader(201)

		// Устаннавливаем тип контента text/plain
		w.Header().Set("content-type", "text/plain")
		//shortLinkWithPrefix := "http://localhost" + h.c.GetValueByIndex("runaddr") + "/" + shortLink
		shortLinkWithPrefix := h.c.GetValueByIndex("baseshorturl") + "/" + shortLink
		w.Write([]byte(shortLinkWithPrefix))
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
