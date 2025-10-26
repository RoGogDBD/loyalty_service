package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
	"github.com/RoGogDBD/loyalty_service/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService interface {
	Register(ctx context.Context, creds *models.UserCredentials) (*models.User, error)
	Login(ctx context.Context, creds *models.UserCredentials) (*models.User, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Register(ctx context.Context, creds *models.UserCredentials) (*models.User, error) {
	existing, err := s.userRepo.GetByLogin(ctx, creds.Login)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, creds.Login, string(hash))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, creds *models.UserCredentials) (*models.User, error) {
	user, err := s.userRepo.GetByLogin(ctx, creds.Login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
