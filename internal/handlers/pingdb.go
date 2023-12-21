package handlers

import (
	"log"
	"net/http"
)

func (h *Handlers) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.s.PingDB()
	if err != nil {
		log.Printf("not ping %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
