package handlers

import (
	"encoding/json"
	"github.com/eampleev23/URLshortener/internal/models"
	"go.uber.org/zap"
	"net/http"
)

func (h *Handlers) GetGlobalStats(w http.ResponseWriter, r *http.Request) {
	h.l.ZL.Info("GetGlobalStats called")
	resp := models.ResponseGlobalStats{
		URLs:  23,
		Users: 27,
	}
	enc := json.NewEncoder(w)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := enc.Encode(resp)
	if err != nil {
		h.l.ZL.Info("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
