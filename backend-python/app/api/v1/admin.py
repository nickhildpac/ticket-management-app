from __future__ import annotations

from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, Response, status

from app.api.deps import DbSession, require_roles
from app.api.v1.serializers import to_user_response
from app.enums.role import UserRole
from app.models.user import User
from app.schemas.user import UpdateUserRoleRequest, UserResponse
from app.services.admin import AdminService

router = APIRouter(prefix="/admin", tags=["Admin"])
AdminUser = Annotated[User, Depends(require_roles(UserRole.ADMIN))]


@router.get("/users", response_model=list[UserResponse])
def get_all_users(current_user: AdminUser, db: DbSession) -> list[UserResponse]:
    users = AdminService(db).list_users(current_user)
    return [to_user_response(user) for user in users]


@router.put("/users/{id}/role", response_model=UserResponse)
def update_user_role(
    id: UUID,
    payload: UpdateUserRoleRequest,
    current_user: AdminUser,
    db: DbSession,
) -> UserResponse:
    user = AdminService(db).update_user_role(current_user, id, payload.role)
    return to_user_response(user)


@router.delete("/users/{id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_user(id: UUID, current_user: AdminUser, db: DbSession) -> Response:
    AdminService(db).delete_user(current_user, id)
    return Response(status_code=status.HTTP_204_NO_CONTENT)
