from __future__ import annotations

from pydantic import BaseModel, EmailStr, field_validator

from app.schemas.common import UserInfoAuth


class LoginRequest(BaseModel):
    email: EmailStr
    password: str

    @field_validator("email", mode="before")
    @classmethod
    def normalize_email(cls, value: str) -> str:
        return value.strip().lower()


class AuthResponse(BaseModel):
    access_token: str
    user: UserInfoAuth
