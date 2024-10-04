package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type AuthManager struct {
	log       *zap.Logger
	secretKey []byte
}

func SetSecretKey(key []byte) Option {
	return func(a *AuthManager) {
		a.secretKey = key
	}
}

func SetLogger(log *zap.Logger) Option {
	return func(a *AuthManager) {
		a.log = log
	}
}

type Option func(*AuthManager)

func New(options ...Option) (*AuthManager, error) {
	a := &AuthManager{
		log: zap.NewNop(),
	}

	return a, nil
}

// VerifyJWT - Проверяет JWT.
func (a *AuthManager) VerifyJWT(signedData string) (string, bool) {
	token, err := jwt.Parse(signedData, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unknown signing method: %v", token.Header["alg"])
		}
		return a.secretKey, nil
	})

	if err != nil {
		a.log.Debug("failed parse jwt token", zap.Error(err))
		return "", false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uniqueID, ok := claims["uniqueID"].(string); ok {
			if uniqueID != "" {
				return uniqueID, true
			}
		}
	}

	return "", false
}

// CreateJWT - Создает JWT ключ и записывает в него ID пользователя.
func (a *AuthManager) CreateJWT(uniqueID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uniqueID": uniqueID,
	})
	tokenString, err := token.SignedString(a.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed signe token: %w", err)
	}

	return tokenString, nil
}
