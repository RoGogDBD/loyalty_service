.PHONY: help build run migrate test docker-up docker-down docker-clean lint

# Переменные
APP_NAME := gophermart
DOCKER_COMPOSE := docker compose

help: ## Показать список команд
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать приложение
	@echo "Building $(APP_NAME)..."
	@cd server && go build -o ../bin/$(APP_NAME) ./cmd/server

docker-build: ## Собрать Docker образы
	@echo "Building Docker images..."
	@$(DOCKER_COMPOSE) build

run: docker-up

docker-up: ## Запустить Docker Compose
	@echo "Starting Docker containers..."
	@$(DOCKER_COMPOSE) up -d

docker-down: ## Остановить Docker Compose
	@echo "Stopping Docker containers..."
	@$(DOCKER_COMPOSE) down

docker-clean: ## Полностью удалить Docker образы и volumes
	@echo "Cleaning up Docker..."
	@$(DOCKER_COMPOSE) down -v --rmi all --remove-orphans
	@docker volume prune -f

docker-logs: ## Показать логи контейнеров
	@$(DOCKER_COMPOSE) logs -f

.DEFAULT_GOAL := help