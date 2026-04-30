from __future__ import annotations

from fastapi import APIRouter

from app.api.v1 import admin, auth, comments, tickets, users

api_router = APIRouter()

api_router.include_router(auth.router)
api_router.include_router(users.router)
api_router.include_router(tickets.router)
api_router.include_router(comments.router)
api_router.include_router(admin.router)
