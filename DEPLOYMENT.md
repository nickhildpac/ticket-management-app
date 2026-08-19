# Ticket Management App - Containerized Deployment

A full-stack ticket management application with containerization, CI/CD, and monitoring.

## Architecture

See `docs/adr/0002-service-topology.md` (topology) and `docs/adr/0003-keycloak-authentication.md`
(authentication) for the authoritative descriptions. In short:

- **keycloak**: identity provider, port 8090. Authoritative for authentication and realm roles;
  both Go services validate its tokens and hold no signing key.
- **ticket-service** (`apps/ticket-service`, Go): authoritative ticket API, port 8080
- **ai-service** (`apps/ai-service`, Go): RAG/triage API on port 8081 plus a separate `worker` container (compose profile `ai`)
- **web** (`apps/web`, React/Vite): SPA pointed at the Go API
- **Monitoring**: Prometheus and Grafana (compose profile `obs`)
- **Orchestration**: Docker Compose profiles + root `Makefile`; images built per app
- **CI/CD**: GitHub Actions

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Git

### Local Development

1. Clone the repository:
```bash
git clone <repository-url>
cd ticket-management-system
```

2. Start the stack (root Makefile wraps `infra/compose/docker-compose.yml`):
```bash
make up        # core: postgres + keycloak + ticket-service + web
make up-ai     # + redis + ai-service (triage)
make up-obs    # + prometheus + grafana
```

3. Access the applications:
- Frontend: http://localhost:3000
- Keycloak admin console: http://localhost:8090 (admin/admin)
- Ticket API (Go): http://localhost:8080
- AI service (Go): http://localhost:8081 (profile `ai`)
- Grafana: http://localhost:3001 (admin/admin, profile `obs`)
- Prometheus: http://localhost:9090 (profile `obs`)

### Environment Variables

Copy `.env.example` to `.env` and fill it in. The security-relevant values:

```env
APP_ENV=production

# Public URL of Keycloak. Must be https outside local/test — it goes into every
# token's `iss`, and both services reject a plain-http issuer in production.
KEYCLOAK_PUBLIC_URL=https://sso.example.com
KEYCLOAK_REALM=ticket-management
KEYCLOAK_AUDIENCE=ticket-service

# Confidential client secrets. Generate with: openssl rand -base64 32
# The values in .env.example are the dev ones baked into the realm export and
# are rejected outside local/test.
AI_KEYCLOAK_CLIENT_SECRET=<generated>
KEYCLOAK_ADMIN_CLIENT_SECRET=<generated>

KEYCLOAK_ADMIN_PASSWORD=<generated>
POSTGRES_PASSWORD=<generated>
GRAFANA_PASSWORD=<generated>
```

There is no `JWT_SECRET` any more — the services verify RS256 tokens against the realm's JWKS.

### Keycloak in production

The compose service runs `start-dev` with a file-backed database, which is for local work only.
For a real deployment: run `start` with `KC_DB=postgres` against a dedicated database, terminate TLS
in front of it with `KC_HOSTNAME` set to the public URL, inject client secrets from your secret
store, and remove the seeded dev users from the imported realm. See `infra/keycloak/README.md`.

## Docker Images

The application uses multi-stage Dockerfiles for optimized production images:

### Backend Dockerfile
- **Builder stage**: Go 1.24.5-alpine for compilation
- **Final stage**: Alpine Linux with compiled binary
- **Features**: Non-root user, health checks, minimal size

### Frontend Dockerfile  
- **Builder stage**: Node.js 18-alpine for building
- **Final stage**: Nginx Alpine for serving static files
- **Features**: Gzip compression, security headers, non-root user

## CI/CD Pipeline

GitHub Actions workflow includes:

1. **Testing**:
   - Backend: Go tests with PostgreSQL service
   - Frontend: TypeScript compilation and linting

2. **Building and Pushing**:
   - Multi-platform Docker builds
   - Automatic tagging based on branch/commit
   - Container registry (GitHub Container Registry)

3. **Deployment**:
   - Production deployment on main branch merges
   - Health checks and rolling updates

## Monitoring

### Prometheus Metrics
- Application metrics: `/metrics` endpoint
- Custom metrics: HTTP requests, response times
- Infrastructure metrics: CPU, memory, disk

### Grafana Dashboards
- Application performance
- System resources
- Error rates and alerts

## Database Migrations

Both services apply their own migrations on startup and track them in separate
tables in the same database: the **ticket-service** owns the ticket/user/comment
schema (`schema_migrations`; set `AUTO_MIGRATE=false` to disable), and the
**ai-service** owns AI tables like `kb_chunks` (`ai_schema_migrations`). Run the
AI-owned ones manually with:

```bash
make migrate                          # from the repo root
# or against a running stack:
docker compose -f infra/compose/docker-compose.yml exec ai-service ./api -migrate-only
```

To disable the ticket-service's auto-migration (e.g. externally managed DB), set
`AUTO_MIGRATE=false` and apply `apps/ticket-service/migrations/` with the
golang-migrate CLI; `apps/ai-service/migrations/` can be applied the same way.

## Production Deployment

### Environment Setup

1. Set up production environment variables
2. Configure SSL/TLS certificates
3. Set up proper DNS records
4. Configure backup strategy

### Docker Compose Production

```bash
docker compose -f infra/compose/docker-compose.yml --profile ai up -d --build
```

### Scaling

```bash
# Scale the ticket API (each replica runs its own outbox relay; delivery is
# at-least-once and the AI worker dedupes on event_id)
docker compose -f infra/compose/docker-compose.yml up -d --scale ticket-service=3

# Scale the web tier (behind a load balancer)
docker compose -f infra/compose/docker-compose.yml up -d --scale web=2
```

When scaling the AI worker, give each replica a distinct `CONSUMER_NAME` so
Redis Streams pending-entry recovery can tell them apart.

## Security Considerations

- Non-root containers
- Security headers in Nginx
- Environment-based secrets
- Network isolation
- Health checks and restart policies

## Performance Optimizations

- Multi-stage builds for smaller images
- Nginx gzip compression and caching
- Database connection pooling
- Prometheus metrics collection optimization

## Troubleshooting

### Health Checks

Check service health:
```bash
docker compose -f infra/compose/docker-compose.yml ps
make logs      # tails all services
```

### Metrics Endpoint

Verify Prometheus metrics:
```bash
curl http://localhost:8080/metrics
```

### Database Issues

Check database connectivity:
```bash
docker compose -f infra/compose/docker-compose.yml exec postgres psql -U postgres -d ticket_management
```

## Development Commands

See `CLAUDE.md` for the full per-app workflow. In short:

### ticket-service (Go)
```bash
cd apps/ticket-service
go run cmd/api/main.go
go test ./...
make sqlc  # Generate SQLC code
```

### ai-service (Go)
```bash
cd apps/ai-service
go run ./cmd/api        # :8081 — also applies the kb_chunks migrations
go run ./cmd/worker     # Redis Streams triage consumer
go test ./...
```

### web (React + Vite)
```bash
cd apps/web
bun run dev
bun run build
bun run lint
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Push to your fork
5. Create a pull request

The CI/CD pipeline will automatically test and validate your changes.
