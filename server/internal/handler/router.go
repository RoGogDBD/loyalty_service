package handler

import (
	"github.com/RoGogDBD/loyalty_service/server/internal/middleware"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
)

func NewRouter(
	authHandler *AuthHandler,
	orderHandler *OrderHandler,
	balanceHandler *BalanceHandler,
	withdrawalHandler *WithdrawalHandler,
	authMiddleware *middleware.AuthMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Post("/orders", orderHandler.UploadOrder)
			r.Get("/orders", orderHandler.GetOrders)
			r.Get("/balance", balanceHandler.GetBalance)
			r.Post("/balance/withdraw", withdrawalHandler.Withdraw)
			r.Get("/withdrawals", withdrawalHandler.GetWithdrawals)
		})
	})

	return r
}
