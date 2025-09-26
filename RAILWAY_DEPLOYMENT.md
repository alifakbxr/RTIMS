# RTIMS Railway Deployment Guide

This guide provides step-by-step instructions for deploying your RTIMS (Real-Time Inventory Management System) application to Railway. RTIMS is a full-stack application with a Go backend, Next.js frontend, and PostgreSQL database.

## Prerequisites

Before deploying to Railway, ensure you have:

- A Railway account ([railway.app](https://railway.app))
- Git repository access to [https://github.com/alifakbxr/RTIMS](https://github.com/alifakbxr/RTIMS)
- Basic understanding of Docker and containerization

## Project Overview

- **Backend**: Go application with Gin framework
- **Frontend**: Next.js application with TypeScript
- **Database**: PostgreSQL with migration support
- **Architecture**: Multi-service Docker setup

## Step 1: Railway Setup

### 1.1 Create a New Project

1. Log in to [Railway](https://railway.app)
2. Click **"New Project"**
3. Select **"Empty Project"**
4. Choose your preferred region (recommended: US-West or Asia-Southeast)

### 1.2 Connect Git Repository

1. In your Railway project dashboard, click **"GitHub"**
2. Authorize Railway to access your GitHub account
3. Select the `RTIMS` repository from the list
4. Click **"Deploy"**

## Step 2: Configure Services

Railway will automatically detect your services based on the configuration files. However, you may need to manually configure some services:

### 2.1 Backend Service

1. Go to your Railway project dashboard
2. Click **"New Service"** → **"Database"** → **"PostgreSQL"**
3. Name it `rtims-postgres`
4. Wait for the database to be provisioned

### 2.2 Redis Service (if needed)

1. Click **"New Service"** → **"Database"** → **"Redis"**
2. Name it `rtims-redis`

## Step 3: Environment Variables

Configure the following environment variables in your Railway project settings:

### Backend Environment Variables

```
DATABASE_URL=postgresql://postgres:password@rtims-postgres:5432/railway
REDIS_URL=redis://rtims-redis:6379
JWT_SECRET=your-super-secret-jwt-key-here
GO_ENV=production
PORT=8080
```

### Frontend Environment Variables

```
NEXT_PUBLIC_API_URL=https://your-backend-service.railway.internal
NEXT_PUBLIC_APP_ENV=production
```

### Database Environment Variables

```
POSTGRES_DB=rtims_production
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your-secure-password
```

## Step 4: Database Migration

### 4.1 Initial Migration

1. Connect to your Railway PostgreSQL database
2. Run the migration script:
   ```bash
   psql $DATABASE_URL -f database/migrations/001_initial_schema.sql
   ```

### 4.2 Using Railway CLI (Alternative)

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login to Railway
railway login

# Connect to your project
railway link

# Run migration on database service
railway run --service rtims-postgres -- psql $DATABASE_URL -f database/migrations/001_initial_schema.sql
```

## Step 5: Deployment Configuration

### 5.1 Build Settings

In your Railway project settings:

1. **Build Command**: Leave empty (uses Dockerfile)
2. **Start Command**: `docker-compose -f docker-compose.production.yml up`
3. **Root Directory**: `./`
4. **Watch Paths**:
   - `backend/**`
   - `frontend/**`
   - `database/**`

### 5.2 Health Checks

Railway will automatically configure health checks based on your application:

- **Backend**: `/api/health` endpoint
- **Frontend**: Default Next.js health check
- **Database**: PostgreSQL health check

## Step 6: Domain Configuration

### 6.1 Custom Domain (Optional)

1. Go to your Railway project settings
2. Click **"Custom Domains"**
3. Add your domain and configure DNS settings
4. Railway will provide the required CNAME records

### 6.2 Service URLs

After deployment, you'll get URLs like:
- Frontend: `https://your-app.railway.app`
- Backend API: `https://your-backend.railway.internal`
- Database: Available via internal networking

## Step 7: Monitoring and Logs

### 7.1 View Logs

1. In Railway dashboard, click on any service
2. Go to **"Logs"** tab to view real-time logs
3. Check for errors during startup and runtime

### 7.2 Monitor Resource Usage

1. Go to **"Metrics"** tab in each service
2. Monitor CPU, memory, and network usage
3. Set up alerts if needed

## Common Errors and Troubleshooting

### Error 1: Build Failures

**Symptoms**: Deployment fails during build phase

**Solutions**:
1. Check Dockerfile syntax:
   ```dockerfile
   # Ensure proper Go version in backend/Dockerfile
   FROM golang:1.21-alpine AS builder

   # Ensure proper Node.js version in frontend/Dockerfile
   FROM node:18-alpine AS base
   ```

2. Verify nixpacks.toml configuration
3. Check available disk space in Railway

### Error 2: Database Connection Issues

**Symptoms**: Backend can't connect to PostgreSQL

**Solutions**:
1. Verify DATABASE_URL format:
   ```
   DATABASE_URL=postgresql://username:password@hostname:port/database
   ```

2. Check if PostgreSQL service is running
3. Ensure database migrations have been applied

### Error 3: Port Binding Issues

**Symptoms**: Application fails to start, port already in use

**Solutions**:
1. Ensure PORT environment variable is set correctly
2. Check if multiple services are trying to bind to the same port
3. Verify Docker port mappings in docker-compose.production.yml

### Error 4: Environment Variable Issues

**Symptoms**: Application starts but behaves unexpectedly

**Solutions**:
1. Double-check all required environment variables
2. Ensure sensitive data is properly encrypted
3. Verify environment variable names match your code

### Error 5: Memory/CPU Limits

**Symptoms**: Application crashes under load

**Solutions**:
1. Monitor resource usage in Railway dashboard
2. Upgrade to a higher-tier plan if needed
3. Optimize your application code for better performance

## Performance Optimization

### 1. Database Optimization

```sql
-- Add indexes for better performance
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_stock_product ON stock(product_id);
```

### 2. Frontend Optimization

```bash
# In your frontend directory
npm run build
# Ensure proper caching headers in next.config.ts
```

### 3. Backend Optimization

```go
// In your Go backend, ensure proper connection pooling
db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
    PrepareStmt: true,
})
```

## Security Considerations

### 1. Environment Variables

- Never commit sensitive data to version control
- Use Railway's built-in secret management
- Rotate secrets regularly

### 2. Database Security

- Use strong passwords for database connections
- Limit database access to necessary services only
- Enable SSL for database connections in production

### 3. API Security

- Implement proper authentication middleware
- Use HTTPS for all communications
- Set up rate limiting

## Backup and Recovery

### 1. Database Backups

Railway automatically backs up your PostgreSQL database:

1. Go to your PostgreSQL service in Railway
2. Click **"Data"** tab
3. Download backups as needed

### 2. Manual Backups

```bash
# Using Railway CLI
railway run --service rtims-postgres -- pg_dump $DATABASE_URL > backup.sql
```

## Scaling Considerations

### 1. Horizontal Scaling

For high-traffic applications:

1. Deploy multiple instances of your frontend
2. Use Railway's load balancer
3. Implement session management for backend scaling

### 2. Database Scaling

1. Monitor database performance
2. Consider read replicas for read-heavy workloads
3. Optimize queries and add proper indexing

## Cost Optimization

### 1. Resource Management

- Monitor service usage in Railway dashboard
- Shut down unused services
- Use appropriate service tiers

### 2. Database Optimization

- Archive old data
- Use connection pooling
- Optimize query performance

## Support and Resources

### Official Documentation

- [Railway Documentation](https://docs.railway.app)
- [Docker Documentation](https://docs.docker.com)
- [PostgreSQL Documentation](https://postgresql.org/docs)

### Community Support

- [Railway Community](https://railway.app/community)
- [GitHub Issues](https://github.com/alifakbxr/RTIMS/issues)

### Monitoring Tools

- Railway Dashboard Metrics
- Application Performance Monitoring (APM) tools
- Log aggregation services

## Troubleshooting Commands

### Railway CLI Commands

```bash
# View service logs
railway logs --service rtims-backend

# Run commands on services
railway run --service rtims-backend -- go version

# View service status
railway status

# Restart services
railway up
```

### Database Troubleshooting

```bash
# Connect to database
railway run --service rtims-postgres -- psql $DATABASE_URL

# Check database connections
SELECT * FROM pg_stat_activity;

# View database size
SELECT schemaname, tablename, attname, n_distinct
FROM pg_stats
WHERE schemaname = 'public';
```

## Next Steps

1. Deploy your application using this guide
2. Test all functionality thoroughly
3. Set up monitoring and alerting
4. Configure backup strategies
5. Plan for scaling as your user base grows

---

**Note**: This deployment guide is specifically tailored for your RTIMS project. Always test deployments in a staging environment before deploying to production.

For additional support, refer to the [Railway documentation](https://docs.railway.app) or create an issue in your [GitHub repository](https://github.com/alifakbxr/RTIMS).