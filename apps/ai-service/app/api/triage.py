from __future__ import annotations

from functools import lru_cache

import jwt
from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer

from app.ai.agent import TriageAgent
from app.ai.schemas import TicketContext, TriageResult
from app.ai.worker import build_agent
from app.core.config import Settings, get_settings

_bearer = HTTPBearer(auto_error=True)


@lru_cache
def _agent() -> TriageAgent:
    """Process-singleton agent. Building it constructs a SQLAlchemy engine (which
    owns a connection pool) and an Anthropic client, so it must not be rebuilt
    per request."""
    return build_agent(get_settings())


def require_auth(
    credentials: HTTPAuthorizationCredentials = Depends(_bearer),
    settings: Settings = Depends(get_settings),
) -> None:
    """Require a valid ticket-service JWT (shared HS256 secret). Prevents anyone
    who can reach the port from triggering paid model calls."""
    try:
        jwt.decode(
            credentials.credentials,
            settings.jwt_secret,
            algorithms=[settings.jwt_algorithm],
            audience=settings.jwt_audience,
            issuer=settings.jwt_issuer,
        )
    except jwt.PyJWTError as exc:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="invalid token"
        ) from exc


router = APIRouter(prefix="/triage", tags=["triage"], dependencies=[Depends(require_auth)])


@router.post("", response_model=TriageResult)
def triage(ticket: TicketContext) -> TriageResult:
    """On-demand triage for a single ticket.

    The primary path is the async worker consuming ticket events; this endpoint
    exists for manual re-runs and testing. It returns the decision without
    applying it back to the ticket service.
    """
    return _agent().triage(ticket)
