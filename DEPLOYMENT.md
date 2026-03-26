# RTIMS - Production Deployment Guide

## Overview

This guide covers deploying RTIMS to production using Docker Compose. For development setup, see [QUICKSTART.md](QUICKSTART.md).

## Prerequisites

- Docker and Docker Compose installed on your server
- Domain name configured (optional but recommended)
- SSL certificates (Let's Encrypt recommended)
- At least 2GB RAM and 2 CPU cores
- PostgreSQL and Redis knowledge

## Step 1: Server Preparation

### Recommended Server Specifications

- **CPU**: 2+ cores
- **RAM**: 2GB+ (4GB recommended)
- **Storage**: 20GB+ SSD
- **OS**: Ubuntu 20.04+ or similar Linux distribution

### Install Docker (if not installed)

```bash
# Update package index
sudo apt update

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

## Step 2: Clone and Configure

```bash
# Clone repository
git clone <repository-url> /opt/rtims
cd /opt/rtims

# Copy production environment file
cp .env.production.example .env.production
```

## Step 3: Configure Environment Variables

Edit `.env.production` with your secure values:

```bash
# Database password (use a strong password)
POSTGRES_PASSWORD=your-very-secure-postgres-password

# JWT Secrets (generate with: openssl rand -base64 32)
JWT_SECRET=your-production-jwt-secret-min-32-characters-long
REFRESH_SECRET=your-production-refresh-secret-min-32-chars

# Docker Registry (optional - for CI/CD)
REGISTRY=ghcr.io
BACKEND_IMAGE=your-username/rtims/backend
FRONTEND_IMAGE=your-username/rtims/frontend

# Domain Configuration
DOMAIN=yourdomain.com
API_DOMAIN=api.yourdomain.com
```

## Step 4: SSL Certificate Setup

### Option A: Let's Encrypt (Recommended)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Get certificates
sudo certbot certonly --standalone -d yourdomain.com -d api.yourdomain.com

# Copy certificates to nginx folder
sudo cp /etc/lets/live/yourdomain.com/fullchain.pem nginx/ssl/
sudo cp /etc/lets/live/yourdomain.com/privkey.pem nginx/ssl/
```

### Option B: Self-Signed (Development Only)

```bash
mkdir -p nginx/ssl
cd nginx/ssl

# Generate self-signed certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout privkey.pem \
  -out fullchain.pem \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=yourdomain.com"
```

## Step 5: Deploy

### Using the Deployment Script

```bash
# Make script executable
chmod +x deploy.sh

# Run deployment
./deploy.sh
```

### Manual Deployment

```bash
# Start all services
docker-compose -f docker-compose.production.yml up -d

# Check status
docker-compose -f docker-compose.production.yml ps

# View logs
docker-compose -f docker-compose.production.yml logs -f
```

## Step 6: Verify Deployment

### Check Service Health

```bash
# Check all containers are running
docker-compose -f docker-compose.production.yml ps

# Expected output:
# NAME                    STATUS    PORTS
# rtims-backend-prod      Up        8080/tcp
# rtims-frontend-prod     Up        3000/tcp
# rtims-postgres-prod     Up        5432/tcp
# rtims-redis-prod        Up        6379/tcp
# rtims-nginx-prod        Up        0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp
```

### Test Endpoints

```bash
# Health check
curl -k https://yourdomain.com/health

# API endpoint
curl -k https://api.yourdomain.com/api/v1/health

# Frontend
curl -k https://yourdomain.com
```

### Access the Application

1. Open https://yourdomain.com in your browser
2. Login with admin credentials:
   - Email: `admin@example.com`
   - Password: `admin123`

**Important**: Change the default passwords immediately!

## Step 7: Database Backup Setup

### Manual Backup

```bash
# Backup database
docker exec rtims-postgres-prod pg_dump -U rtims_user rtims_prod > backup_$(date +%Y%m%d).sql

# Restore database
docker exec -i rtims-postgres-prod psql -U rtims_user rtims_prod < backup_20240101.sql
```

### Automated Backups

Create a cron job for automated backups:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * docker exec rtims-postgres-prod pg_dump -U rtims_user rtims_prod > /backups/rtims_$(date +\%Y\%m\%d).sql
```

## Monitoring and Maintenance

### View Logs

```bash
# All services
docker-compose -f docker-compose.production.yml logs -f

# Specific service
docker-compose -f docker-compose.production.yml logs -f backend-prod

# Last 100 lines
docker-compose -f docker-compose.production.yml logs --tail=100 backend-prod
```

### Resource Usage

```bash
# Check container resource usage
docker stats
```

### Update Application

```bash
# Pull latest changes
git pull

# Rebuild and restart
docker-compose -f docker-compose.production.yml up -d --build

# Clean up old images
docker image prune -f
```

### Restart Services

```bash
# Restart all services
docker-compose -f docker-compose.production.yml restart

# Restart specific service
docker-compose -f docker-compose.production.yml restart backend-prod
```

### Stop Services

```bash
# Stop all services
docker-compose -f docker-compose.production.yml down

# Stop and remove volumes (WARNING: deletes data!)
docker-compose -f docker-compose.production.yml down -v
```

## Troubleshooting

### Backend Won't Start

```bash
# Check logs
docker-compose -f docker-compose.production.yml logs backend-prod

# Common issues:
# 1. Database not ready - wait for postgres to start
# 2. Wrong connection string - check DATABASE_URL in .env.production
# 3. Port conflict - check if port 8080 is in use
```

### Frontend Shows Connection Error

```bash
# Check NEXT_PUBLIC_API_URL in frontend Dockerfile
# Ensure it points to the correct backend URL

# Rebuild frontend
docker-compose -f docker-compose.production.yml build frontend-prod
docker-compose -f docker-compose.production.yml up -d frontend-prod
```

### Database Connection Issues

```bash
# Test database connection
docker exec -it rtims-postgres-prod psql -U rtims_user -d rtims_prod

# Check if tables exist
\dt

# View database size
SELECT pg_size_pretty(pg_database_size('rtims_prod'));
```

### High Memory Usage

```bash
# Check memory limits in docker-compose.production.yml
# Adjust as needed:
# deploy:
#   resources:
#     limits:
#       memory: 512M
```

## Security Best Practices

1. **Change Default Passwords**: Immediately change admin@example.com password
2. **Use Strong Secrets**: Generate secure JWT secrets (min 32 chars)
3. **Enable Firewall**: Only allow ports 80, 443, and SSH
4. **Regular Updates**: Keep Docker images and system packages updated
5. **Monitor Logs**: Set up log monitoring for suspicious activity
6. **Backup Regularly**: Implement automated daily backups
7. **Use HTTPS**: Never run production without SSL
8. **Rate Limiting**: Already configured in nginx, adjust if needed

## Performance Tuning

### Database Optimization

```sql
-- Analyze tables for query optimization
ANALYZE;

-- Add indexes for frequently queried columns
CREATE INDEX IF NOT EXISTS idx_products_low_stock ON products(stock, minimum_threshold) 
  WHERE stock <= minimum_threshold;
```

### Backend Optimization

Edit `docker-compose.production.yml` to increase replicas:

```yaml
backend-prod:
  deploy:
    replicas: 3  # Increase for high traffic
```

### Frontend Optimization

Enable caching in nginx.conf:

```nginx
location /_next/static/ {
    expires 365d;
    add_header Cache-Control "public, immutable";
}
```

## Support

For production issues:
1. Check application logs
2. Review resource usage with `docker stats`
3. Verify all services are healthy
4. Check database connection and disk space
5. Review nginx error logs: `docker exec rtims-nginx-prod tail -f /var/log/nginx/error.log`

---

**Note**: This deployment guide assumes a single-server setup. For high-availability deployments, consider using Kubernetes or Docker Swarm with load balancers and replicated databases.
