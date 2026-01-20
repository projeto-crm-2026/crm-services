help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

up: ## Start all services with docker-compose
	docker-compose up -d

down: ## Stop all services
	docker-compose down

build: ## Build docker images
	docker-compose build

restart: ## Restart all services
	docker-compose restart

logs: ## Show logs from all services
	docker-compose logs -f

clean: ## Stop containers and remove volumes
	docker-compose down -v

clean-all: ## Stop containers, remove volumes, and remove images
	docker-compose down -v --rmi all

dev: ## Start services and watch logs
	docker-compose up

ps: ## Show running containers
	docker-compose ps