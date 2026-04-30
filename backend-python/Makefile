.PHONY: run test migrate-up migrate-down openapi

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
