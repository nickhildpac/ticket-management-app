from __future__ import annotations

from fastapi import APIRouter

from app.api import triage

# The ticket lifecycle now lives in the Go ticket-service (ADR 0002). This
# service exposes only AI/RAG triage. The former ticket/comment/user/admin/auth
# routers under app/api/v1 are retained on disk for reference during migration
# but are no longer mounted.
api_router = APIRouter()

api_router.include_router(triage.router)
