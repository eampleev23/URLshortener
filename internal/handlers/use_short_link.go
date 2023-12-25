package handlers

import (
	"go.uber.org/zap"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) UseShortLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		loc, err := h.s.GetLongLinkByShort(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			h.l.ZL.Info("error GetLongLinkByShort: ", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Header().Add("Location", loc)
		w.WriteHeader(http.StatusTemporaryRedirect)
		// Если совпадений в бд нет, то ставим статус код бэд реквест
		if loc == "no match" {
			w.WriteHeader(http.StatusBadRequest)
		} else { // иначе это успех, есть совпадение, ставим 307 и в заголовок ответа локейшн отправляем длинную ссылку
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}
