from __future__ import annotations

from fastapi import APIRouter

from app.api.deps import CurrentUser, DbSession
from app.api.v1.serializers import to_assignment_user, to_user_response
from app.schemas.common import AssignmentUserResponse
from app.schemas.user import MePatchRequest, UserResponse
from app.services.user import UserService

router = APIRouter(tags=["Users"])


@router.get("/me", response_model=UserResponse)
def get_me(current_user: CurrentUser) -> UserResponse:
    return to_user_response(current_user)


@router.patch("/me", response_model=UserResponse)
def patch_me(payload: MePatchRequest, current_user: CurrentUser, db: DbSession) -> UserResponse:
    user = UserService(db).patch_my_skills(current_user, payload.skills)
    return to_user_response(user)


@router.get("/users", response_model=list[AssignmentUserResponse])
def list_assignment_candidates(current_user: CurrentUser, db: DbSession) -> list[AssignmentUserResponse]:
    _ = current_user
    users = UserService(db).list_assignment_candidates()
    return [to_assignment_user(user) for user in users]
