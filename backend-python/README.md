# FastAPI Backend (SPEC-1)

This module implements the Python/FastAPI backend defined in SPEC-1 while keeping the existing Go backend untouched.

## Run locally

Install [uv](https://docs.astral.sh/uv/) if it is not already on your PATH.

```bash
cd backend-python
uv sync --extra dev
cp .env.example .env
uv run alembic upgrade head
uv run uvicorn app.main:app --reload --port 8081
```

API base path: `http://localhost:8081/api/v1`

## Env vars

See `.env.example` for defaults.

## OpenAPI generation

```bash
python scripts/generate_openapi.py
```

This writes `openapi.json` in this folder.
