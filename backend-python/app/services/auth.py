from __future__ import annotations

from datetime import timedelta

from sqlalchemy.orm import Session

from app.core.exceptions import ConflictError, UnauthorizedError
from app.core.security import (
    create_access_token,
    generate_refresh_token,
    hash_password,
    hash_refresh_token,
    now_utc,
    verify_password,
)
from app.models.refresh_token import RefreshToken
from app.models.user import User
from app.repositories.refresh_token import RefreshTokenRepository
from app.repositories.user import UserRepository
from app.schemas.user import UserCreateRequest


class AuthService:
    def __init__(self, session: Session):
        self.session = session
        self.user_repo = UserRepository(session)
        self.refresh_repo = RefreshTokenRepository(session)

    def register_user(self, payload: UserCreateRequest) -> User:
        existing = self.user_repo.get_by_email(payload.email)
        if existing:
            raise ConflictError("email already exists")

        user = User(
            first_name=payload.first_name,
            last_name=payload.last_name,
            email=payload.email,
            password_hash=hash_password(payload.password),
        )
        self.user_repo.create(user)
        self.user_repo.set_skills(user, payload.skills)

        self.session.commit()
        self.session.refresh(user)
        return user

    def login(
        self,
        *,
        email: str,
        password: str,
        user_agent: str | None,
        ip_address: str | None,
        refresh_expiry_hours: int,
    ) -> tuple[str, str, User]:
        user = self.user_repo.get_by_email(email)
        if not user or not verify_password(password, user.password_hash):
            raise UnauthorizedError("invalid credentials")
        if not user.is_active:
            raise UnauthorizedError("inactive user")

        access_token = create_access_token(subject=str(user.id), role=user.role.value)

        raw_refresh_token = generate_refresh_token()
        refresh_token = RefreshToken(
            user_id=user.id,
            token_hash=hash_refresh_token(raw_refresh_token),
            expires_at=now_utc() + timedelta(hours=refresh_expiry_hours),
            user_agent=user_agent,
            ip_address=ip_address,
        )
        self.refresh_repo.create(refresh_token)
        self.session.commit()

        return access_token, raw_refresh_token, user

    def refresh(
        self,
        *,
        raw_refresh_token: str,
        user_agent: str | None,
        ip_address: str | None,
        refresh_expiry_hours: int,
    ) -> tuple[str, str, User]:
        token_hash = hash_refresh_token(raw_refresh_token)
        current_token = self.refresh_repo.get_active_by_hash(token_hash, now_utc())
        if not current_token:
            raise UnauthorizedError("invalid refresh token")

        user = self.user_repo.get_by_id(current_token.user_id)
        if not user or not user.is_active:
            raise UnauthorizedError("invalid refresh token")

        current_token.revoked_at = now_utc()

        new_raw_token = generate_refresh_token()
        new_token = RefreshToken(
            user_id=user.id,
            token_hash=hash_refresh_token(new_raw_token),
            expires_at=now_utc() + timedelta(hours=refresh_expiry_hours),
            user_agent=user_agent,
            ip_address=ip_address,
        )
        self.refresh_repo.create(new_token)

        access_token = create_access_token(subject=str(user.id), role=user.role.value)
        self.session.commit()

        return access_token, new_raw_token, user

    def logout(self, raw_refresh_token: str | None) -> None:
        if not raw_refresh_token:
            return

        token_hash = hash_refresh_token(raw_refresh_token)
        token = self.refresh_repo.get_active_by_hash(token_hash, now_utc())
        if token:
            token.revoked_at = now_utc()
            self.session.commit()
