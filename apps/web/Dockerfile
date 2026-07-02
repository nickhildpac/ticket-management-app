# Multi-stage build for React frontend
FROM oven/bun:1-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy package.json and lock file
COPY package.json bun.lock* ./

# Install dependencies
RUN bun install --frozen-lockfile

# Copy the source code
COPY . .

# Build the application (VITE_API_URL can be passed as build arg)
ARG VITE_API_URL=http://localhost:8081
ENV VITE_API_URL=${VITE_API_URL}
RUN bun run build

# Production stage
FROM nginx:1.25-alpine

# Install wget for health check
RUN apk add --no-cache wget

# Remove default nginx static content
RUN rm -rf /usr/share/nginx/html/*

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy our custom nginx.conf as main config
COPY nginx.conf /etc/nginx/nginx.conf

# Expose port
EXPOSE 8080

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
