from __future__ import annotations

from app.enums.ticket_state import ticket_state_to_wire
from app.models.comment import Comment
from app.models.ticket import Ticket
from app.models.user import User
from app.schemas.comment import CommentResponse
from app.schemas.common import AssignmentUserResponse, UserInfo, UserInfoAuth
from app.schemas.ticket import TicketResponse, TicketSummaryResponse
from app.schemas.user import UserResponse


def to_user_info(user: User) -> UserInfo:
    return UserInfo(
        id=str(user.id),
        first_name=user.first_name,
        last_name=user.last_name,
        email=user.email,
    )


def to_user_auth_info(user: User) -> UserInfoAuth:
    return UserInfoAuth(
        id=str(user.id),
        first_name=user.first_name,
        last_name=user.last_name,
        email=user.email,
        role=user.role,
    )


def to_assignment_user(user: User) -> AssignmentUserResponse:
    return AssignmentUserResponse(
        id=str(user.id),
        first_name=user.first_name,
        last_name=user.last_name,
        email=user.email,
        role=user.role,
    )


def to_user_response(user: User) -> UserResponse:
    return UserResponse(
        id=str(user.id),
        first_name=user.first_name,
        last_name=user.last_name,
        email=user.email,
        role=user.role,
        skills=user.skill_names,
        created_at=user.created_at,
        updated_at=user.updated_at,
    )


def to_ticket_summary(ticket: Ticket) -> TicketSummaryResponse:
    creator = to_user_info(ticket.creator) if ticket.creator else None
    return TicketSummaryResponse(
        id=str(ticket.id),
        ticket_number=ticket.ticket_number,
        created_by=str(ticket.created_by),
        creator=creator,
        title=ticket.title,
        description=ticket.description,
        state=ticket_state_to_wire(ticket.state),
        priority=ticket.priority.value,
        created_at=ticket.created_at,
        updated_at=ticket.updated_at,
    )


def to_ticket_response(ticket: Ticket) -> TicketResponse:
    summary = to_ticket_summary(ticket)
    return TicketResponse(
        **summary.model_dump(),
        assigned_to=[str(user_id) for user_id in ticket.assignee_ids],
        skills=ticket.skill_names,
    )


def to_comment_response(comment: Comment) -> CommentResponse:
    creator = to_user_info(comment.creator)
    return CommentResponse(
        id=str(comment.id),
        ticket_id=str(comment.ticket_id),
        created_by=str(comment.created_by),
        creator=creator,
        description=comment.description,
        created_at=comment.created_at,
    )
