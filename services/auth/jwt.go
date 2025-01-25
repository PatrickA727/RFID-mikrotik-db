package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"context"
	"github.com/PatrickA727/mikrotik-db-sys/types"
	"github.com/PatrickA727/mikrotik-db-sys/utils"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string
const UserKey contextKey = "userID"

func CreateJWT(secret []byte, userID int) (string, error) {
	expiration := time.Duration(900) * time.Second	// 15 minutes

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

func CreateRefreshJWT(secret []byte, userID int) (string, error) {
	expiration := time.Duration(3600 * 24 * 30) * time.Second // 30 days

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

func WithJWTAuth(handlerFunc http.HandlerFunc, store types.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from cookies
		tokenString := utils.GetTokenFromCookie(r)

		// Validate JWT
		token, err := ValidateJWT(tokenString)
		if err != nil {
			log.Println("token not valid: ", err)
			utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied: %v", err))
			return
		}

		if !token.Valid {
			log.Println("invalid token")
			utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied: %v", err))
			return
		}

		// Get userID from JWT claims
		claims := token.Claims.(jwt.MapClaims)
		str := claims["userID"].(string)

		userID, err := strconv.Atoi(str)
		if err != nil {
			log.Printf("failed to convert userID to int: %v", err)
			utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied: %v", err))
			return
		}

		// Fetch user by id from database
		u, err := store.GetUserById(userID)
		if err != nil {
			log.Printf("failed to get user by id: %v", err)
			utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied: %v", err))
			return
		}

		// Set the userId to the ctx(context) so the handler functions have access to current user id in the ctx
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserKey, u.ID) // Creates a new context that contains UserKey("userid") as the key and user.id as the value
		r = r.WithContext(ctx)	// Attaches the new context to the original request containing the userID

		// Run the handler func with validated user JWT cookie
		handlerFunc(w, r)
	}
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {	// Validates JWT by checking its signing method
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {	// JWT Parse method takes tokenString and a callback func to check/validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {	// Accesses and checks the token signing method (has to be HMAC)
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])	// Shows the signing method of the incorrect jwt token
		}

		return []byte(os.Getenv("JWT_SECRET")), nil	// The CALLBACK FUNC returns the secret key to be used by the jwt.parse func
	})
}

// For validateJWT the if statement and checking signing method IS THE CALLBACK PARAM for the jwt.parse function
