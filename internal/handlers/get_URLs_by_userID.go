package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

func (h *Handlers) GetURLsByUserID(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("ID")
	if err != nil {
		h.l.ZL.Info("No cookie", zap.Error(err))
		// Cookie не установлена, устанавливаем
		err := h.setNewCookie(w)
		if err != nil {
			h.l.ZL.Info("Error setting cookie:", zap.Error(err))
		}
		h.l.ZL.Info("Success setted cookie")
		enc := json.NewEncoder(w)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := enc.Encode("[{}]"); err != nil {
			h.l.ZL.Info("error encoding response in handler", zap.Error(err))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		return
	}
	// Cookie установлена
	// Надо проверить на валидность
	validCookie, err := h.isValidateCookie(cookie.Value)
	if err != nil {
		h.l.ZL.Info("Ошибка при проверке на валидность токена из куки", zap.Error(err))
	}
	// Обрабатываем если значение не валидное
	if !validCookie {
		// не валидная
		w.WriteHeader(http.StatusUnauthorized)
		return

	}
	// Получаем все ссылки для пользователя
	// Создаем экземпляр структуры с утверждениями
	claims := &Claims{}
	// Парсим из строки токена tokenString в структуру claims
	_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.c.SecretKey), nil
	})
	if err != nil {
		h.l.ZL.Info("Failed in case to get ownerId from token ", zap.Error(err))
	}

	ownersURLs, err := h.s.GetURLsByOwnerID(r.Context(), claims.UserID)
	log.Println("ownersURLs", ownersURLs)
	if err != nil {
		h.l.ZL.Info("Error GetURLsByOwnerID:", zap.Error(err))
	}
	if len(ownersURLs) == 0 {
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

// Claims описывает утверждения, хранящиеся в токене + добавляет кастомное UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

// isValidateCookie проверяет валидность токена
func (h *Handlers) isValidateCookie(tokenString string) (bool, error) {
	// Создаем экземпляр структуры с утверждениями
	claims := &Claims{}
	// Парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.c.SecretKey), nil
	})
	if err != nil {
		return false, fmt.Errorf("ошибка при парсинге токена %w", err)
	}
	return token.Valid, nil
}

func (h *Handlers) setNewCookie(w http.ResponseWriter) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// Когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.c.TokenEXP)),
		},
		// Собственное утверждение
		UserID: 12,
	})
	tokenString, err := token.SignedString([]byte(h.c.SecretKey))
	if err != nil {
		return fmt.Errorf("ошибка при генерации нового токена %w", err)
	}
	cookie := http.Cookie{
		Name:  "ID",
		Value: tokenString,
	}
	http.SetCookie(w, &cookie)
	return nil
}
