package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
	"github.com/RoGogDBD/loyalty_service/server/internal/service"
	"github.com/sirupsen/logrus"
)

type AccrualWorker struct {
	accrualURL     string
	orderService   service.OrderService
	balanceService service.BalanceService
	logger         *logrus.Logger
	client         *http.Client
}

func NewAccrualWorker(
	accrualURL string,
	orderService service.OrderService,
	balanceService service.BalanceService,
	logger *logrus.Logger,
) *AccrualWorker {
	return &AccrualWorker{
		accrualURL:     accrualURL,
		orderService:   orderService,
		balanceService: balanceService,
		logger:         logger,
		client:         &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *AccrualWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	w.logger.Info("Accrual worker started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Accrual worker stopped")
			return
		case <-ticker.C:
			if err := w.processOrders(ctx); err != nil {
				w.logger.WithError(err).Error("Failed to process orders")
			}
		}
	}
}

func (w *AccrualWorker) processOrders(ctx context.Context) error {
	orders, err := w.orderService.GetPendingOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending orders: %w", err)
	}

	for _, order := range orders {
		if err := w.processOrder(ctx, order); err != nil {
			w.logger.WithError(err).WithField("orderNumber", order.Number).Error("Failed to process order")
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (w *AccrualWorker) processOrder(ctx context.Context, order *models.Order) error {
	url := fmt.Sprintf("%s/api/orders/%s", w.accrualURL, order.Number)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call accrual system: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp struct {
			Order   string  `json:"order"`
			Status  string  `json:"status"`
			Accrual float64 `json:"accrual,omitempty"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		var orderStatus string
		switch accrualResp.Status {
		case "REGISTERED":
			orderStatus = models.OrderStatusNew
		case "PROCESSING":
			orderStatus = models.OrderStatusProcessing
		case "INVALID":
			orderStatus = models.OrderStatusInvalid
		case "PROCESSED":
			orderStatus = models.OrderStatusProcessed
		default:
			w.logger.WithField("status", accrualResp.Status).Warn("Unknown status from accrual system")
			orderStatus = accrualResp.Status
		}

		if err := w.orderService.UpdateStatus(ctx, order.ID, orderStatus, accrualResp.Accrual); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		if orderStatus == models.OrderStatusProcessed && accrualResp.Accrual > 0 {
			if err := w.balanceService.AddAccrual(ctx, order.UserID, accrualResp.Accrual); err != nil {
				return fmt.Errorf("failed to add accrual: %w", err)
			}
			w.logger.WithFields(logrus.Fields{
				"orderNumber": order.Number,
				"accrual":     accrualResp.Accrual,
				"userID":      order.UserID,
			}).Info("Accrual added")
		}

		return nil

	case http.StatusNoContent:
		w.logger.WithField("orderNumber", order.Number).Debug("Order not registered in accrual system yet")
		return nil

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				w.logger.WithField("retryAfter", seconds).Warn("Rate limited by accrual system")
				time.Sleep(time.Duration(seconds) * time.Second)
			}
		}
		return nil

	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
