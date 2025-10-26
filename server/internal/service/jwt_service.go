package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type (
	JWTService interface {
		GenerateToken(userID int) (string, error)
		ValidateToken(tokenString string) (int, error)
	}

	jwtService struct {
		secretKey string
		duration  time.Duration
	}

	Claims struct {
		UserID int `json:"user_id"`
		jwt.RegisteredClaims
	}
)

func NewJWTService(secretKey string, duration time.Duration) JWTService {
	return &jwtService{
		secretKey: secretKey,
		duration:  duration,
	}
}

func (s *jwtService) GenerateToken(userID int) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *jwtService) ValidateToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, fmt.Errorf("invalid token")
}
