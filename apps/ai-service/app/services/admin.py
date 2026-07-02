from __future__ import annotations

from uuid import UUID

from sqlalchemy.orm import Session

from app.enums.role import UserRole
from app.models.user import User
from app.services.user import UserService


class AdminService:
    def __init__(self, session: Session):
        self.user_service = UserService(session)

    def list_users(self, current_user: User):
        return self.user_service.list_all_users(current_user)

    def update_user_role(self, current_user: User, target_user_id: UUID, role: UserRole):
        return self.user_service.update_user_role(current_user, target_user_id, role)

    def delete_user(self, current_user: User, target_user_id: UUID) -> None:
        self.user_service.delete_user(current_user, target_user_id)
