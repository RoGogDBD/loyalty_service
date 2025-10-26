package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
	"github.com/RoGogDBD/loyalty_service/server/internal/repository"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type BalanceService interface {
	GetBalance(ctx context.Context, userID int) (*models.Balance, error)
	AddAccrual(ctx context.Context, userID int, amount float64) error
}

type balanceService struct {
	balanceRepo repository.BalanceRepository
}

func NewBalanceService(balanceRepo repository.BalanceRepository) BalanceService {
	return &balanceService{balanceRepo: balanceRepo}
}

func (s *balanceService) GetBalance(ctx context.Context, userID int) (*models.Balance, error) {
	balance, err := s.balanceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (s *balanceService) AddAccrual(ctx context.Context, userID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("invalid accrual amount")
	}

	if err := s.balanceRepo.AddAccrual(ctx, userID, amount); err != nil {
		return fmt.Errorf("failed to add accrual: %w", err)
	}

	return nil
}
