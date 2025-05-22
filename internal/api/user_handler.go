package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mmd-moradi/goup/internal/middleware"
	"github.com/mmd-moradi/goup/internal/service"
	"github.com/mmd-moradi/goup/pkg/apperrors"
	"github.com/mmd-moradi/goup/pkg/response"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with the provided information
// @Tags auth
// @Accept json
// @Produce json
// @Param input body service.UserRegistrationInput true "User registration information"
// @Success 201 {object} response.Response{data=service.AuthResponse} "User registered successfully"
// @Failure 400 {object} response.Response{error=response.ErrorInfo} "Invalid request payload"
// @Failure 409 {object} response.Response{error=response.ErrorInfo} "Username or email already exists"
// @Failure 500 {object} response.Response{error=response.ErrorInfo} "Internal server error"
// @Router /auth/register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input service.UserRegistrationInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		response.Error(w, apperrors.New(apperrors.BadRequest, "invalid request payload"))
		return
	}

	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(input.Email)

	authResponse, err := h.userService.Register(r.Context(), input)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, authResponse)
}

// Login handles user authentication
// @Summary Login a user
// @Description Authenticate a user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param input body service.UserLoginInput true "User login information"
// @Success 200 {object} response.Response{data=service.AuthResponse} "User logged in successfully"
// @Failure 400 {object} response.Response{error=response.ErrorInfo} "Invalid request payload"
// @Failure 401 {object} response.Response{error=response.ErrorInfo} "Invalid email or password"
// @Failure 500 {object} response.Response{error=response.ErrorInfo} "Internal server error"
// @Router /auth/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.UserLoginInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		response.Error(w, apperrors.New(apperrors.BadRequest, "invalid request payload"))
		return
	}

	input.Email = strings.TrimSpace(input.Email)

	authResponse, err := h.userService.Login(r.Context(), input)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, authResponse)

}

// GetProfile handles getting the current user's profile
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags auth
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=service.UserResponse} "User profile retrieved successfully"
// @Failure 401 {object} response.Response{error=response.ErrorInfo} "User not authenticated"
// @Failure 500 {object} response.Response{error=response.ErrorInfo} "Internal server error"
// @Router /auth/profile [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, user)
}

// Logout handles user logout
// @Summary Logout a user
// @Description Invalidate the user's authentication token
// @Tags auth
// @Produce json
// @Security Bearer
// @Success 204 "User logged out successfully"
// @Failure 401 {object} response.Response{error=response.ErrorInfo} "User not authenticated"
// @Failure 500 {object} response.Response{error=response.ErrorInfo} "Internal server error"
// @Router /auth/logout [post]
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		response.Error(w, apperrors.New(apperrors.Unauthorized, "authorization token is required"))
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		response.Error(w, apperrors.New(apperrors.Unauthorized, "authorization token is malformed, expected format: Bearer <token>"))
		return
	}

	token := parts[1]

	err := h.userService.Logout(r.Context(), token)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.NoContent(w)

}

func (h *UserHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Get("/logout", h.Logout)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/profile", h.GetProfile)
	})

}
