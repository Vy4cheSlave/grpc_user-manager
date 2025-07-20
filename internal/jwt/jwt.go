package jwt

import (
	"fmt"
	domainAuth "github.com/Vy4cheSlave/grpc_user-manager/internal/domain/auth"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func NewToken(user *domainAuth.User, duration time.Duration, secret []byte) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.Id
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(duration).Unix()
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateToken(tokenString string, secret []byte) (*jwt.MapClaims, error) {
	// Парсим токен с указанием ключа и метода подписи
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}
