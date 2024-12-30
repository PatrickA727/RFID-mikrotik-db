package user

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"github.com/PatrickA727/mikrotik-db-sys/services/auth"
	"github.com/PatrickA727/mikrotik-db-sys/types"
	"github.com/PatrickA727/mikrotik-db-sys/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store types.UserStore
}

func NewHandler (store types.UserStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/register-user", h.handleRegisterUser).Methods("POST")
	router.HandleFunc("/login", h.handleLoginUser).Methods("POST")
	router.HandleFunc("/logout", h.handleLogout).Methods("POST")

}

func (h *Handler) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	var DefaultRole = "user"
	// Get JSON
	var payload types.UserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parsing user: %v", err))
		return
	}

	// Validate JSON
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// Check if user exists
	_, err := h.store.GetUserByEmail(payload.Email)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user already exists: %v", payload.Email))
		return
	}

	// Hash pass
	hashedPass, err := auth.HashPass(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error hashing pass: %v", err))
	}

	// Register User
	if err := h.store.RegisterNewUser(
		types.User{
			Username: payload.Username,
			Email: payload.Email,
			Password: hashedPass,
			Role: DefaultRole,
		},
	); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error registering user: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "New User Created")
}

func (h *Handler) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	// Get JSON
	var payload types.LoginPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parsing user: %v", err))
		return
	}

	// Validate JSON
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// Check if user email exists
	u, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("email or password incorrect"))
		return
	}

	// Check user password
	if !auth.ComparePasswords(u.Password, []byte(payload.Password)) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("email or password incorrect"))
		return
	}

	// Create JWT token
	secret := []byte(os.Getenv("JWT_SECRET"))
	token, err := auth.CreateJWT(secret, u.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error generating token: %v", err))
		return
	}

	// Create JWT token and log in user
	cookie := &http.Cookie{
		Name:     "jwt",                
		Value:    token,                      
		Expires:  time.Now().Add(7 * 24 * time.Hour), 
		HttpOnly: true,                       
		Secure:   true,                       // ENABLE ON DEPLOYMENT
		SameSite: http.SameSiteStrictMode,       
	}

	http.SetCookie(w, cookie)

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *Handler) handleLogout (w http.ResponseWriter, r *http.Request) {
	// Create a cookie with the same name as the JWT cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "jwt", // Use the same name as your JWT cookie
        Value:    "",
        // Path:     "/",
        HttpOnly: true,
        Expires:  time.Unix(0, 0), 
        MaxAge:   -1,             
        Secure:   true,           // ENABLE ON DEPLOYMENT
    })

	utils.WriteJSON(w, http.StatusOK, map[string]string{"res": "Successfully logged out"})
}

func (h *Handler) handleDeleteUser (w http.ResponseWriter, r *http.Request) {
	// Get the JWT from the cookie
    cookie, err := r.Cookie("jwt")
    if err != nil {
        utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("missing or invalid token"))
        return
    }

    // Parse and validate the JWT
    token, claims, err := auth.ParseJWT(cookie.Value, []byte(os.Getenv("JWT_SECRET")))
    if err != nil || !token.Valid {
        utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
        return
    }

    // Extract the user ID from the claims
    userID, ok := claims["id"].(string) // Assuming ID is stored as a string in the token
    if !ok {
        utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid token claims"))
        return
    }
}