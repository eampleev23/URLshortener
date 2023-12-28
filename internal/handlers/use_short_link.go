package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) UseShortLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		loc, err := h.s.GetOriginalURLByShort(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
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
