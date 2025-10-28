package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/RoGogDBD/loyalty_service/server/internal/config"
	"github.com/RoGogDBD/loyalty_service/server/internal/handler"
	"github.com/RoGogDBD/loyalty_service/server/internal/middleware"
	"github.com/RoGogDBD/loyalty_service/server/internal/repository"
	"github.com/RoGogDBD/loyalty_service/server/internal/service"
	"github.com/RoGogDBD/loyalty_service/server/internal/worker"
	"github.com/sirupsen/logrus"
)

type App struct {
	cfg           *config.Config
	logger        *logrus.Logger
	db            *repository.DB
	server        *http.Server
	workerContext context.Context
	workerCancel  context.CancelFunc
}

func New(cfg *config.Config, logger *logrus.Logger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Initialize() error {
	if err := a.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := a.initServer(); err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	return nil
}
func (a *App) initDatabase() error {
	dbCfg := &repository.DatabaseConfig{
		Host:            a.cfg.Database().Host,
		Port:            a.cfg.Database().Port,
		User:            a.cfg.Database().User,
		Password:        a.cfg.Database().Password,
		Database:        a.cfg.Database().Name,
		SSLMode:         "disable",
		MaxOpenConns:    a.cfg.Database().MaxOpenConns,
		MaxIdleConns:    a.cfg.Database().MaxIdleConns,
		ConnMaxLifetime: a.cfg.Database().ConnMaxLifetime,
		ConnMaxIdleTime: a.cfg.Database().ConnMaxIdleTime,
	}

	db, err := repository.NewDB(dbCfg, a.logger)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if err := db.RunMigrations(); err != nil {
		if cerr := db.Close(); cerr != nil {
			return fmt.Errorf("failed to close db after migration error: %v; original error: %w", cerr, err)
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	a.db = db
	return nil
}
func (a *App) initServer() error {
	// Repositories.
	userRepo := repository.NewUserRepository(a.db)
	orderRepo := repository.NewOrderRepository(a.db)
	balanceRepo := repository.NewBalanceRepository(a.db)
	withdrawalRepo := repository.NewWithdrawalRepository(a.db)

	// JWT Service.
	jwtService := service.NewJWTService(
		a.cfg.JWT().SecretKey,
		a.cfg.JWT().TokenDuration,
	)

	// Auth Service.
	authService := service.NewAuthService(userRepo)

	// Other Services.
	orderService := service.NewOrderService(orderRepo)
	balanceService := service.NewBalanceService(balanceRepo)
	withdrawalService := service.NewWithdrawalService(withdrawalRepo, balanceRepo)

	// Запускаем воркер, если указан адрес системы начисления.
	if a.cfg.Accrual().Address != "" {
		a.workerContext, a.workerCancel = context.WithCancel(context.Background())
		accrualWorker := worker.NewAccrualWorker(
			a.cfg.Accrual().Address,
			orderService,
			balanceService,
			a.logger,
		)
		go accrualWorker.Start(a.workerContext)
		a.logger.Info("Accrual worker initialized")
	} else {
		a.logger.Warn("Accrual system address not configured, worker not started")
	}

	// Handlers.
	authHandler := handler.NewAuthHandler(authService, jwtService, a.logger)
	orderHandler := handler.NewOrderHandler(orderService, a.logger)
	balanceHandler := handler.NewBalanceHandler(balanceService, a.logger)
	withdrawalHandler := handler.NewWithdrawalHandler(withdrawalService, a.logger)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	router := handler.NewRouter(authHandler, orderHandler, balanceHandler, withdrawalHandler, authMiddleware)

	a.server = &http.Server{
		Addr:         a.cfg.Server().Port,
		Handler:      router,
		ReadTimeout:  a.cfg.Server().ReadTimeout,
		WriteTimeout: a.cfg.Server().WriteTimeout,
	}

	return nil
}
func (a *App) Run() error {
	a.logger.Infof("Starting server on port %s", a.cfg.Server().Port)

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down server...")

	if a.workerCancel != nil {
		a.workerCancel()
	}

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Errorf("Server shutdown error: %v", err)
	}

	if err := a.db.Close(); err != nil {
		a.logger.Errorf("Database close error: %v", err)
	}

	return nil
}
