package handler

import (
	"encoding/json"
	"net/http"

	"github.com/slchris/wg-mgt/internal/middleware"
	"github.com/slchris/wg-mgt/internal/pkg/response"
	"github.com/slchris/wg-mgt/internal/service"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents a registration request.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangePasswordRequest represents a password change request.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// Login authenticates a user.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		response.Unauthorized(w, "invalid credentials")
		return
	}

	response.Success(w, map[string]string{"token": token})
}

// Setup creates the first admin user.
func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	isFirst, err := h.authService.IsFirstUser()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	if !isFirst {
		response.BadRequest(w, "setup already completed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	user, err := h.authService.Register(req.Username, req.Password, "admin")
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Created(w, map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

// CheckSetup checks if setup is needed.
func (h *AuthHandler) CheckSetup(w http.ResponseWriter, r *http.Request) {
	isFirst, err := h.authService.IsFirstUser()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, map[string]bool{"needs_setup": isFirst})
}

// ChangePassword changes the current user's password.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "unauthorized")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.authService.ChangePassword(claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "password changed"})
}

// Me returns the current user's information.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "unauthorized")
		return
	}

	response.Success(w, map[string]interface{}{
		"id":       claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}
