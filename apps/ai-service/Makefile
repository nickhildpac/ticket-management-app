.PHONY: run test migrate-up migrate-down openapi ticket-state-contract

run:
	uvicorn app.main:app --reload --port 8081

test:
	pytest

migrate-up:
	alembic upgrade head

migrate-down:
	alembic downgrade -1

openapi:
	python scripts/generate_openapi.py

ticket-state-contract:
	cd ../.. && python3 tools/generate_ticket_state_contract.py
