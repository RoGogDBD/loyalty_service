package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
	"github.com/RoGogDBD/loyalty_service/server/internal/repository"
)

var ErrInvalidWithdrawalOrder = errors.New("invalid withdrawal order number")

type WithdrawalService interface {
	Withdraw(ctx context.Context, userID int, req *models.WithdrawalRequest) (*models.Withdrawal, error)
	GetWithdrawals(ctx context.Context, userID int) ([]*models.Withdrawal, error)
}

type withdrawalService struct {
	withdrawalRepo repository.WithdrawalRepository
	balanceRepo    repository.BalanceRepository
}

func NewWithdrawalService(
	withdrawalRepo repository.WithdrawalRepository,
	balanceRepo repository.BalanceRepository,
) WithdrawalService {
	return &withdrawalService{
		withdrawalRepo: withdrawalRepo,
		balanceRepo:    balanceRepo,
	}
}

func (s *withdrawalService) Withdraw(ctx context.Context, userID int, req *models.WithdrawalRequest) (*models.Withdrawal, error) {
	if !repository.ValidateLuhn(req.OrderNumber) {
		return nil, ErrInvalidWithdrawalOrder
	}

	if req.Sum <= 0 {
		return nil, fmt.Errorf("invalid withdrawal sum")
	}

	if err := s.balanceRepo.Withdraw(ctx, userID, req.Sum); err != nil {
		if err.Error() == "insufficient funds" {
			return nil, ErrInsufficientFunds
		}
		return nil, fmt.Errorf("failed to withdraw from balance: %w", err)
	}

	withdrawal, err := s.withdrawalRepo.Create(ctx, userID, req.OrderNumber, req.Sum)
	if err != nil {
		return nil, fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return withdrawal, nil
}

func (s *withdrawalService) GetWithdrawals(ctx context.Context, userID int) ([]*models.Withdrawal, error) {
	withdrawals, err := s.withdrawalRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	return withdrawals, nil
}
