package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, login, passwordHash string) (*models.User, error)
	GetByLogin(ctx context.Context, login string) (*models.User, error)
	GetByID(ctx context.Context, id int) (*models.User, error)
}

type userRepository struct {
	db *DB
}

func NewUserRepository(db *DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, login, passwordHash string) (*models.User, error) {
	query := `
        INSERT INTO users (login, password_hash, created_at)
        VALUES ($1, $2, NOW())
        RETURNING id, login, created_at
    `

	user := &models.User{PasswordHash: passwordHash}
	err := r.db.QueryRowContext(ctx, query, login, passwordHash).Scan(
		&user.ID,
		&user.Login,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *userRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `
        SELECT id, login, password_hash, created_at
        FROM users
        WHERE login = $1
    `

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	return user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
        SELECT id, login, password_hash, created_at
        FROM users
        WHERE id = $1
    `

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}
