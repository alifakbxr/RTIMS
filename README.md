# RTIMS - Real-Time Inventory Management System

[![Build Status](https://github.com/yourusername/rtims/workflows/CI/CD/badge.svg)](https://github.com/yourusername/rtims/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/rtims)](https://goreportcard.com/report/github.com/yourusername/rtims)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive, production-ready, full-stack inventory management system built with modern technologies for real-time stock tracking, user management, and reporting.

## 🚀 Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd RTIMS

# Start with Docker (easiest)
docker-compose up -d

# Or use make commands
make setup
make dev
```

Access the application:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Swagger Docs**: http://localhost:8080/swagger/index.html

**Default Login**: `admin@example.com` / `admin123`

📖 See [QUICKSTART.md](QUICKSTART.md) for detailed setup instructions.
📖 See [DEPLOYMENT.md](DEPLOYMENT.md) for production deployment guide.

## 📋 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Quick Start](#-quick-start)
- [Development](#-development)
- [API Documentation](#-api-documentation)
- [Deployment](#-deployment)
- [Project Structure](#-project-structure)
- [Contributing](#-contributing)
- [License](#-license)

## 🚀 Features

### Core Features
- **Real-time Inventory Tracking**: Live stock updates via WebSocket
- **User Management**: Role-based access control (Staff/Admin)
- **Product Management**: Complete CRUD operations with advanced filtering
- **Stock Movement Tracking**: Automatic logging of all inventory changes
- **Low Stock Alerts**: Automated notifications for inventory thresholds
- **Audit Logging**: Complete audit trail for compliance
- **Dashboard Analytics**: Visual insights with charts and metrics
- **Report Generation**: Export reports in multiple formats (CSV, PDF, Excel)

### Technical Features
- **RESTful API**: Comprehensive API with Swagger documentation
- **WebSocket Integration**: Real-time updates for connected clients
- **JWT Authentication**: Secure token-based authentication with refresh tokens
- **Role-Based Access Control**: Granular permissions for different user roles
- **Responsive Design**: Optimized for desktop and mobile devices
- **Email Notifications**: Automated email alerts for important events
- **Database Migrations**: Version-controlled database schema management

## 🛠️ Tech Stack

### Backend
- **Go**: High-performance backend with Gin framework
- **PostgreSQL**: Robust relational database with advanced features
- **Redis**: Session management and caching
- **WebSocket**: Real-time communication (Gorilla WebSocket)
- **JWT**: Secure authentication and authorization
- **Swagger**: API documentation

### Frontend
- **Next.js 14**: React framework with App Router
- **TypeScript**: Type-safe development
- **TailwindCSS**: Utility-first CSS framework
- **shadcn/ui**: Modern component library
- **Recharts**: Data visualization
- **React Hook Form**: Form handling with validation
- **Axios**: HTTP client with interceptors

### DevOps & Tools
- **Docker**: Containerization for easy deployment
- **GitHub Actions**: CI/CD pipeline (bonus feature)
- **Environment Configuration**: Multi-environment support

## 📋 Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 13+
- Redis 6+
- Docker (optional)

## 🚀 Quick Start

### 1. Clone the Repository
```bash
git clone <repository-url>
cd RTIMS
```

### 2. Backend Setup

#### Install Dependencies
```bash
cd backend
go mod download
```

#### Environment Configuration
Create a `.env` file in the backend directory:
```env
ENVIRONMENT=development
PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/rtims?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
REFRESH_SECRET=your-super-secret-refresh-key-change-this-in-production
EMAIL_API_KEY=your-email-api-key
EMAIL_FROM=noreply@rtims.com
```

#### Database Setup
```bash
# Run database migrations
go run main.go migrate up

# Seed initial data
go run main.go seed
```

#### Start the Backend
```bash
go run main.go
```

The backend will start on `http://localhost:8080`

### 3. Frontend Setup

#### Install Dependencies
```bash
cd frontend
npm install
```

#### Environment Configuration
Create a `.env.local` file in the frontend directory:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

#### Start the Frontend
```bash
npm run dev
```

The frontend will start on `http://localhost:3000`

### 4. Database Setup

#### Using Docker (Recommended)
```bash
# Start PostgreSQL and Redis
docker-compose up -d
```

#### Manual Setup
1. Install PostgreSQL and Redis
2. Create database: `rtims`
3. Run the migration file: `database/migrations/001_initial_schema.sql`

## 🔧 Development

### Project Structure
```
RTIMS/
├── backend/                 # Complete Go backend
│   ├── config/             # Configuration management
│   ├── docs/               # API documentation
│   ├── internal/           # Private application code
│   │   ├── database/       # Database connection and models
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # Custom middleware
│   │   ├── models/         # Data models
│   │   └── websocket/      # WebSocket implementation
│   └── main.go             # Application entry point
├── frontend/               # Complete Next.js frontend
│   ├── src/
│   │   ├── app/           # Next.js app router pages
│   │   ├── components/    # React components
│   │   ├── contexts/      # React contexts
│   │   ├── lib/           # Utility libraries
│   │   └── types/         # TypeScript type definitions
│   └── public/            # Static assets
├── database/              # Database migrations and seeds
└── docker-compose.yml     # Docker configuration
```

### API Documentation

Access Swagger documentation at: `http://localhost:8080/swagger/index.html`

### Available Scripts

#### Backend
```bash
go run main.go          # Start the server
go test ./...           # Run tests
go mod tidy             # Clean up dependencies
```

#### Frontend
```bash
npm run dev             # Start development server
npm run build           # Build for production
npm run start           # Start production server
npm run lint            # Run ESLint
npm run type-check      # Run TypeScript check
```

## 🔐 Authentication

### Demo Accounts
- **Admin**: `admin@example.com` / `admin123`
- **Staff**: `staff@example.com` / `staff123`

### API Authentication
Include JWT token in Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## 📊 Key Features Explained

### Real-Time Updates
- WebSocket connections for live stock updates
- Automatic notifications for low stock items
- Real-time dashboard statistics

### Role-Based Access Control
- **Staff**: Can manage products and stock levels
- **Admin**: Full system access including user management and reports

### Audit Trail
- All actions are logged with user, timestamp, and IP address
- Complete history of changes for compliance

### Advanced Reporting
- Inventory reports with customizable filters
- Stock movement analysis
- Export capabilities (CSV, PDF, Excel)

## 🧪 Testing

### Backend Tests
```bash
go test ./internal/... -v
```

### Frontend Tests
```bash
npm run test
```

## 🚢 Deployment

### Docker Deployment
```bash
# Build and run with Docker Compose
docker-compose up --build
```

### Production Build
```bash
# Backend
go build -o bin/rtims-backend ./backend

# Frontend
cd frontend && npm run build
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support, please contact the development team or create an issue in the repository.

## 🔄 Changelog

### Version 1.0.0
- Initial release with core inventory management features
- Real-time WebSocket integration
- Role-based authentication
- Comprehensive API documentation
- Modern React frontend with TypeScript

---

Built with ❤️ using Go, Next.js, and modern web technologies.