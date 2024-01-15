package handlers

import "net/http"

func (h *Handlers) DeleteURLS(w http.ResponseWriter, r *http.Request) {
	h.l.ZL.Info("DeleteURLS start..")
}
