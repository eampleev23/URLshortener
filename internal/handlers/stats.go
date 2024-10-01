package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/models"
	"go.uber.org/zap"
	"net/http"
	"net/netip"
)

// GetGlobalStats - хэндлер для получения общей статистики внутри доверенной сети.
func (h *Handlers) GetGlobalStats(w http.ResponseWriter, r *http.Request) {
	h.l.ZL.Info("GetGlobalStats called")
	// получаем айпишник от resolveIP
	ip, err := resolveIP(r)
	// Если ip нет в заголовке.
	if err != nil {
		// Логируем ошибку
		h.l.ZL.Info("Failed to resolve IP", zap.Error(err))
		// Ставим заголовок 403
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// Логируем полученный ip
	h.l.ZL.Info("Got IP", zap.Any("ip", ip))
	network, err := netip.ParsePrefix(h.c.TrustedSubnet)
	if err != nil {
		h.l.ZL.Info("Failed to parse trusted subnet", zap.Error(err))
		w.WriteHeader(http.StatusForbidden)
		return
	}
	isIPWeTrust := network.Contains(ip)
	if !isIPWeTrust {
		h.l.ZL.Info("Is IP WeTrust", zap.Bool("isIPWeTrust", isIPWeTrust))
		w.WriteHeader(http.StatusForbidden)
		return
	}
	resp := models.ResponseGlobalStats{}
	resp.URLs, err = h.serv.GetURLsSavedCount(r.Context())
	if err != nil {
		h.l.ZL.Info("Failed to get URLs saved count", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp.Users, err = h.serv.GetUsersCount(r.Context())
	if err != nil {
		h.l.ZL.Info("Failed to get Users saved count", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	enc := json.NewEncoder(w)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = enc.Encode(resp)
	if err != nil {
		h.l.ZL.Info("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// resolveIP - внутренняя функция для вычисления оригинального ip-адреса. Получает запрос и объект resolveIPOpts,
// возвращает net.IP и error.
func resolveIP(r *http.Request) (netip.Addr, error) {
	// смотрим заголовок запроса X-Real-IP
	ipStr := r.Header.Get("X-Real-IP")
	// парсим ip
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return ip, fmt.Errorf("failed to parse IP: %w", err)
	}
	return ip, nil
}
