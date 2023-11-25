package handlers

import (
	"github.com/eampleev23/URLshortener/internal/store"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
)

type Handlers struct {
	s *store.Store
}

func NewHandlers(s *store.Store) *Handlers {
	return &Handlers{
		s: s,
	}
}

func (h *Handlers) CreateShortLink(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {

		var longLink string
		if b, err := io.ReadAll(req.Body); err == nil {
			longLink = string(b)
		}
		// Генерируем и сразу сохраняем короткую ссылку для переданной длинной
		shortLink, err := h.s.SetShortURL(longLink)
		if err != nil {
			log.Print(err)
		}
		// Устанавливаем статус 201
		res.WriteHeader(201)

		// Устаннавливаем тип контента text/plain
		res.Header().Set("content-type", "text/plain")
		shortLinkWithPrefix := "localhost:8080" + "/" + shortLink
		res.Write([]byte(shortLinkWithPrefix))
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (h *Handlers) UseShortLink(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {
		str := chi.URLParam(req, "id")
		loc, err := h.s.GetLongLinkByShort(str)
		if err != nil {
			log.Print(err)
			return
		}
		res.Header().Add("Location", loc)
		// добавляю для эксперимента
		res.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}

}
