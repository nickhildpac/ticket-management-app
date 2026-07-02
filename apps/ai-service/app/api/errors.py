from __future__ import annotations

import logging

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

from app.core.exceptions import AppError

logger = logging.getLogger(__name__)


def register_exception_handlers(app: FastAPI) -> None:
    @app.exception_handler(AppError)
    async def app_error_handler(_: Request, exc: AppError) -> JSONResponse:
        return JSONResponse(
            status_code=exc.status_code,
            content={"code": exc.code, "message": exc.message, "details": exc.details},
        )

    @app.exception_handler(RequestValidationError)
    async def validation_error_handler(_: Request, exc: RequestValidationError) -> JSONResponse:
        details = [
            {
                "field": ".".join(str(part) for part in err["loc"]),
                "message": err["msg"],
                "type": err["type"],
            }
            for err in exc.errors()
        ]
        return JSONResponse(
            status_code=400,
            content={
                "code": "validation_error",
                "message": "request validation failed",
                "details": details,
            },
        )

    @app.exception_handler(Exception)
    async def generic_error_handler(request: Request, exc: Exception) -> JSONResponse:
        request_id = request.headers.get("x-request-id", "-")
        logger.error("internal server error request_id=%s", request_id, exc_info=(type(exc), exc, exc.__traceback__))
        return JSONResponse(
            status_code=500,
            content={"code": "internal_server_error", "message": "internal server error", "details": None},
        )
