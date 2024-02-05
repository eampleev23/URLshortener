package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// CreateShortURL получает в пост запросе урл, который необходимо сократить и возвращает сокращенный в виде text/plain.
func (h *Handlers) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	// Получаем originalURL в виде строки
	originalURL, err := getOriginURLFromReq(r)
	if err != nil {
		h.l.ZL.Info("getOriginURLFromReq error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.l.ZL.Debug("Got originalURL", zap.String("originalURL", originalURL))

	// Получаем id пользователя.
	userID, _, err := h.GetUserID(r)
	if err != nil {
		h.l.ZL.Info("Error getting userID", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.l.ZL.Debug("Got ID of user", zap.Int("userID", userID)) //nolint:goconst // it's just ok

	// Далее пробуем создать ссылку, но нам нужно знать есть ли конфликт данных
	shortURL, err := h.s.SetShortURL(r.Context(), originalURL, userID)
	if err != nil {
		h.l.ZL.Info("SetShortURL error", zap.Error(err))
		shortURL, err = returnShortURLIfConflict(h, r, originalURL, err)
		if err != nil {
			h.l.ZL.Info("returnShortURLIfConflict error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusConflict)
		w.Header().Set("content-type", "text/plain")
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			h.l.ZL.Info("Ошибка в хэндлере CreateShortLink при вызове w.Write", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	h.l.ZL.Debug("SetShortURL success", zap.String("shortURL", shortURL))
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("content-type", "text/plain")
	shortURL, err = url.JoinPath(h.c.BaseShortURL, shortURL)
	if err != nil {
		h.l.ZL.Info("error url.joinpath", zap.Error(err))
	}
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		h.l.ZL.Info("w.Write error: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getOriginURLFromReq(r *http.Request) (originalURL string, err error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("io.ReadAll(r.Body) error: %w", err)
	}
	originalURL = string(reqBody)
	return originalURL, nil
}

// LabelError описывает ошибку с дополнительной меткой.
type LabelError struct {
	Err   error
	Label string // метка должна быть в верхнем регистре

}

// Error добавляет поддержку интерфейса error для типа LabelError.
func (le *LabelError) Error() string {
	return fmt.Sprintf("[%s] %v", le.Label, le.Err)
}

// NewLabelError упаковывает ошибку err в тип LabelError.
func NewLabelError(label string, err error) error {
	return &LabelError{
		Label: strings.ToUpper(label),
		Err:   err,
	}
}

func returnShortURLIfConflict(
	h *Handlers,
	r *http.Request,
	originalURL string,
	errIn error) (shortURL string, errOut error) {
	var pgErr *pgconn.PgError
	if errors.As(errIn, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		if strings.Contains(errIn.Error(), "links_couples_index_by_original_url_unique") {
			// Здесь логика обработки конфликта.
			myErr := NewLabelError("conflict", errIn)
			h.l.ZL.Debug(" ", zap.Error(myErr))
			h.l.ZL.Debug("This originalURL already exists", zap.String("originalURL", originalURL))
			shortURL, errOut = h.s.GetShortURLByOriginal(r.Context(), originalURL)
			if errOut != nil {
				return "", fmt.Errorf("GetShortURLByOriginal error: %w", errOut)
			}
			shortURL, errOut = url.JoinPath(h.c.BaseShortURL, shortURL)
			if errOut != nil {
				return "", fmt.Errorf("url.JoinPath error: %w", errOut)
			}
			return shortURL, nil
		}
		if strings.Contains(errIn.Error(), "links_couples_index_by_short_url_unique") {
			// здесь логика обработки коллизии
			myErr := NewLabelError("collision", errIn)
			h.l.ZL.Debug(" ", zap.Error(myErr))
			return "", fmt.Errorf("collision: %w", myErr)
		}
	}
	return "", errors.New("no pgErr")
}
