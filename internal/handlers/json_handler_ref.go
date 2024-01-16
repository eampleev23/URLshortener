package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/models"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func (h *Handlers) JSONHandler1(w http.ResponseWriter, r *http.Request) {
	// Получаем id пользователя.
	userID, err := h.GetUserID(r)
	if err != nil {
		h.l.ZL.Info("Error getting userID", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.l.ZL.Info("Got userID", zap.Int("userID", userID))

	reqModel, err := getModelRequest(r)
	if err != nil {
		h.l.ZL.Info("getModelRequest error", zap.Error(err))
	}
	log.Println(reqModel)
}

func getModelRequest(r *http.Request) (reqModel models.RequestAddShortURL, err error) {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reqModel); err != nil {
		return models.RequestAddShortURL{}, fmt.Errorf("dec.Decode error: %w", err)
	}
	return reqModel, nil
}
