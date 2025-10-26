package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
)

type WithdrawalRepository interface {
	Create(ctx context.Context, userID int, orderNumber string, sum float64) (*models.Withdrawal, error)
	GetByUserID(ctx context.Context, userID int) ([]*models.Withdrawal, error)
}

type withdrawalRepository struct {
	db *DB
}

func NewWithdrawalRepository(db *DB) WithdrawalRepository {
	return &withdrawalRepository{db: db}
}

func (r *withdrawalRepository) Create(ctx context.Context, userID int, orderNumber string, sum float64) (*models.Withdrawal, error) {
	query := `
        INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
        VALUES ($1, $2, $3, NOW())
        RETURNING id, user_id, order_number, sum, processed_at
    `

	withdrawal := &models.Withdrawal{}
	err := r.db.QueryRowContext(ctx, query, userID, orderNumber, sum).Scan(
		&withdrawal.ID,
		&withdrawal.UserID,
		&withdrawal.OrderNumber,
		&withdrawal.Sum,
		&withdrawal.ProcessedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return withdrawal, nil
}

func (r *withdrawalRepository) GetByUserID(ctx context.Context, userID int) ([]*models.Withdrawal, error) {
	query := `
        SELECT id, user_id, order_number, sum, processed_at
        FROM withdrawals
        WHERE user_id = $1
        ORDER BY processed_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var withdrawals []*models.Withdrawal
	for rows.Next() {
		withdrawal := &models.Withdrawal{}
		if err := rows.Scan(
			&withdrawal.ID,
			&withdrawal.UserID,
			&withdrawal.OrderNumber,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	return withdrawals, nil
}
