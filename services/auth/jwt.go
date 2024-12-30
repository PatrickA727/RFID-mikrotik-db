package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateJWT(secret []byte, userID int) (string, error) {
	expiration := time.Second + time.Duration(3600*24*7) 

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{	// Create new JWt token, with claims(key value pairs embedded in the token)
		"userID": strconv.Itoa(userID),									// Uses the HS256 signing method, its  fast method for single server systems with low complexity
		"expiredAt": time.Now().Add(expiration).Unix(),
	})

	tokenString, err := token.SignedString(secret)	// The final token signed with the secret key
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
