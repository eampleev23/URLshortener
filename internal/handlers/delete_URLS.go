package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handlers) DeleteURLS(w http.ResponseWriter, r *http.Request) {
	h.l.ZL.Info("DeleteURLS start..")
	// Сначала преобразуем входящие данные в массив моделей.
	var req []string
	// Далее создаем декодер
	dec := json.NewDecoder(r.Body)
	// В случае некорректного запроса, возвращаем соответствующий заголовок
	if err := dec.Decode(&req); err != nil {
		h.l.ZL.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// В случае успешного парсинга в массив моделей, возвращаем статус 202 Accepted
	w.WriteHeader(http.StatusAccepted)

	// Далее передаем в модель данные для обработки.
	userIDCtx, ok := r.Context().Value(keyUserIDCtx).(int)
	if !ok {
		h.l.ZL.Info("Error getting if set new cookie")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userIDCtx != 0 {
		err := h.s.DeleteURLS(r.Context(), userIDCtx, req)
		if err != nil {
			h.l.ZL.Info("h.s.DeleteURLS error: ", zap.Error(err))
		}
	} else {
		cookie, _ := r.Cookie("token")
		userID, _ := h.au.GetUserID(cookie.Value)
		err := h.s.DeleteURLS(r.Context(), userID, req)
		if err != nil {
			h.l.ZL.Info("h.s.DeleteURLS error: ", zap.Error(err))
		}
	}

}