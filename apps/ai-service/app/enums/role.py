from enum import StrEnum


class UserRole(StrEnum):
    USER = "user"
    AGENT = "agent"
    ADMIN = "admin"
