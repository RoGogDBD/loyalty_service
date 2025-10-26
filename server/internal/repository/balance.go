package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
)

type BalanceRepository interface {
	GetByUserID(ctx context.Context, userID int) (*models.Balance, error)
	Create(ctx context.Context, userID int) error
	AddAccrual(ctx context.Context, userID int, amount float64) error
	Withdraw(ctx context.Context, userID int, amount float64) error
}

type balanceRepository struct {
	db *DB
}

func NewBalanceRepository(db *DB) BalanceRepository {
	return &balanceRepository{db: db}
}

func (r *balanceRepository) GetByUserID(ctx context.Context, userID int) (*models.Balance, error) {
	query := `
        SELECT user_id, current, withdrawn
        FROM balance
        WHERE user_id = $1
    `

	balance := &models.Balance{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&balance.UserID,
		&balance.Current,
		&balance.Withdrawn,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return &models.Balance{UserID: userID, Current: 0, Withdrawn: 0}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (r *balanceRepository) Create(ctx context.Context, userID int) error {
	query := `
        INSERT INTO balance (user_id, current, withdrawn)
        VALUES ($1, 0, 0)
        ON CONFLICT (user_id) DO NOTHING
    `

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to create balance: %w", err)
	}

	return nil
}

func (r *balanceRepository) AddAccrual(ctx context.Context, userID int, amount float64) error {
	query := `
        INSERT INTO balance (user_id, current, withdrawn)
        VALUES ($1, $2, 0)
        ON CONFLICT (user_id)
        DO UPDATE SET current = balance.current + EXCLUDED.current
    `

	_, err := r.db.ExecContext(ctx, query, userID, amount)
	if err != nil {
		return fmt.Errorf("failed to add accrual: %w", err)
	}

	return nil
}

func (r *balanceRepository) Withdraw(ctx context.Context, userID int, amount float64) error {
	query := `
        UPDATE balance
        SET current = current - $1, withdrawn = withdrawn + $1
        WHERE user_id = $2 AND current >= $1
    `

	result, err := r.db.ExecContext(ctx, query, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("insufficient funds")
	}

	return nil
}
