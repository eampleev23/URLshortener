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
	// Десериализуем запрос в структуру модели
	logger.Log.Info("decoding request")
	var req models.RequestAddShortURL
	log.Printf("r.Body:%s", r.Body)
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&req); err != nil {
		logger.Log.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// создаем новую пару ссылок
	shortURL, err := h.s.SetShortURL(req.LongURL)
	shortURL = h.c.GetValueByIndex("baseshorturl") + "/" + shortURL
	if err != nil {
		logger.Log.Info("cannot set shortURL:", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	// Заполняем модель ответа
	resp := models.ResponseAddShortURL{ShortURL: shortURL}

	// Сериализуем ответ сервера
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		logger.Log.Info("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	logger.Log.Debug("Sending HTTP 201 response")
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Accept", "application/json")
	w.WriteHeader(201)

}

func (h *Handlers) CreateShortLink(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {

		var longLink string
		if b, err := io.ReadAll(req.Body); err == nil {
			longLink = string(b)
		}
		// Генерируем и сразу сохраняем короткую ссылку для переданной длинной
		var shortLink string
		shortLink = ""
		for shortLink == "" {
			shortLink, _ = h.s.SetShortURL(longLink)
		}

		// Устанавливаем статус 201
		res.WriteHeader(201)

		// Устаннавливаем тип контента text/plain
		res.Header().Set("content-type", "text/plain")
		//shortLinkWithPrefix := "http://localhost" + h.c.GetValueByIndex("runaddr") + "/" + shortLink
		shortLinkWithPrefix := h.c.GetValueByIndex("baseshorturl") + "/" + shortLink
		res.Write([]byte(shortLinkWithPrefix))
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (h *Handlers) UseShortLink(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusBadRequest)
	} else {
		//res.Header().Add("Location", linksCouples[chi.URLParam(req, "id")])
		loc, err := h.s.GetLongLinkByShort(chi.URLParam(req, "id"))
		if err != nil {
			log.Print(err)
			res.WriteHeader(http.StatusBadRequest)
		}
		res.Header().Add("Location", loc)
		// добавляю для эксперимента
		res.WriteHeader(http.StatusTemporaryRedirect)

		// Если совпадений в бд нет, то ставим статус код бэд реквест
		if loc == "no match" {
			res.WriteHeader(http.StatusBadRequest)
		} else { // иначе это успех, есть совпадение, ставим 307 и в заголовок ответа локейшн отправляем длинную ссылку
			res.WriteHeader(http.StatusTemporaryRedirect)
		}

	}

}
