package handlers

import (
	"net/http"

	"go.uber.org/zap"
)

func (h *Handlers) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	err := h.s.PingDB(r.Context(), h.c.TLimitQuery)
	if err != nil {
		h.l.ZL.Info("not ping", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
