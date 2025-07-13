package jwt

import (
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
