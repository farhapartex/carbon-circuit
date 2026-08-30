SERVICES := identity-service
REVISION := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

.PHONY: tidy build test lint fmt up down logs ps reset env

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
