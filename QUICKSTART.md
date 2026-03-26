# RTIMS - Quick Start Guide

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go** 1.23 or later
- **Node.js** 18 or later
- **PostgreSQL** 13 or later
- **Redis** 6 or later
- **Docker** and **Docker Compose** (optional, for containerized deployment)

## Option 1: Quick Start with Docker (Recommended)

This is the easiest way to get RTIMS up and running.

### Step 1: Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd RTIMS

# Copy environment files
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local
```

### Step 2: Start All Services

```bash
# Start PostgreSQL, Redis, Backend, and Frontend
docker-compose up -d

# View logs
docker-compose logs -f
```

### Step 3: Access the Application

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Swagger Documentation**: http://localhost:8080/swagger/index.html

### Default Login Credentials

- **Admin**: `admin@example.com` / `admin123`
- **Staff**: `staff@example.com` / `staff123`

---

## Option 2: Manual Development Setup

### Step 1: Database Setup

```bash
# Start PostgreSQL and Redis using Docker
docker-compose up -d postgres redis

# Or install them manually on your system
```

### Step 2: Backend Setup

```bash
cd backend

# Copy environment file
cp .env.example .env

# Edit .env and update database credentials if needed

# Install dependencies
go mod download

# The database migration will run automatically on first start
# Or run manually: go run main.go migrate up

# Start the backend server
go run main.go
```

The backend will start on http://localhost:8080

### Step 3: Frontend Setup

```bash
cd frontend

# Copy environment file
cp .env.example .env.local

# Install dependencies
npm install

# Start the development server
npm run dev
```

The frontend will start on http://localhost:3000

---

## Verification

### Check Backend Health

```bash
curl http://localhost:8080/health
```

Expected response: `{"status":"healthy","timestamp":"..."}`

### Check Frontend

Open http://localhost:3000 in your browser. You should be redirected to the login page.

### Test API with Swagger

1. Open http://localhost:8080/swagger/index.html
2. Click on any endpoint to see details
3. Use `/api/v1/auth/login` to get an access token
4. The token will be automatically used for authenticated requests

---

## Common Issues and Solutions

### Issue: Port Already in Use

**Solution**: Change the port in the configuration:
- Backend: Edit `backend/.env` and change `PORT`
- Frontend: Edit `frontend/package.json` and change the dev script port

### Issue: Database Connection Failed

**Solution**: 
1. Ensure PostgreSQL is running: `docker ps | grep postgres`
2. Check the connection string in `backend/.env`
3. Verify the database exists: `docker exec -it rtims-postgres psql -U rtims_user -l`

### Issue: JWT Token Errors

**Solution**: 
1. Ensure `JWT_SECRET` is at least 32 characters long
2. Regenerate secrets: `openssl rand -base64 32`

### Issue: Frontend Can't Connect to Backend

**Solution**: 
1. Ensure backend is running on port 8080
2. Check `NEXT_PUBLIC_API_URL` in `frontend/.env.local`
3. Clear browser cache and cookies

---

## Development Commands

### Backend

```bash
# Run development server
go run main.go

# Run tests
go test ./... -v

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Build production binary
go build -o bin/rtims-backend ./backend

# Generate Swagger docs
swag init
```

### Frontend

```bash
# Development server
npm run dev

# Production build
npm run build

# Start production server
npm run start

# Run linter
npm run lint

# Type check
npx tsc --noEmit
```

---

## Project Structure

```
RTIMS/
├── backend/                 # Go backend
│   ├── config/             # Configuration
│   ├── internal/           # Application code
│   │   ├── database/       # Database layer
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # Middleware
│   │   ├── models/         # Data models
│   │   └── websocket/      # WebSocket hub
│   └── main.go             # Entry point
├── frontend/               # Next.js frontend
│   ├── src/
│   │   ├── app/           # Pages
│   │   ├── components/    # React components
│   │   ├── contexts/      # React contexts
│   │   ├── lib/           # Utilities
│   │   └── types/         # TypeScript types
│   └── public/            # Static assets
├── database/              # Database migrations
├── docker-compose.yml     # Docker configuration
└── README.md              # This file
```

---

## Next Steps

1. **Explore the Dashboard**: Login and explore the dashboard features
2. **Add Products**: Use the Products page to add inventory items
3. **Manage Users**: Admin users can manage team members
4. **Generate Reports**: Create and export inventory reports
5. **Configure Settings**: Customize system settings to your needs

---

## Support

For issues or questions:
- Check the [README.md](README.md) for detailed documentation
- Review the [Swagger API docs](http://localhost:8080/swagger/index.html)
- Check the application logs: `docker-compose logs -f`

---

## Production Deployment

For production deployment, see [DEPLOYMENT.md](DEPLOYMENT.md) or run:

```bash
# Using the deployment script
./deploy.sh  # Linux/Mac
deploy.bat   # Windows
```

Make sure to:
1. Copy `.env.production.example` to `.env.production`
2. Update all environment variables with secure values
3. Configure your domain and SSL certificates
