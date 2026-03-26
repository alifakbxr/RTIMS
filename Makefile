# RTIMS Makefile
# Common commands for development and deployment

.PHONY: help dev prod test clean build

# Default target
help:
	@echo "RTIMS - Real-Time Inventory Management System"
	@echo ""
	@echo "Development Commands:"
	@echo "  make dev              - Start all development services (Docker)"
	@echo "  make dev-backend      - Start backend only (Docker)"
	@echo "  make dev-frontend     - Start frontend only (Docker)"
	@echo "  make dev-db           - Start database and Redis only (Docker)"
	@echo "  make logs             - View all service logs"
	@echo "  make logs-backend     - View backend logs"
	@echo "  make logs-frontend    - View frontend logs"
	@echo ""
	@echo "  make backend-run      - Run backend locally (requires Go)"
	@echo "  make frontend-run     - Run frontend locally (requires Node.js)"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build            - Build all services"
	@echo "  make build-backend    - Build backend binary"
	@echo "  make build-frontend   - Build frontend production bundle"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test             - Run all tests"
	@echo "  make test-backend     - Run backend tests"
	@echo "  make test-frontend    - Run frontend tests"
	@echo ""
	@echo "Database Commands:"
	@echo "  make migrate          - Run database migrations"
	@echo "  make seed             - Seed database with sample data"
	@echo "  make db-backup        - Backup database"
	@echo "  make db-restore       - Restore database from backup"
	@echo "  make db-clean         - Clean database (WARNING: deletes all data)"
	@echo ""
	@echo "Production Commands:"
	@echo "  make prod             - Deploy to production"
	@echo "  make prod-stop        - Stop production services"
	@echo "  make prod-logs        - View production logs"
	@echo "  make prod-restart     - Restart production services"
	@echo ""
	@echo "Utility Commands:"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make clean-all        - Remove all generated files including data"
	@echo "  make setup            - Initial project setup"
	@echo "  make lint             - Run linters"
	@echo ""

# Development
dev:
	@echo "Starting development environment..."
	docker-compose up -d
	@echo "Services started:"
	@echo "  - Frontend: http://localhost:3000"
	@echo "  - Backend: http://localhost:8080"
	@echo "  - Swagger: http://localhost:8080/swagger/index.html"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"

dev-backend:
	docker-compose up -d backend

dev-frontend:
	docker-compose up -d frontend

dev-db:
	docker-compose up -d postgres redis

logs:
	docker-compose logs -f

logs-backend:
	docker-compose logs -f backend

logs-frontend:
	docker-compose logs -f frontend

# Local Development (without Docker)
backend-run:
	cd backend && go run main.go

frontend-run:
	cd frontend && npm run dev

# Build
build: build-backend build-frontend

build-backend:
	@echo "Building backend..."
	cd backend && go build -o bin/rtims-backend .
	@echo "Backend built: backend/bin/rtims-backend"

build-frontend:
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "Frontend built: frontend/.next"

# Test
test: test-backend test-frontend

test-backend:
	@echo "Running backend tests..."
	cd backend && go test ./... -v

test-frontend:
	@echo "Running frontend tests..."
	cd frontend && npm run test

# Database
migrate:
	@echo "Running database migrations..."
	docker-compose exec postgres psql -U rtims_user -d rtims -f /docker-entrypoint-initdb.d/001_initial_schema.sql

seed:
	@echo "Seeding database..."
	@echo "Database is automatically seeded on first run."

db-backup:
	@echo "Backing up database..."
	@mkdir -p backups
	docker exec rtims-postgres pg_dump -U rtims_user rtims > backups/backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "Backup created in backups/ directory"

db-restore:
	@echo "Restoring database..."
	@echo "Available backups:"
	@ls -1 backups/*.sql 2>/dev/null || echo "No backups found"
	@read -p "Enter backup filename: " backup; \
	docker exec -i rtims-postgres psql -U rtims_user rtims < backups/$$backup
	@echo "Database restored from backups/$$backup"

db-clean:
	@echo "WARNING: This will delete all database data!"
	@read -p "Are you sure? (y/N): " confirm; \
	if [ "$$confirm" = "y" ]; then \
		docker-compose down -v; \
		echo "Database volumes removed"; \
	else \
		echo "Operation cancelled"; \
	fi

# Production
prod:
	@echo "Deploying to production..."
	./deploy.sh

prod-stop:
	docker-compose -f docker-compose.production.yml down

prod-logs:
	docker-compose -f docker-compose.production.yml logs -f

prod-restart:
	docker-compose -f docker-compose.production.yml restart

# Utility
clean:
	@echo "Cleaning build artifacts..."
	rm -rf backend/bin
	rm -rf frontend/.next
	rm -rf frontend/node_modules
	@echo "Clean complete."

clean-all: clean
	@echo "Removing all generated files..."
	rm -rf backups/*
	docker-compose down -v
	@echo "Full clean complete. All data has been removed."

setup:
	@echo "Setting up RTIMS..."
	cp -n backend/.env.example backend/.env || true
	cp -n frontend/.env.example frontend/.env.local || true
	cp -n .env.production.example .env.production || true
	@echo "Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit backend/.env with your database credentials"
	@echo "2. Edit frontend/.env.local with your API URL"
	@echo "3. Run 'make dev' to start the development environment"

lint: lint-backend lint-frontend

lint-backend:
	@echo "Linting backend..."
	cd backend && golangci-lint run

lint-frontend:
	@echo "Linting frontend..."
	cd frontend && npm run lint

# Health Check
health:
	@echo "Checking service health..."
	@echo "Backend: $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/health)"
	@echo "Frontend: $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3000)"

# Stop all services
stop:
	docker-compose down
	@echo "All services stopped."

# Restart all services
restart:
	docker-compose restart
