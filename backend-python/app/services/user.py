from __future__ import annotations

from uuid import UUID

from sqlalchemy.orm import Session

from app.core.exceptions import BadRequestError, ForbiddenError, NotFoundError
from app.enums.role import UserRole
from app.models.user import User
from app.policies.rbac import can_manage_users
from app.repositories.user import UserRepository


class UserService:
    def __init__(self, session: Session):
        self.session = session
        self.repo = UserRepository(session)

    def get_user_or_404(self, user_id: UUID) -> User:
        user = self.repo.get_by_id(user_id)
        if not user:
            raise NotFoundError("user not found")
        return user

    def patch_my_skills(self, current_user: User, skills: list[str]) -> User:
        self.repo.set_skills(current_user, skills)
        self.session.commit()
        self.session.refresh(current_user)
        return current_user

    def list_assignment_candidates(self) -> list[User]:
        return self.repo.list_assignment_candidates()

    def list_all_users(self, current_user: User) -> list[User]:
        if not can_manage_users(current_user.role):
            raise ForbiddenError("access denied")
        return self.repo.list_all()

    def update_user_role(self, current_user: User, target_user_id: UUID, role: UserRole) -> User:
        if not can_manage_users(current_user.role):
            raise ForbiddenError("access denied")

        user = self.get_user_or_404(target_user_id)
        user.role = role
        self.session.commit()
        self.session.refresh(user)
        return user

    def delete_user(self, current_user: User, target_user_id: UUID) -> None:
        if not can_manage_users(current_user.role):
            raise ForbiddenError("access denied")
        if current_user.id == target_user_id:
            raise BadRequestError("cannot delete your own account")

        user = self.get_user_or_404(target_user_id)
        self.repo.delete(user)
        self.session.commit()
