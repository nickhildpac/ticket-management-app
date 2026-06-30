from __future__ import annotations

from uuid import UUID

from fastapi import APIRouter, Depends, Response, status

from app.api.deps import CurrentUser, DbSession
from app.api.v1.serializers import to_ticket_response, to_ticket_summary
from app.schemas.ticket import (
    CreateTicketRequest,
    TicketListQuery,
    TicketResponse,
    TicketStatsResponse,
    TicketSummaryResponse,
    UpdateTicketRequest,
    ticket_list_query_params,
)
from app.services.ticket import TicketService

router = APIRouter(prefix="/tickets", tags=["Tickets"])


@router.get("/stats", response_model=TicketStatsResponse)
def get_ticket_stats(current_user: CurrentUser, db: DbSession) -> TicketStatsResponse:
    stats = TicketService(db).get_stats(current_user)
    return TicketStatsResponse(**stats)


@router.get("", response_model=list[TicketSummaryResponse])
def list_tickets(
    current_user: CurrentUser,
    db: DbSession,
    query: TicketListQuery = Depends(ticket_list_query_params),
) -> list[TicketSummaryResponse]:
    if query.assigned_to == "me" and query.assignee in (None, "me"):
        query.assigned_to = None
        query.assignee = None
        tickets = TicketService(db).list_assigned_tickets(current_user, query)
        return [to_ticket_summary(ticket) for ticket in tickets]

    tickets = TicketService(db).list_tickets_for_actor(current_user, query)
    return [to_ticket_summary(ticket) for ticket in tickets]


@router.post("", response_model=TicketSummaryResponse, status_code=status.HTTP_202_ACCEPTED)
def create_ticket(payload: CreateTicketRequest, current_user: CurrentUser, db: DbSession) -> TicketSummaryResponse:
    ticket = TicketService(db).create_ticket(current_user, payload)
    return to_ticket_summary(ticket)


@router.get("/number/{number}", response_model=TicketResponse)
def get_ticket_by_number(number: int, current_user: CurrentUser, db: DbSession) -> TicketResponse:
    ticket = TicketService(db).get_ticket_by_number(current_user, number)
    return to_ticket_response(ticket)


@router.get("/{ticket_id}", response_model=TicketResponse)
def get_ticket(ticket_id: UUID, current_user: CurrentUser, db: DbSession) -> TicketResponse:
    ticket = TicketService(db).get_ticket(current_user, ticket_id)
    return to_ticket_response(ticket)


@router.patch("/{ticket_id}", response_model=TicketSummaryResponse)
def update_ticket(
    ticket_id: UUID,
    payload: UpdateTicketRequest,
    current_user: CurrentUser,
    db: DbSession,
) -> TicketSummaryResponse:
    ticket = TicketService(db).update_ticket(current_user, ticket_id, payload)
    return to_ticket_summary(ticket)


@router.delete("/{ticket_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_ticket(ticket_id: UUID, current_user: CurrentUser, db: DbSession) -> Response:
    TicketService(db).delete_ticket(current_user, ticket_id)
    return Response(status_code=status.HTTP_204_NO_CONTENT)
