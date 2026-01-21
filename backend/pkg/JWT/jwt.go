package JWT

import (
	"backend/internal/domain/entities"
	"github.com/golang-jwt/jwt/v4"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT(user *entities.UserInfo) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID(),
		"username": user.Username(),
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
