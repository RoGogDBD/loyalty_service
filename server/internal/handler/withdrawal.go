package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RoGogDBD/loyalty_service/server/internal/middleware"
	"github.com/RoGogDBD/loyalty_service/server/internal/models"
	"github.com/RoGogDBD/loyalty_service/server/internal/service"
	"github.com/sirupsen/logrus"
)

type WithdrawalHandler struct {
	withdrawalService service.WithdrawalService
	logger            *logrus.Logger
}

func NewWithdrawalHandler(withdrawalService service.WithdrawalService, logger *logrus.Logger) *WithdrawalHandler {
	return &WithdrawalHandler{withdrawalService: withdrawalService, logger: logger}
}

func (h *WithdrawalHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode withdrawal request")
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	_, err := h.withdrawalService.Withdraw(r.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidWithdrawalOrder) {
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, service.ErrInsufficientFunds) {
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
			return
		}
		h.logger.WithError(err).WithField("userID", userID).Error("Failed to withdraw")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WithdrawalHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.withdrawalService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("userID", userID).Error("Failed to get withdrawals")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
