package myauth

import (
	"context"
	"fmt"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Authorizer struct {
	SecretKey string
	TokenExp  time.Duration
	l         *logger.ZapLog
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
	KeyAuthCtx Key = "set"
)

func (au *Authorizer) Auth(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		// Получаем логгер из контекста запроса
		logger, ok := r.Context().Value(keyLogger).(*logger.ZapLog)
		if !ok {
			log.Printf("Error getting logger")
			return
		}
		settedNewCookie := false
		_, err := r.Cookie("token")
		if err != nil {
			logger.ZL.Info("No cookie", zap.Error(err))
			// Cookie не установлена, устанавливаем
			newRandomUserId, err := au.setNewCookie(w)
			if err != nil {
				logger.ZL.Info("Error setting cookie:", zap.Error(err))
			}
			logger.ZL.Info("Success setted cookie", zap.Int("newRandomUserId", newRandomUserId))
			settedNewCookie = true
			ctx := context.WithValue(r.Context(), KeyAuthCtx, settedNewCookie)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx := context.WithValue(r.Context(), KeyAuthCtx, settedNewCookie)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
func (au *Authorizer) setNewCookie(w http.ResponseWriter) (int, error) {
	// Генерируем случайны ид пользователя
	randomID := rand.Intn(100)
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

// Claims описывает утверждения, хранящиеся в токене + добавляет кастомное UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// isValidateCookie проверяет валидность токена
func (au *Authorizer) isValidateCookie(tokenString string) (bool, error) {
	// Создаем экземпляр структуры с утверждениями
	claims := &Claims{}
	// Парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(au.SecretKey), nil
	})
	if err != nil {
		return false, fmt.Errorf("ошибка при парсинге токена %w", err)
	}
	return token.Valid, nil
}

// GetUserID возвращает ID пользователя
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
