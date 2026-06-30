from __future__ import annotations

from fastapi import APIRouter, Request, Response, status

from app.api.deps import DbSession, get_request_ip
from app.api.v1.serializers import to_user_auth_info, to_user_response
from app.core.config import get_settings
from app.core.exceptions import UnauthorizedError
from app.core.security import clear_refresh_cookie, set_refresh_cookie
from app.schemas.auth import AuthResponse, LoginRequest
from app.schemas.user import UserCreateRequest, UserResponse
from app.services.auth import AuthService

router = APIRouter(prefix="/auth", tags=["Authentication"])
registration_router = APIRouter(tags=["Users"])
settings = get_settings()


@registration_router.post("/users", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
def register_user(payload: UserCreateRequest, db: DbSession) -> UserResponse:
    user = AuthService(db).register_user(payload)
    return to_user_response(user)


@router.post("/login", response_model=AuthResponse)
def login(payload: LoginRequest, request: Request, response: Response, db: DbSession) -> AuthResponse:
    access_token, refresh_token, user = AuthService(db).login(
        email=payload.email,
        password=payload.password,
        user_agent=request.headers.get("user-agent"),
        ip_address=get_request_ip(request),
        refresh_expiry_hours=settings.refresh_token_expire_hours,
    )
    set_refresh_cookie(response, refresh_token)
    return AuthResponse(access_token=access_token, user=to_user_auth_info(user))


@router.get("/refresh", response_model=AuthResponse)
def refresh_token(request: Request, response: Response, db: DbSession) -> AuthResponse:
    raw_token = request.cookies.get(settings.refresh_cookie_name)
    if not raw_token:
        raise UnauthorizedError("missing refresh cookie")

    access_token, refresh_token, user = AuthService(db).refresh(
        raw_refresh_token=raw_token,
        user_agent=request.headers.get("user-agent"),
        ip_address=get_request_ip(request),
        refresh_expiry_hours=settings.refresh_token_expire_hours,
    )
    set_refresh_cookie(response, refresh_token)
    return AuthResponse(access_token=access_token, user=to_user_auth_info(user))


@router.get("/logout", status_code=status.HTTP_202_ACCEPTED)
def logout(request: Request, response: Response, db: DbSession) -> Response:
    raw_token = request.cookies.get(settings.refresh_cookie_name)
    AuthService(db).logout(raw_token)
    clear_refresh_cookie(response)
    response.status_code = status.HTTP_202_ACCEPTED
    return response
