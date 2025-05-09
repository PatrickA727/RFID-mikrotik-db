package utils

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"regexp"
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

func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	input = html.EscapeString(input)	// Changes the HTML characters such as <, >, etc so it cant run a script
	re := regexp.MustCompile(`[^\w\s@.-]`)	// Removes unallowed characters such as %, $, &, etc
	return re.ReplaceAllString(input, "")	// Usees the regex from before "re"
}
