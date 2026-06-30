# ADR 0001: FastAPI Backend Direction

## Status

Accepted

## Context

The repository currently contains two backend implementations:

- `backend/`: the original Go service using domain, service, ports, and adapters layers.
- `backend-python/`: the Python/FastAPI service using API, service, repository, model, and policy layers.

Maintaining two production-capable backends without a clear owner creates ambiguity for API changes, domain-rule parity, documentation, frontend integration, deployment, and testing. Recent API normalization work already updates both implementations, but future work needs a single authoritative backend target.

## Decision

`backend-python/` is the authoritative backend going forward and is intended to replace the Go backend.

The Go backend remains in the repository temporarily as a reference implementation and compatibility baseline during migration. It should not receive new product features unless they are required to preserve parity during the transition.

## Parity Expectations

Until the Go backend is removed, both backends should remain compatible for externally visible behavior that the frontend or API clients depend on:

- HTTP routes, methods, request bodies, response bodies, status codes, auth behavior, and cookie behavior.
- Ticket state-machine semantics and wire values.
- RBAC decisions for ticket, comment, user, and admin operations.
- Error response envelope shape.
- Database-visible lifecycle behavior where both implementations target the same product capability.

When behavior differs, the FastAPI implementation is the source of truth. The divergence should be documented in the related issue or pull request, and the Go implementation should either be aligned or explicitly marked as deprecated for that behavior.

Refresh-token storage is one such divergence during migration: FastAPI persists hashed refresh tokens and revokes them on rotation; the Go backend still uses stateless refresh JWTs. New authentication work should target the FastAPI model rather than extending the Go implementation.

## Migration Strategy

1. Make new backend features in `backend-python/` first.
2. Keep frontend API callers pointed at the canonical API contract, not implementation-specific routes.
3. Add or update FastAPI tests for new behavior. Add Go parity tests only when the Go backend is still expected to support the same behavior during migration.
4. Keep generated OpenAPI/Postman/API docs aligned with FastAPI as the target contract.
5. Move deployment and local-development documentation toward FastAPI as the default backend.
6. Remove the Go backend after FastAPI covers the required API surface, migration risks are resolved, and deployment no longer depends on `backend/`.

## Consequences

- Engineers have one backend direction for new work: FastAPI.
- The Go backend can still be used for comparison and temporary compatibility, but it is no longer the strategic implementation.
- Any duplicated business logic must be treated as migration debt and resolved in favor of the FastAPI behavior.
- Documentation and operational scripts should be updated incrementally so the repository no longer presents both backends as equal choices.
