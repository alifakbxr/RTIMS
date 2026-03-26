#!/bin/bash

# RTIMS Production Deployment Script
# This script helps deploy RTIMS to production

set -e

echo "======================================"
echo "RTIMS Production Deployment"
echo "======================================"

# Check if .env.production exists
if [ ! -f .env.production ]; then
    echo "Error: .env.production file not found!"
    echo "Please copy .env.production.example to .env.production and fill in your values."
    exit 1
fi

# Load environment variables
source .env.production

# Create necessary directories
echo "Creating necessary directories..."
mkdir -p nginx/ssl
mkdir -p nginx/logs

# Generate nginx.conf if it doesn't exist
if [ ! -f nginx/nginx.conf ]; then
    echo "Creating nginx configuration..."
    cat > nginx/nginx.conf << 'NGINX_EOF'
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    # Logging
    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=web_limit:10m rate=30r/s;

    # Upstream backend
    upstream backend {
        server rtims-backend-prod:8080;
    }

    # Upstream frontend
    upstream frontend {
        server rtims-frontend-prod:3000;
    }

    # HTTPS Server
    server {
        listen 443 ssl http2;
        server_name ${API_DOMAIN};

        # SSL configuration
        ssl_certificate /etc/nginx/ssl/fullchain.pem;
        ssl_certificate_key /etc/nginx/ssl/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;

        # API routes
        location /api/ {
            limit_req zone=api_limit burst=20 nodelay;
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Frontend
        location / {
            limit_req zone=web_limit burst=50 nodelay;
            proxy_pass http://frontend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # HTTP to HTTPS redirect
    server {
        listen 80;
        server_name ${API_DOMAIN};
        return 301 https://$server_name$request_uri;
    }
}
NGINX_EOF
fi

# Pull latest images (if using registry)
echo "Pulling latest Docker images..."
docker-compose -f docker-compose.production.yml pull

# Start services
echo "Starting production services..."
docker-compose -f docker-compose.production.yml up -d

# Wait for services to be healthy
echo "Waiting for services to start..."
sleep 10

# Check service health
echo "Checking service health..."
docker-compose -f docker-compose.production.yml ps

# Run database migrations
echo "Running database migrations..."
docker-compose -f docker-compose.production.yml exec -T backend-prod /bin/sh -c "cd /root && ./main migrate up" 2>/dev/null || true

echo ""
echo "======================================"
echo "Deployment Complete!"
echo "======================================"
echo ""
echo "Services:"
echo "  - Frontend: https://${DOMAIN}"
echo "  - Backend API: https://${API_DOMAIN}/api/v1"
echo "  - Swagger Docs: https://${API_DOMAIN}/swagger/index.html"
echo ""
echo "To view logs:"
echo "  docker-compose -f docker-compose.production.yml logs -f"
echo ""
echo "To stop services:"
echo "  docker-compose -f docker-compose.production.yml down"
echo ""
