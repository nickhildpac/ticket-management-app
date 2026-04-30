from __future__ import annotations

from uuid import UUID

from fastapi import APIRouter, status

from app.api.deps import CurrentUser, DbSession
from app.api.v1.serializers import to_comment_response
from app.schemas.comment import CommentResponse, CreateCommentRequest
from app.services.comment import CommentService

router = APIRouter(tags=["Comments"])


@router.post("/comment", response_model=CommentResponse, status_code=status.HTTP_202_ACCEPTED)
def create_comment(payload: CreateCommentRequest, current_user: CurrentUser, db: DbSession) -> CommentResponse:
    comment = CommentService(db).create_comment(current_user, payload)
    return to_comment_response(comment)


@router.get("/comment/{comment_id}", response_model=CommentResponse)
def get_comment(comment_id: UUID, current_user: CurrentUser, db: DbSession) -> CommentResponse:
    comment = CommentService(db).get_comment(current_user, comment_id)
    return to_comment_response(comment)


@router.get("/ticket/{ticket_id}/comments", response_model=list[CommentResponse])
def list_ticket_comments(ticket_id: UUID, current_user: CurrentUser, db: DbSession) -> list[CommentResponse]:
    comments = CommentService(db).list_ticket_comments(current_user, ticket_id)
    return [to_comment_response(comment) for comment in comments]
