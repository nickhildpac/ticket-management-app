from __future__ import annotations

from pydantic import BaseModel, ConfigDict, EmailStr

from app.enums.role import UserRole


class ErrorResponse(BaseModel):
    error: str


class UserInfo(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: str
    first_name: str
    last_name: str
    email: EmailStr


class AssignmentUserResponse(UserInfo):
    role: UserRole


class UserInfoAuth(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: str
    first_name: str
    last_name: str
    email: EmailStr
    role: UserRole


class SkillsMixin:
    @staticmethod
    def normalize_skills(skills: list[str] | None) -> list[str]:
        if not skills:
            return []
        seen: set[str] = set()
        normalized: list[str] = []
        for skill in skills:
            item = skill.strip().lower()
            if not item:
                continue
            if len(item) > 64:
                raise ValueError("skills entries must be 64 chars or fewer")
            if item in seen:
                continue
            seen.add(item)
            normalized.append(item)
        return normalized


class EmailMixin:
    @staticmethod
    def normalize_email(value: str) -> str:
        return value.strip().lower()
