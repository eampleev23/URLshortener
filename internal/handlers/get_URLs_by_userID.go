package handlers

import (
	"encoding/json"
	myauth "github.com/eampleev23/URLshortener/internal/auth"
	"github.com/eampleev23/URLshortener/internal/models"
	"go.uber.org/zap"
	"log"
	"net/http"
)

var keyAuth myauth.Key = myauth.KeyAuthCtx

func (h *Handlers) GetURLsByUserID(w http.ResponseWriter, r *http.Request) {
	setNewCookie, ok := r.Context().Value(keyAuth).(bool)
	if !ok {
		h.l.ZL.Info("Error getting if set new cookie")
		return
	}
	if setNewCookie {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Сюда мы попадем только если пользователь авторизован
	h.l.ZL.Info("GetURLsByUserID here")
	cookie, _ := r.Cookie("token")
	userID, _ := h.au.GetUserID(cookie.Value)
	h.l.ZL.Info("userID", zap.Int("userID", userID))
	ownersURLs, err := h.s.GetURLsByOwnerID(r.Context(), userID)
	if err != nil {
		h.l.ZL.Info("Error GetURLsByOwnerID:", zap.Error(err))
	}
	if len(ownersURLs) == 0 {
		log.Println("user ID ------->", userID)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	enc := json.NewEncoder(w)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json")
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
