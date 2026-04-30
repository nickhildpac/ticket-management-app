from __future__ import annotations

from datetime import datetime
from typing import Annotated

from fastapi import Query
from pydantic import BaseModel, ConfigDict, Field, field_validator

from app.enums.priority import TicketPriority
from app.enums.ticket_state import TicketState, normalize_ticket_state_input
from app.schemas.common import SkillsMixin, UserInfo


class CreateTicketRequest(BaseModel, SkillsMixin):
    title: str = Field(min_length=1, max_length=200)
    description: str = Field(min_length=1)
    priority: TicketPriority = TicketPriority.LOW
    skills: list[str] = Field(default_factory=list)

    @field_validator("skills", mode="before")
    @classmethod
    def normalize_skills_input(cls, value: list[str] | None) -> list[str]:
        return cls.normalize_skills(value)


class UpdateTicketRequest(BaseModel, SkillsMixin):
    title: str | None = Field(default=None, min_length=1, max_length=200)
    description: str | None = Field(default=None, min_length=1)
    state: TicketState | None = None
    priority: TicketPriority | None = None
    assigned_to: list[str] | None = None
    skills: list[str] | None = None

    @field_validator("state", mode="before")
    @classmethod
    def normalize_state(cls, value: str | None) -> TicketState | None:
        if value is None:
            return None
        return normalize_ticket_state_input(value)

    @field_validator("skills", mode="before")
    @classmethod
    def normalize_skills_input(cls, value: list[str] | None) -> list[str] | None:
        if value is None:
            return None
        return cls.normalize_skills(value)


class TicketSummaryResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: str
    ticket_number: int
    created_by: str
    creator: UserInfo | None = None
    title: str
    description: str
    state: str
    priority: str
    created_at: datetime
    updated_at: datetime


class TicketResponse(TicketSummaryResponse):
    assigned_to: list[str]
    skills: list[str]


class TicketStatsResponse(BaseModel):
    total: int
    open: int
    pending: int
    resolved: int
    mine: int


class TicketListQuery(BaseModel):
    limit: int = 20
    offset: int = 0
    state: TicketState | None = None
    priority: TicketPriority | None = None
    created_by: str | None = None
    assignee: str | None = None
    assigned_to: str | None = None
    ticket_number: int | None = None
    sort: str = "created_at"
    order: str = "desc"

    @field_validator("state", mode="before")
    @classmethod
    def normalize_state(cls, value: str | None) -> TicketState | None:
        if value is None:
            return None
        return normalize_ticket_state_input(value)

    @field_validator("limit")
    @classmethod
    def validate_limit(cls, value: int) -> int:
        if value < 1 or value > 100:
            raise ValueError("limit must be between 1 and 100")
        return value

    @field_validator("offset")
    @classmethod
    def validate_offset(cls, value: int) -> int:
        if value < 0:
            raise ValueError("offset must be >= 0")
        return value

    @field_validator("sort")
    @classmethod
    def validate_sort(cls, value: str) -> str:
        normalized = value.strip().lower()
        if normalized not in {"ticket_number", "created_at"}:
            raise ValueError("sort must be ticket_number or created_at")
        return normalized

    @field_validator("order")
    @classmethod
    def validate_order(cls, value: str) -> str:
        normalized = value.strip().lower()
        if normalized not in {"asc", "desc"}:
            raise ValueError("order must be asc or desc")
        return normalized

    def effective_assignee(self) -> str | None:
        if self.assignee and self.assigned_to and self.assignee != self.assigned_to:
            raise ValueError("assignee and assigned_to must match when both are provided")
        return self.assignee or self.assigned_to


def ticket_list_query_params(
    limit: Annotated[int, Query()] = 20,
    offset: Annotated[int, Query()] = 0,
    state: Annotated[str | None, Query()] = None,
    priority: Annotated[str | None, Query()] = None,
    created_by: Annotated[str | None, Query()] = None,
    assignee: Annotated[str | None, Query()] = None,
    assigned_to: Annotated[str | None, Query()] = None,
    ticket_number: Annotated[int | None, Query()] = None,
    sort: Annotated[str, Query()] = "created_at",
    order: Annotated[str, Query()] = "desc",
) -> TicketListQuery:
    return TicketListQuery(
        limit=limit,
        offset=offset,
        state=state,
        priority=priority,
        created_by=created_by,
        assignee=assignee,
        assigned_to=assigned_to,
        ticket_number=ticket_number,
        sort=sort,
        order=order,
    )
