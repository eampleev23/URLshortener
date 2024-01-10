package handlers

import (
	"encoding/json"
	"github.com/eampleev23/URLshortener/internal/models"
	"go.uber.org/zap"
	"net/http"
)

func (h *Handlers) GetURLsByUserID(w http.ResponseWriter, r *http.Request) {
	userIDCtx, ok := r.Context().Value(keyAuth).(int)
	if !ok {
		h.l.ZL.Info("Error getting if set new cookie")
		return
	}
	if userIDCtx == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Сюда мы попадем только если пользователь авторизован
	cookie, _ := r.Cookie("token")
	userID, _ := h.au.GetUserID(cookie.Value)
	h.l.ZL.Info("User id перед получением из базы", zap.Int("userID", userID))
	ownersURLs, err := h.s.GetURLsByOwnerID(r.Context(), userID)
	if err != nil {
		h.l.ZL.Info("Error GetURLsByOwnerID:", zap.Error(err))
	}
	if len(ownersURLs) == 0 {
		h.l.ZL.Info("User id когда попали в условие что 0 контента", zap.Int("userID", userID))
		w.WriteHeader(http.StatusNoContent)
		return
	}
	enc := json.NewEncoder(w)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	ownersURLsD, err := models.GetResponseGetOwnerURLs(ownersURLs)
	if err != nil {
		h.l.ZL.Info("GetResponseGetOwnerURLs error", zap.Error(err))
		return
	}
	if err := enc.Encode(ownersURLsD); err != nil {
		h.l.ZL.Info("error encoding response in handler", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
}
