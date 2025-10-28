package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
	"github.com/RoGogDBD/loyalty_service/server/internal/service"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authService service.AuthService
	jwtService  service.JWTService
	logger      *logrus.Logger
}

func NewAuthHandler(authService service.AuthService, jwtService service.JWTService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtService:  jwtService,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		h.logger.WithError(err).Error("Failed to decode registration request")
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), &creds)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			http.Error(w, "Login already exists", http.StatusConflict)
			return
		}
		h.logger.WithError(err).WithField("login", creds.Login).Error("Failed to register user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID)
	if err != nil {
		h.logger.WithError(err).WithField("userID", user.ID).Error("Failed to generate token")
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Expose-Headers", "Authorization")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		h.logger.WithError(err).Error("Failed to decode login request")
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Login(r.Context(), &creds)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}
		h.logger.WithError(err).WithField("login", creds.Login).Error("Failed to login user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID)
	if err != nil {
		h.logger.WithError(err).WithField("userID", user.ID).Error("Failed to generate token")
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Expose-Headers", "Authorization")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}
