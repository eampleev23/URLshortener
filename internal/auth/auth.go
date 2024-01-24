package myauth

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type Authorizer struct {
	l         *logger.ZapLog
	SecretKey string
	TokenExp  time.Duration
}

var keyLogger logger.Key = logger.KeyLoggerCtx

// Initialize инициализирует синглтон авторизовывальщика с секретным ключом.
func Initialize(secretKey string, tokenExp time.Duration, l *logger.ZapLog) (*Authorizer, error) {
	au := &Authorizer{
		SecretKey: secretKey,
		TokenExp:  tokenExp,
		l:         l,
	}
	return au, nil
}

type Key string

const (
	KeyUserIDCtx Key = "user_id_ctx"
)

func (au *Authorizer) Auth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("token")
		if err != nil {
			// Получаем логгер из контекста запроса
			logger, ok := r.Context().Value(keyLogger).(*logger.ZapLog)
			if !ok {
				log.Printf("Error getting logger")
				return
			}
			logger.ZL.Info("No cookie", zap.String("err", err.Error()))
			// Cookie не установлена, устанавливаем
			newRandomUserID, err := au.setNewCookie(w)
			if err != nil {
				logger.ZL.Info("Error setting cookie:", zap.String("err", err.Error()))
			}
			logger.ZL.Info("Success setted cookie", zap.Int("newRandomUserId", newRandomUserID))
			ctx := context.WithValue(r.Context(), KeyUserIDCtx, newRandomUserID)
			logger.ZL.Info("Передали newRandomUserID", zap.Int("newRandomUserID", newRandomUserID))
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		// если кука уже установлена, то через контекст передаем 0
		ctx := context.WithValue(r.Context(), KeyUserIDCtx, 0)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
func (au *Authorizer) setNewCookie(w http.ResponseWriter) (int, error) {
	// Генерируем случайный ид пользователя.
	maxID := 10000
	randomID := rand.Intn(maxID)
	au.l.ZL.Info("Generated random ID", zap.Int("randomID", randomID))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// Когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(au.TokenExp)),
		},
		// Собственное утверждение
		UserID: randomID,
	})
	tokenString, err := token.SignedString([]byte(au.SecretKey))
	if err != nil {
		return randomID, fmt.Errorf("ошибка при генерации нового токена %w", err)
	}
	cookie := http.Cookie{
		Name:  "token",
		Value: tokenString,
	}
	http.SetCookie(w, &cookie)
	return randomID, nil
}

// Claims описывает утверждения, хранящиеся в токене + добавляет кастомное UserID.
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// GetUserID возвращает ID пользователя.
func (au *Authorizer) GetUserID(tokenString string) (int, error) {
	// Создаем экземпляр структуры с утверждениями
	claims := &Claims{}
	// Парсим из строки токена tokenString в структуру claims
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(au.SecretKey), nil
	})
	if err != nil {
		au.l.ZL.Info("Failed in case to get ownerId from token ", zap.Error(err))
	}

	return claims.UserID, nil
}
