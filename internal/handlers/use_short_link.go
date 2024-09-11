package handlers

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

// UseShortLink - хэндлер для использования короткой ссылки.
func (h *Handlers) UseShortLink(w http.ResponseWriter, r *http.Request) {
	linksCouple, err := h.s.GetLinksCoupleByShortURL(r.Context(), chi.URLParam(r, "id"))
	if linksCouple.DeletedFlag {
		w.WriteHeader(http.StatusGone)
		return
	}
	if err != nil {
		h.l.ZL.Info("GetLinksCoupleByShortURL error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", linksCouple.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
