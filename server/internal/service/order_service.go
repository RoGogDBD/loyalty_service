package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
	"github.com/RoGogDBD/loyalty_service/server/internal/repository"
)

var (
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrOrderExists        = errors.New("order already uploaded by this user")
	ErrOrderConflict      = errors.New("order already uploaded by another user")
)

type OrderService interface {
	UploadOrder(ctx context.Context, number string, userID int) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]*models.Order, error)
	UpdateStatus(ctx context.Context, orderID int, status string, accrual float64) error
	GetPendingOrders(ctx context.Context) ([]*models.Order, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(orderRepo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

func (s *orderService) UploadOrder(ctx context.Context, number string, userID int) (*models.Order, error) {
	// Валидация по алгоритму Луна
	if !repository.ValidateLuhn(number) {
		return nil, ErrInvalidOrderNumber
	}

	// Проверяем существование заказа
	existing, err := s.orderRepo.GetByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to check order existence: %w", err)
	}

	if existing != nil {
		if existing.UserID == userID {
			return nil, ErrOrderExists
		}
		return nil, ErrOrderConflict
	}

	// Создаём заказ
	order, err := s.orderRepo.Create(ctx, number, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

func (s *orderService) GetUserOrders(ctx context.Context, userID int) ([]*models.Order, error) {
	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, nil
}

func (s *orderService) UpdateStatus(ctx context.Context, orderID int, status string, accrual float64) error {
	return s.orderRepo.UpdateStatusByID(ctx, orderID, status, accrual)
}

func (s *orderService) GetPendingOrders(ctx context.Context) ([]*models.Order, error) {
	orders, err := s.orderRepo.GetPendingOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}
	return orders, nil
}
