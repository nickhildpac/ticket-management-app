from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, ConfigDict, Field

from app.schemas.common import UserInfo


class CreateCommentRequest(BaseModel):
    ticket_id: str
    description: str = Field(min_length=1)


class CommentResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: str
    ticket_id: str
    created_by: str
    creator: UserInfo
    description: str
    created_at: datetime
