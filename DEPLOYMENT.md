# Ticket Management App - Containerized Deployment

A full-stack ticket management application with containerization, CI/CD, and monitoring.

## Architecture

- **Backend**: Go API with PostgreSQL database
- **Frontend**: React SPA with TypeScript and Tailwind CSS
- **Monitoring**: Prometheus and Grafana
- **Containerization**: Multi-stage Dockerfiles
- **CI/CD**: GitHub Actions

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Git

### Local Development

1. Clone the repository:
```bash
git clone <repository-url>
cd ticket-management-app
```

2. Start all services:
```bash
docker-compose up -d
```

3. Access the applications:
- Frontend: http://localhost:80
- Backend API: http://localhost:8080
- Grafana: http://localhost:3001 (admin/admin)
- Prometheus: http://localhost:9090

### Environment Variables

Create a `.env` file in the root directory:

```env
JWT_SECRET=your-jwt-secret-key
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

Run migrations manually:
```bash
docker-compose exec backend make migrateup
```

Rollback migrations:
```bash
docker-compose exec backend make migratedown
```

## Production Deployment

### Environment Setup

1. Set up production environment variables
2. Configure SSL/TLS certificates
3. Set up proper DNS records
4. Configure backup strategy

### Docker Compose Production

Use the production-optimized docker-compose.yml:
```bash
docker-compose -f docker-compose.yml up -d
```

### Scaling

```bash
# Scale backend
docker-compose up -d --scale backend=3

# Scale frontend (with load balancer)
docker-compose up -d --scale frontend=2
```

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
docker-compose ps
docker-compose logs backend
docker-compose logs frontend
```

### Metrics Endpoint

Verify Prometheus metrics:
```bash
curl http://localhost:8080/metrics
```

### Database Issues

Check database connectivity:
```bash
docker-compose exec postgres psql -U postgres -d ticket_management
```

## Development Commands

### Backend
```bash
cd backend
go run cmd/api/main.go
go test -v ./...
make sqlc  # Generate SQLC code
```

### Frontend
```bash
cd frontend
npm run dev
npm run build
npm run lint
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Push to your fork
5. Create a pull request

The CI/CD pipeline will automatically test and validate your changes.