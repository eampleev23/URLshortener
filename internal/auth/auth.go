package myauth

import (
	"fmt"
	"github.com/eampleev23/URLshortener/internal/logger"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"log"
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

// Auth — middleware авторизации.
func (au *Authorizer) Auth(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		// Получаем логгер из контекста запроса
		logger, ok := r.Context().Value(keyLogger).(*logger.ZapLog)
		if !ok {
			log.Printf("Error getting logger")
			return
		}

		cookie, err := r.Cookie("token")
		fmt.Println(cookie)
		if err != nil {
			logger.ZL.Info("No cookie", zap.Error(err))
			// Cookie не установлена, устанавливаем
			err := au.setNewCookie(w)
			if err != nil {
				logger.ZL.Info("Error setting cookie:", zap.Error(err))
			}
			logger.ZL.Info("Success setted cookie")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)

	}
	return http.HandlerFunc(fn)
}
func (au *Authorizer) setNewCookie(w http.ResponseWriter) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// Когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(au.TokenExp)),
		},
		// Собственное утверждение
		UserID: 12,
	})
	tokenString, err := token.SignedString([]byte(au.SecretKey))
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
