package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/RoGogDBD/loyalty_service/server/internal/middleware"
	"github.com/RoGogDBD/loyalty_service/server/internal/service"
	"github.com/sirupsen/logrus"
)

type OrderHandler struct {
	orderService service.OrderService
	logger       *logrus.Logger
}

func NewOrderHandler(orderService service.OrderService, logger *logrus.Logger) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		logger:       logger,
	}
}
func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read request body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.WithError(err).Error("Failed to close request body")
		}
	}()

	orderNumber := string(body)
	if orderNumber == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}

	_, err = h.orderService.UploadOrder(r.Context(), orderNumber, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderExists):
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, service.ErrOrderConflict):
			w.WriteHeader(http.StatusConflict)
		case errors.Is(err, service.ErrInvalidOrderNumber):
			w.WriteHeader(http.StatusUnprocessableEntity) // 422
		default:
			h.logger.WithError(err).Error("Failed to upload order")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetUserOrders(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get orders")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
