package handler

import (
	"encoding/json"
	"net/http"

	"github.com/RoGogDBD/loyalty_service/server/internal/middleware"
	"github.com/RoGogDBD/loyalty_service/server/internal/service"
	"github.com/sirupsen/logrus"
)

type BalanceHandler struct {
	balanceService service.BalanceService
	logger         *logrus.Logger
}

func NewBalanceHandler(balanceService service.BalanceService, logger *logrus.Logger) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService, logger: logger}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("userID", userID).Error("Failed to get balance")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
