package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/models"
	"go.uber.org/zap"
	"net/http"
	"net/url"
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
	shortURL, err := h.s.SetShortURL(r.Context(), reqModel.OriginalURL, userID)
	if err != nil {
		h.l.ZL.Info("SetShortURL error", zap.Error(err))
		shortURL, err = returnShortURLIfConflict(h, w, r, reqModel.OriginalURL, err)
		if err != nil {
			h.l.ZL.Info("returnShortURLIfConflict error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusConflict)
		resp := models.ResponseAddShortURL{ShortURL: shortURL}
		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			h.l.ZL.Info("error encoding response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	shortURL, err = url.JoinPath(h.c.BaseShortURL, shortURL)
	if err != nil {
		h.l.ZL.Info("url.JoinPath error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := models.ResponseAddShortURL{ShortURL: shortURL}
	enc := json.NewEncoder(w)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = enc.Encode(resp)
	if err != nil {
		h.l.ZL.Info("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getModelRequest(r *http.Request) (reqModel models.RequestAddShortURL, err error) {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reqModel); err != nil {
		return models.RequestAddShortURL{}, fmt.Errorf("dec.Decode error: %w", err)
	}
	return reqModel, nil
}
