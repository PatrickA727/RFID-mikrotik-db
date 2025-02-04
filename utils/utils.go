package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()	

func ParseJSON(r *http.Request, payload any) error {	
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)	
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {	
	w.Header().Add("Content-Type", "application/json")	
	w.WriteHeader(status)	

	return json.NewEncoder(w).Encode(v)	
}

func WriteError(w http.ResponseWriter, status int, err error) {		
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func GetTokenFromCookie(r *http.Request) string {
    // Retrieve the access token cookie from the request
    cookie, err := r.Cookie("access_token")
    if err != nil {
		if err == http.ErrNoCookie {
			// log.Println("cookie not found")
			return ""
		}
		
		log.Println("error getting jwt cookie: ", err)
        return ""
    }

    return cookie.Value
}

func MobileAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		sigHeader := r.Header.Get("Signature")
		timeHeader := r.Header.Get("Timestamp")
		apiUrl := r.URL

		timeNow := time.Now()
	
		// Convert the string timestamp to an integer (assuming it's in Unix seconds)
		timeInt, err := strconv.ParseInt(timeHeader, 10, 64)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, fmt.Errorf("invalid timestamp format: %v", err))
			return 
		}
	
		// Convert the Unix timestamp to a time.Time object
		requestTime := time.Unix(timeInt, 0)
	
		// Check if the request time is older than 5 minutes
		if timeNow.Sub(requestTime) > 5*time.Minute {
			WriteError(w, http.StatusBadRequest, fmt.Errorf("request time expired: %v", err))
			return 
		}
	
		signData := timeHeader + apiUrl.Path
		key := os.Getenv("SIGN_SECRET")

		h := hmac.New(sha256.New, []byte(key))
		h.Write([]byte(signData))
		sigBackend := hex.EncodeToString(h.Sum(nil))

		if sigHeader != sigBackend {
			WriteError(w, http.StatusForbidden, fmt.Errorf("invalid signature"))
			return	
		}
	
		handlerFunc(w, r)
	}
}