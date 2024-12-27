package user

import (
	"fmt"
	"net/http"

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

	// Register User
	if err := h.store.RegisterNewUser(
		types.User{
			Username: payload.Username,
			Email: payload.Email,
			Password: payload.Password,
			Role: DefaultRole,
		},
	); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error registering user: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "New User Created")
}


