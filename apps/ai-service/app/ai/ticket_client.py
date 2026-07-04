from __future__ import annotations

import datetime as dt

import httpx
import jwt

from app.core.config import Settings


class TicketServiceClient:
    """Calls back into the authoritative Go ticket API to apply triage results.

    Auth: mints a short-lived HS256 JWT for the AI service account. The claim
    shape (RegisteredClaims + ``role``) mirrors the ticket-service auth
    middleware; the ``sub`` must be an existing agent/admin user id.
    """

    def __init__(self, settings: Settings, client: httpx.Client | None = None) -> None:
        self.settings = settings
        self.base_url = settings.ticket_service_url.rstrip("/") + "/api/v1"
        self.client = client or httpx.Client(timeout=10.0)

    def _token(self) -> str:
        now = dt.datetime.now(tz=dt.UTC)
        payload = {
            "sub": self.settings.service_account_id,
            "role": self.settings.service_account_role,
            "iss": self.settings.jwt_issuer,
            "aud": self.settings.jwt_audience,
            "iat": now,
            "exp": now + dt.timedelta(minutes=5),
        }
        return jwt.encode(payload, self.settings.jwt_secret, algorithm=self.settings.jwt_algorithm)

    def _headers(self) -> dict[str, str]:
        return {"Authorization": f"Bearer {self._token()}"}

    def add_comment(self, ticket_id: str, description: str) -> None:
        resp = self.client.post(
            f"{self.base_url}/comments",
            json={"ticket_id": ticket_id, "description": description},
            headers=self._headers(),
        )
        resp.raise_for_status()

    def set_state(self, ticket_id: str, state: str) -> None:
        resp = self.client.patch(
            f"{self.base_url}/tickets/{ticket_id}",
            json={"state": state},
            headers=self._headers(),
        )
        resp.raise_for_status()

    def verify_access(self) -> None:
        """Probe the ticket API with the service token; raises on failure.

        Surfaces misconfiguration (unset/invalid AI_SERVICE_ACCOUNT_ID, wrong role,
        iss/aud mismatch, or an unreachable API) loudly at worker startup instead
        of silently 403-ing on every consumed event.
        """
        resp = self.client.get(
            f"{self.base_url}/tickets",
            params={"limit": 1, "offset": 0},
            headers=self._headers(),
        )
        resp.raise_for_status()
