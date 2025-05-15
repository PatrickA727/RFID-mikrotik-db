package user

import (
	// "context"
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
	router.HandleFunc("/register-user", auth.WithJWTAuth(h.handleRegisterUser, h.store)).Methods("POST")
	router.HandleFunc("/login", h.handleLoginUser).Methods("POST")
	router.HandleFunc("/logout", auth.WithJWTAuth(h.handleLogout, h.store)).Methods("POST")
	router.HandleFunc("/logout-all", auth.WithJWTAuth(h.handleLogoutAllDevice, h.store)).Methods("POST")
	router.HandleFunc("/delete-user", auth.WithJWTAuth(h.handleDeleteCurrentUser, h.store)).Methods("DELETE")
	router.HandleFunc("/refresh", h.handleRenewToken).Methods("POST")
	router.HandleFunc("/auth-client-mk", auth.WithJWTAuth(h.handleCheckAuthClient, h.store)).Methods("GET")
}

func (h *Handler) handleRenewToken(w http.ResponseWriter, r *http.Request) {
	refCookie, err := r.Cookie("refresh_token")
	if err != nil {
		if err == http.ErrNoCookie {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("permission denied3: %v", err))
			return
		}
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error retrieving refCookie: %v", err))
		return
	}

	sessionExists, userID, err := h.store.CheckSession(refCookie.Value)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error checking session: %v", err))
		return
	} 

	if !sessionExists {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unauthorized, session does not exist"))
		return
	} else {
		secret := []byte(os.Getenv("JWT_SECRET"))
		token, err := auth.CreateJWT(secret, userID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error generating token: %v", err))
			return
		}

		accessCookie := &http.Cookie{
			Name:     "access_token",                
			Value:    token,                      
			Expires:  time.Now().Add(time.Duration(600) * time.Second), 
			HttpOnly: true,	// SET TO TRUE FOR DEPLOY       
			Path: "/",        
			Secure:   true,                       
			SameSite: http.SameSiteNoneMode,       
		}

		http.SetCookie(w, accessCookie)
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{"msg": "new access token created"})
}

func (h *Handler) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
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
		},
	); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error registering user: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "New User Created")
}

func (h *Handler) handleCheckAuthClient(w http.ResponseWriter, r *http.Request) {
	// Check cookies
	cookie, err := r.Cookie("access_token")
	if err != nil {
		if err == http.ErrNoCookie {
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("error no cookie"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting access token: %v", err))
		return
	}

	// Validate JWT token
	token, err := auth.ValidateJWT(cookie.Value)
	if err != nil {
		// log.Println("token not valid: ", err)
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied1: %v", err))
		return
	}

	if !token.Valid {
		// log.Println("invalid token")
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied2: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusAccepted, map[string]string{"msg": "authorized"})
}

func (h *Handler) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	// Email sanitation
	payload.Email = utils.SanitizeInput(payload.Email)

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

	// Create access token
	secret := []byte(os.Getenv("JWT_SECRET"))
	token, err := auth.CreateJWT(secret, u.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error generating token: %v", err))
		return
	}

	// Create refresh token
	refToken, err := auth.CreateRefreshJWT(secret, u.ID) 
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error generating token: %v", err))
		return
	}

	// Create db session
	err = h.store.CreateSession(ctx, types.Session{
		Userid: u.ID,
		RefreshToken: refToken,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating session: %v", err))
		return
	}

	// Create JWT access token and log in user
	accessCookie := &http.Cookie{
		Name:     "access_token",                
		Value:    token,                      
		Expires:  time.Now().Add(time.Duration(900) * time.Second), 
		HttpOnly: true,	// SET TO TRUE FOR DEPLOY       
		Path: "/",        
		Secure:   true,                       
		SameSite: http.SameSiteNoneMode,       
	}

	// Create JWT refresh cookie
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",                
		Value:    refToken,                      
		Expires:  time.Now().Add(time.Duration(3600 * 24) * time.Second), // 1 day
		HttpOnly: true,	// SET TO TRUE FOR DEPLOY       
		Path: "/",        
		Secure:   true,                       
		SameSite: http.SameSiteNoneMode,       
	}

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *Handler) handleLogout (w http.ResponseWriter, r *http.Request) {
	// Get refresh token
	cookie, err := r.Cookie("refresh_token")
    if err != nil {
        // If there's no cookie, or any error retrieving it
        if err == http.ErrNoCookie {
            utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("refresh token not found"))
			return
        } else {
            utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error retrieving refresh token"))
			return
        }
    }

	// Get user ID from context
	ctx := r.Context()
	userID := ctx.Value(auth.UserKey)
	intID, ok := userID.(int)
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("ID type invalid"))
		return
	} 

    refreshToken := cookie.Value

	// Revoke Session
	err = h.store.RevokeSession(types.Session{
		Userid: intID,
		RefreshToken: refreshToken,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error revoking cookie: %v", err))
		return
	}

	// Create a cookie with the same name as the access token and refresh cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    "",
        Path:     "/",
        HttpOnly: true,	// SET TO TRUE FOR DEPLOY
        Expires:  time.Unix(0, 0), 
        MaxAge:   -1,             
        Secure:   true,        
    })

	http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Path:     "/",
        HttpOnly: true,	// SET TO TRUE FOR DEPLOY
        Expires:  time.Unix(0, 0), 
        MaxAge:   -1,             
        Secure:   true,        
    })

	utils.WriteJSON(w, http.StatusOK, map[string]string{"res": "Successfully logged out"})
}

func (h *Handler) handleLogoutAllDevice (w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	ctx := r.Context()
	userID := ctx.Value(auth.UserKey)

	intID, ok := userID.(int)
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("ID type invalid"))
		return
	} 

	// logout all sessions with this ID
	err := h.store.RevokeSessionBulk(intID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error logging out devices: %v", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    "",
        Path:     "/",
        HttpOnly: true,	// SET TO TRUE FOR DEPLOY
        Expires:  time.Unix(0, 0), 
        MaxAge:   -1,             
        Secure:   true,        
    })

	http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Path:     "/",
        HttpOnly: true,	// SET TO TRUE FOR DEPLOY
        Expires:  time.Unix(0, 0), 
        MaxAge:   -1,             
        Secure:   true,        
    })

	utils.WriteJSON(w, http.StatusOK, map[string]string{"msg": "Logged out all devices"})
}

func (h *Handler) handleDeleteCurrentUser (w http.ResponseWriter, r *http.Request) {
	// Get current userID from context
	ctx := r.Context()
	userID := ctx.Value(auth.UserKey)
	intID, ok := userID.(int)
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("ID type invalid"))
		return
	} 
	
	// Get user by id
	u, err := h.store.GetUserById(intID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user not found"))
		return
	}

	if u == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user not found"))
		return
	}

	// Expire cookie
	http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    "",
        Path:     "/",
        HttpOnly: true,	// SET TO TRUE FOR DEPLOY
        Expires:  time.Unix(0, 0), 
        MaxAge:   -1,             
        Secure:   true,        
    })

	http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Path:     "/",
        HttpOnly: true,	// SET TO TRUE FOR DEPLOY
        Expires:  time.Unix(0, 0), 
        MaxAge:   -1,             
        Secure:   true,        
    })

	// Delete user
	err = h.store.DeleteUserById(u.ID, ctx)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error deleting user"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"res": "user deleted"})
}