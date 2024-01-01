package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/store"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

/*
GetURLsByUserID возвращает пользователю все когда-либо сокращённые им URL в формате(id получает в куке):
[

	{
	    "short_url": "http://...",
	    "original_url": "http://..."
	},
	...

]
*/
func (h *Handlers) GetURLsByUserID(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		log.Printf("no cookie: %v", err)
		// Cookie не установлена, устанавливаем
		err := h.setNewCookie(w)
		if err != nil {
			log.Printf("error setting cookie: %v", err)
		}
		log.Printf("success setted cookie")
		log.Printf("token was empty:`%s`", cookie)
		enc := json.NewEncoder(w)
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
		log.Printf("ошибка при проверке на валидность токена из куки %v", err)
	}
	// Обрабатываем если значение не валидное
	if !validCookie {
		// не валидная
		w.WriteHeader(http.StatusUnauthorized)
		return

	}
	// Получаем все ссылки для пользователя
	db, err := sql.Open("pgx", h.c.DBDSN) //nolint:goconst // не понятно зачем константа
	if err != nil {
		log.Printf("failed to open a connection to the DB in GetURLsByUserID %v", err)
	}
	// Создаем экземпляр структуры с утверждениями
	claims := &Claims{}
	// Парсим из строки токена tokenString в структуру claims
	_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.c.SecretKey), nil
	})
	if err != nil {
		log.Printf("failed in case to get ownerId from token %v", err)
	}

	ownersURLs, err := store.GetURLsByOwnerID(context.Background(), db, claims.UserID)
	if err != nil {
		log.Printf("error GetURLsByOwnerID: %v", err)
	}
	if len(ownersURLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(ownersURLs); err != nil {
		h.l.ZL.Info("error encoding response in handler", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusOK)

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
		Name:  "token",
		Value: tokenString,
	}
	http.SetCookie(w, &cookie)
	return nil
}
