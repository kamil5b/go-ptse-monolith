# Makefile - helpers for migrations

.PHONY: install-goose migrate-up migrate-down

install-goose:
	@echo "Installing goose (pressly/goose)..."
	go install github.com/pressly/goose/v3/cmd/goose@latest

# Run Postgres migrations (goose folder)
migrate-up:
	if [ -z "$(DATABASE_URL)" ]; then echo "Set DATABASE_URL env var"; exit 1; fi
	goose -dir migrations/goose postgres "$(DATABASE_URL)" up

migrate-down:
	if [ -z "$(DATABASE_URL)" ]; then echo "Set DATABASE_URL env var"; exit 1; fi
	goose -dir migrations/goose postgres "$(DATABASE_URL)" down
