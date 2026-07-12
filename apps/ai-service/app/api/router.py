from __future__ import annotations

from fastapi import APIRouter

from app.api import triage

# The ticket lifecycle lives in the Go ticket-service (ADR 0002). This service
# exposes only AI/RAG triage. The former ticket/comment/user/admin/auth routers
# have been removed (they survive in git history if ever needed).
api_router = APIRouter()

api_router.include_router(triage.router)
