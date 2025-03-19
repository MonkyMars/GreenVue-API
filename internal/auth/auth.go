package auth

// import (
// 	"time"

// 	"github.com/golang-jwt/jwt/v4"
// )

// var secretKey = []byte("your_secret_key")

// func GenerateToken(userID string) (string, error) {
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"user_id": userID,
// 		"exp":     time.Now().Add(time.Hour * 24).Unix(),
// 	})
// 	return token.SignedString(secretKey)
// }
