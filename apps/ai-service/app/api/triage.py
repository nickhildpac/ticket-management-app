from __future__ import annotations

from fastapi import APIRouter

from app.ai.schemas import TicketContext, TriageResult
from app.ai.worker import build_agent
from app.core.config import get_settings

router = APIRouter(prefix="/triage", tags=["triage"])


@router.post("", response_model=TriageResult)
def triage(ticket: TicketContext) -> TriageResult:
    """On-demand triage for a single ticket.

    The primary path is the async worker consuming ticket events; this endpoint
    exists for manual re-runs and testing. It returns the decision without
    applying it back to the ticket service.
    """
    settings = get_settings()
    agent = build_agent(settings)
    return agent.triage(ticket)
