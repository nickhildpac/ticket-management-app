from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, ConfigDict, EmailStr, Field, field_validator

from app.enums.role import UserRole
from app.schemas.common import AssignmentUserResponse, EmailMixin, SkillsMixin


class UserCreateRequest(BaseModel, EmailMixin, SkillsMixin):
    first_name: str = Field(min_length=1, max_length=100)
    last_name: str = Field(min_length=1, max_length=100)
    email: EmailStr
    password: str = Field(min_length=8, max_length=256)
    skills: list[str] = Field(default_factory=list)

    @field_validator("email", mode="before")
    @classmethod
    def normalize_email_input(cls, value: str) -> str:
        return cls.normalize_email(value)

    @field_validator("skills", mode="before")
    @classmethod
    def normalize_skills_input(cls, value: list[str] | None) -> list[str]:
        return cls.normalize_skills(value)


class UserResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: str
    first_name: str
    last_name: str
    email: EmailStr
    role: UserRole
    skills: list[str]
    created_at: datetime
    updated_at: datetime


class MePatchRequest(BaseModel, SkillsMixin):
    skills: list[str]

    @field_validator("skills", mode="before")
    @classmethod
    def normalize_skills_input(cls, value: list[str] | None) -> list[str]:
        return cls.normalize_skills(value)


class UpdateUserRoleRequest(BaseModel):
    role: UserRole


class AssignmentUsersResponse(BaseModel):
    items: list[AssignmentUserResponse]
