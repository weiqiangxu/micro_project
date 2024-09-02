package common

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func GetUid(c *gin.Context) (uint64, error) {
	user, ok := c.Get("user")
	if !ok {
		return 0, errors.New("fail get user from context")
	}
	token, ok := user.(*jwt.Token)
	if !ok {
		return 0, errors.New("fail assert token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("fail assert claims")
	}
	uid, ok := claims["uid"].(float64)
	if !ok {
		return 0, errors.New("fail decode uid")
	}
	return uint64(uid), nil
}

func GenerateJwtToken(userId uint64, jwtSignKey string) (string, error) {
	claims := jwt.MapClaims{}
	claims["uid"] = userId
	claims["exp"] = time.Now().Add(7 * 24 * time.Hour).Unix()
	mySigningKey := []byte(jwtSignKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	return ss, err
}
