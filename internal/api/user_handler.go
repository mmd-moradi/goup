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
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

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
