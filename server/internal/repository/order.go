package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, number string, userID int) (*models.Order, error)
	GetByNumber(ctx context.Context, number string) (*models.Order, error)
	GetByUserID(ctx context.Context, userID int) ([]*models.Order, error)
	UpdateStatus(ctx context.Context, number, status string, accrual *float64) error
	UpdateStatusByID(ctx context.Context, orderID int, status string, accrual float64) error
	GetPendingOrders(ctx context.Context) ([]*models.Order, error)
}

type orderRepository struct {
	db *DB
}

func NewOrderRepository(db *DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, number string, userID int) (*models.Order, error) {
	query := `
        INSERT INTO orders (number, user_id, status, uploaded_at)
        VALUES ($1, $2, $3, NOW())
        RETURNING id, number, user_id, status, accrual, uploaded_at
    `

	order := &models.Order{}
	var accrual sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, number, userID, models.OrderStatusNew).Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.Status,
		&accrual,
		&order.UploadedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if accrual.Valid {
		order.Accrual = &accrual.Float64
	}

	return order, nil
}

func (r *orderRepository) GetByNumber(ctx context.Context, number string) (*models.Order, error) {
	query := `
        SELECT id, number, user_id, status, accrual, uploaded_at
        FROM orders
        WHERE number = $1
    `

	order := &models.Order{}
	var accrual sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, number).Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.Status,
		&accrual,
		&order.UploadedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if accrual.Valid {
		order.Accrual = &accrual.Float64
	}

	return order, nil
}

func (r *orderRepository) GetByUserID(ctx context.Context, userID int) ([]*models.Order, error) {
	query := `
        SELECT id, number, user_id, status, accrual, uploaded_at
        FROM orders
        WHERE user_id = $1
        ORDER BY uploaded_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close rows: %w", closeErr)
		}
	}()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		var accrual sql.NullFloat64

		if err := rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if accrual.Valid {
			order.Accrual = &accrual.Float64
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, number, status string, accrual *float64) error {
	query := `
        UPDATE orders
        SET status = $1, accrual = $2
        WHERE number = $3
    `

	var accrualValue interface{}
	if accrual != nil {
		accrualValue = *accrual
	}

	_, err := r.db.ExecContext(ctx, query, status, accrualValue, number)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

func (r *orderRepository) UpdateStatusByID(ctx context.Context, orderID int, status string, accrual float64) error {
	query := `
        UPDATE orders
        SET status = $1, accrual = $2
        WHERE id = $3
    `

	_, err := r.db.ExecContext(ctx, query, status, accrual, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (r *orderRepository) GetPendingOrders(ctx context.Context) ([]*models.Order, error) {
	query := `
        SELECT id, number, user_id, status, accrual, uploaded_at
        FROM orders
        WHERE status IN ('NEW', 'PROCESSING')
        ORDER BY uploaded_at ASC
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close rows: %w", closeErr)
		}
	}()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		var accrual sql.NullFloat64

		if err := rows.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if accrual.Valid {
			order.Accrual = &accrual.Float64
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, nil
}
