# Ticket Management App - Containerized Deployment

A full-stack ticket management application with containerization, CI/CD, and monitoring.

## Architecture

See `docs/adr/0002-service-topology.md` for the authoritative description. In short:

- **ticket-service** (`apps/ticket-service`, Go): authoritative ticket API, port 8080
- **ai-service** (`apps/ai-service`, FastAPI): RAG/triage worker + endpoint, port 8081 (compose profile `ai`)
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
make up        # core: postgres + ticket-service + web
make up-ai     # + redis + ai-service (triage)
make up-obs    # + prometheus + grafana
```

3. Access the applications:
- Frontend: http://localhost:3000
- Ticket API (Go): http://localhost:8080
- AI service (FastAPI): http://localhost:8081 (profile `ai`)
- Grafana: http://localhost:3001 (admin/admin, profile `obs`)
- Prometheus: http://localhost:9090 (profile `obs`)

### Environment Variables

Create a `.env` file in the root directory:

```env
JWT_SECRET=your-jwt-secret-key
APP_ENV=production
GRAFANA_PASSWORD=your-grafana-password
```

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

The Go **ticket-service applies its own migrations on startup** (owns the
ticket/user/comment schema); set `AUTO_MIGRATE=false` to disable. The
**ai-service applies its Alembic migrations** (AI-owned tables like `kb_chunks`)
on container start, or run them manually:

```bash
make migrate                          # AI-service Alembic (from repo root)
# or against a running stack:
docker compose -f infra/compose/docker-compose.yml exec ai-service uv run alembic upgrade head
```

To disable the Go service's auto-migration (e.g. externally managed DB), set
`AUTO_MIGRATE=false` and apply `apps/ticket-service/migrations/` with the
golang-migrate CLI.

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

### ai-service (Python)
```bash
cd apps/ai-service
uv run uvicorn app.main:app --reload --port 8081
uv run python -m app.ai.worker
uv run pytest
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
