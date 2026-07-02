from __future__ import annotations

from uuid import UUID

from sqlalchemy.orm import Session

from app.core.exceptions import BadRequestError, ForbiddenError, NotFoundError
from app.models.comment import Comment
from app.models.user import User
from app.policies.rbac import can_view_ticket
from app.repositories.comment import CommentRepository
from app.schemas.comment import CreateCommentRequest
from app.services.ticket import TicketService


class CommentService:
    def __init__(self, session: Session):
        self.session = session
        self.repo = CommentRepository(session)
        self.ticket_service = TicketService(session)

    def create_comment(self, current_user: User, payload: CreateCommentRequest) -> Comment:
        try:
            ticket_id = UUID(payload.ticket_id)
        except ValueError as exc:
            raise BadRequestError("invalid ticket_id") from exc
        ticket = self.ticket_service.get_ticket(current_user, ticket_id)

        if not can_view_ticket(current_user.role, current_user.id, ticket.created_by, ticket.assignee_ids):
            raise ForbiddenError("access denied")

        comment = Comment(
            ticket_id=ticket.id,
            created_by=current_user.id,
            description=payload.description,
        )
        self.repo.create(comment)
        self.session.commit()
        self.session.refresh(comment)
        return comment

    def get_comment(self, current_user: User, comment_id: UUID) -> Comment:
        comment = self.repo.get_by_id(comment_id)
        if not comment:
            raise NotFoundError("comment not found")
        ticket = self.ticket_service.get_ticket(current_user, comment.ticket_id)
        if not can_view_ticket(current_user.role, current_user.id, ticket.created_by, ticket.assignee_ids):
            raise ForbiddenError("access denied")
        return comment

    def list_ticket_comments(self, current_user: User, ticket_id: UUID) -> list[Comment]:
        _ = self.ticket_service.get_ticket(current_user, ticket_id)
        return self.repo.list_by_ticket(ticket_id)
