package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/eampleev23/URLshortener/internal/models"
	"go.uber.org/zap"
)

func (h *Handlers) GetURLsByUserID(w http.ResponseWriter, r *http.Request) {
	//userIDCtx, ok := r.Context().Value(keyUserIDCtx).(int)
	//if !ok {
	//	h.l.ZL.Debug("Error getting if set new cookie")
	//	return
	//}
	//if userIDCtx != 0 {
	//	// Значит это первый запрос пользователя (куку установили и у нас есть ид, но статус надо отдать не авторизован).
	//	w.WriteHeader(http.StatusUnauthorized)
	//	return
	//}
	//// Значит пользователь авторизован, надо получить id из куки
	//cookie, err := r.Cookie("token")
	//if err != nil {
	//	h.l.ZL.Info("Error getting cookie", zap.Error(err))
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	//userID, err := h.au.GetUserID(cookie.Value)
	//if err != nil {
	//	h.l.ZL.Info("Error getting userID from cookie", zap.Error(err))
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	//h.l.ZL.Debug("User id получили из куки (не из контекста)", zap.Int("userID", userID))

	userID, isAuth, err := h.GetUserID(r)
	if err != nil {
		h.l.ZL.Info("getting userID", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isAuth {
		// Значит это первый запрос пользователя (куку установили и у нас есть ид, но статус надо отдать не авторизован).
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ownersURLs, err := h.s.GetURLsByOwnerID(r.Context(), userID)
	if err != nil {
		h.l.ZL.Debug("Error GetURLsByOwnerID:", zap.Error(err))
	}
	if len(ownersURLs) == 0 {
		h.l.ZL.Debug("User id когда попали в условие что 0 контента", zap.Int("userID", userID))
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
