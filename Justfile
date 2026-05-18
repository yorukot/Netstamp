set dotenv-load := true

server_dir := "server"
web_filter := "@netstamp/web"
docs_filter := "@netstamp/docs"

alias dev := backend-dev
alias fmt := backend-fmt
alias tidy := backend-tidy
alias migrate-status := backend-migrate-status

# Misc

# List available recipes.
default:
    @just --list --unsorted

# Install workspace dependencies.
install:
    pnpm install

# Format frontend, docs, and shared files with Prettier.
format:
    pnpm format

# Build all runnable surfaces.
build: docs-build web-build backend-build

# Lint all available targets.
lint: web-lint backend-lint

# Run all available tests.
test: backend-test

# Remove local build and coverage artifacts.
clean:
    rm -rf docs/dist web/dist server/bin server/tmp server/coverage.out

# Documentation

# Start the documentation dev server.
docs-dev:
    pnpm --filter {{ docs_filter }} dev

# Build documentation.
docs-build:
    pnpm --filter {{ docs_filter }} build

# Start Storybook for shared UI components.
storybook-dev:
    pnpm --filter @netstamp/ui storybook

# Build static Storybook into docs public assets for Astro to copy.
storybook-build:
    pnpm --filter @netstamp/ui build:storybook -o ../../docs/public/storybook

# Preview the built documentation.
docs-preview:
    pnpm --filter {{ docs_filter }} preview

# Web

# Start the web dev server.
web-dev:
    pnpm --filter {{ web_filter }} dev

# Build the web app.
web-build:
    pnpm --filter {{ web_filter }} build

# Lint the web app.
web-lint:
    pnpm --filter {{ web_filter }} lint

# Preview the built web app.
web-preview:
    pnpm --filter {{ web_filter }} preview

# Backend

# Start the backend API server with hot reload.
backend-dev:
    cd {{ server_dir }} && air -c .air.toml

# Build the backend controller and probe agent binaries.
backend-build:
    mkdir -p {{ server_dir }}/bin
    cd {{ server_dir }} && go build -o bin/controller ./cmd/controller
    cd {{ server_dir }} && go build -o bin/agent ./cmd/agent

# Run the probe agent with an env file inside server/.
backend-probe env_file="probe.env": backend-build
    cd {{ server_dir }} && sh -c 'set -a && . "./{{ env_file }}" && set +a && ./bin/agent'

# Generate OpenAPI JSON for the docs explorer.
backend-openapi:
    cd {{ server_dir }} && go run ./cmd/openapi -output ../docs/public/openapi.json
    pnpm exec prettier --write docs/public/openapi.json

# Run backend tests.
backend-test:
    cd {{ server_dir }} && go test ./...

# Run backend API E2E tests against an explicit PostgreSQL URL.
backend-test-integration:
    cd {{ server_dir }} && go test -v -tags=integration ./internal/e2e/...

# Format backend code with golangci formatters.
backend-fmt:
    cd {{ server_dir }} && golangci-lint fmt --config ../golangci.yaml

# Run golangci-lint on backend code.
backend-lint:
    cd {{ server_dir }} && golangci-lint run --config ../golangci.yaml ./...

# Apply safe golangci-lint fixes.
backend-lint-fix:
    cd {{ server_dir }} && golangci-lint run --fix --config ../golangci.yaml ./...

# Tidy backend Go modules.
backend-tidy:
    cd {{ server_dir }} && go mod tidy
    cd {{ server_dir }}/tools && go mod tidy

# Generate SQLC code.
backend-sqlc:
    cd {{ server_dir }}/tools && go run github.com/sqlc-dev/sqlc/cmd/sqlc generate -f ../sqlc.yaml

# Show database migration status.
backend-migrate-status:
    cd {{ server_dir }} && go run ./cmd/migrate -command status

# Apply database migrations.
backend-migrate-up:
    cd {{ server_dir }} && go run ./cmd/migrate -command up

# Roll back the latest database migration.
backend-migrate-down:
    cd {{ server_dir }} && go run ./cmd/migrate -command down
