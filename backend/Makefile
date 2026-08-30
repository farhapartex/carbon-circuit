SERVICES := identity-service billing-service api-gateway
REVISION := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

.PHONY: tidy build test lint fmt up down logs ps reset env migrate-up migrate-down migrate-version migrate-force

tidy:
	go mod tidy

build:
	go build ./...

test:
	go test ./... -race -count=1

lint:
	go vet ./...

fmt:
	gofmt -l -w .

env:
	@for service in $(SERVICES); do \
		if [ ! -f services/$$service/.env ]; then \
			cp services/$$service/.env.example services/$$service/.env; \
			echo "created services/$$service/.env"; \
		fi; \
	done

up: env
	docker compose up -d --build

down:
	docker compose down

reset:
	docker compose down -v

logs:
	docker compose logs -f

ps:
	docker compose ps

SERVICE ?= identity-service
MIGRATION_DSN ?= postgres://$(SERVICE_ROLE):$(SERVICE_ROLE)@localhost:5432/carboncircuit?sslmode=disable
SERVICE_ROLE = $(subst -,_,$(subst -service,_service,$(SERVICE)))

migrate-up:
	DATABASE_MIGRATION_DSN="$(MIGRATION_DSN)" go run ./services/$(SERVICE)/cmd/migrate up

migrate-down:
	DATABASE_MIGRATION_DSN="$(MIGRATION_DSN)" go run ./services/$(SERVICE)/cmd/migrate down $(STEPS)

migrate-version:
	DATABASE_MIGRATION_DSN="$(MIGRATION_DSN)" go run ./services/$(SERVICE)/cmd/migrate version

migrate-force:
	DATABASE_MIGRATION_DSN="$(MIGRATION_DSN)" go run ./services/$(SERVICE)/cmd/migrate force $(VERSION)
