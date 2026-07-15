# Single entrypoint for the monorepo. Every workflow is a `make <verb>` that
# fans out to the per-app toolchains. See docs/adr/0002-service-topology.md.

# --env-file .env: compose resolves project dir from the -f path (infra/compose/),
# so without this the repo-root .env is ignored (ANTHROPIC/OPENAI keys stay empty).
COMPOSE := docker compose --env-file .env -f infra/compose/docker-compose.yml
TICKET  := apps/ticket-service
AI      := apps/ai-service
WEB     := apps/web
KB_PATH ?= knowledge

.DEFAULT_GOAL := help

.PHONY: help
help: ## List available targets
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "} {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

# ---- setup --------------------------------------------------------------
.PHONY: setup
setup: ## Install dependencies for all apps
	cd $(TICKET) && go mod download
	cd $(AI) && uv sync --extra dev
	cd $(WEB) && bun install

# ---- run (compose profiles) --------------------------------------------
.PHONY: up
up: ## Start core stack (postgres + ticket-service + web)
	$(COMPOSE) up -d --build

.PHONY: up-ai
up-ai: ## Start core + AI stack (redis + ai-service)
	$(COMPOSE) --profile ai up -d --build

.PHONY: up-obs
up-obs: ## Start core + AI + observability (prometheus + grafana)
	$(COMPOSE) --profile ai --profile obs up -d --build

.PHONY: down
down: ## Stop all services
	$(COMPOSE) --profile ai --profile obs down

.PHONY: logs
logs: ## Tail all service logs
	$(COMPOSE) --profile ai --profile obs logs -f

# ---- quality ------------------------------------------------------------
.PHONY: test
test: test-ticket test-ai test-web ## Run all test suites

.PHONY: test-ticket
test-ticket: ## Run Go ticket-service tests
	cd $(TICKET) && go test ./...

.PHONY: test-ai
test-ai: ## Run Python ai-service tests
	cd $(AI) && uv run pytest

.PHONY: test-web
test-web: ## Type-check + build the web app
	cd $(WEB) && bun run build

.PHONY: lint
lint: ## Lint all apps
	cd $(TICKET) && go vet ./...
	cd $(AI) && uv run ruff check .
	cd $(WEB) && bun run lint

# ---- contracts ----------------------------------------------------------
.PHONY: contracts
contracts: ## Regenerate all generated contract artifacts from contracts/
	python3 tools/generate_ticket_state_contract.py

# ---- database -----------------------------------------------------------
# The Go ticket-service owns the ticket/user/comment schema and applies its own
# migrations automatically on startup (AUTO_MIGRATE). This target applies the
# AI-service's Alembic migrations (kb_chunks and other AI-owned tables).
.PHONY: migrate
migrate: ## Apply AI-service Alembic migrations (Go schema auto-migrates on startup)
	$(COMPOSE) up -d postgres
	cd $(AI) && uv run alembic upgrade head

# ---- knowledge base -----------------------------------------------------
# knowledge/ is in .dockerignore, so compose ingest bind-mounts the host tree.
.PHONY: ingest
ingest: ## Ingest KB into pgvector via ai-service container (KB_PATH=knowledge)
	$(COMPOSE) --profile ai run --rm --no-deps \
		-v "$(CURDIR)/$(AI)/$(KB_PATH):/app/knowledge:ro" \
		ai-service \
		uv run python -m app.ai.ingest /app/knowledge

.PHONY: worker
worker: ## Run the AI triage worker (Redis Streams consumer)
	cd $(AI) && uv run python -m app.ai.worker
