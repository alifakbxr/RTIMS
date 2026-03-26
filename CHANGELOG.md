# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-01-XX

### Added

#### Backend
- **Authentication System**
  - JWT-based authentication with access and refresh tokens
  - User registration and login endpoints
  - Password reset functionality via email
  - Role-based access control (Admin/Staff)

- **Product Management**
  - Complete CRUD operations for products
  - Stock movement tracking with multiple reasons (purchase, sale, adjustment, return, damage, transfer)
  - Product categorization
  - Low stock threshold alerts
  - Advanced filtering and pagination

- **User Management**
  - Admin-only user management endpoints
  - User profile management
  - Role assignment and activation/deactivation

- **Dashboard & Analytics**
  - Real-time dashboard statistics
  - Low stock alerts
  - System status monitoring
  - Activity tracking

- **Reporting**
  - Inventory reports with multiple export formats (JSON, CSV, PDF)
  - Stock movement reports
  - User activity reports
  - Customizable date ranges and filters

- **Audit Logging**
  - Complete audit trail for all CRUD operations
  - IP address and user agent tracking
  - Old/new value storage for updates

- **WebSocket Integration**
  - Real-time stock updates
  - Live notifications
  - System status broadcasting

- **Security**
  - CORS middleware
  - Rate limiting
  - Security headers (XSS protection, clickjacking prevention)
  - Password hashing with bcrypt

#### Frontend
- **Pages**
  - Login page with form validation
  - Dashboard with statistics and alerts
  - Products management page with filters
  - Users management page (Admin only)
  - Reports page with export functionality
  - Settings page

- **Components**
  - Responsive navigation sidebar
  - Product modal for create/edit
  - Reusable UI components (cards, buttons, badges, inputs)
  - Alert notifications
  - Loading states

- **Features**
  - Real-time authentication context
  - Automatic token refresh
  - Protected routes
  - Responsive design (mobile-friendly)
  - Form validation with Zod

#### DevOps
- **Docker Support**
  - Multi-stage Dockerfiles for backend and frontend
  - Docker Compose for development, staging, and production
  - Production-ready configuration with nginx

- **CI/CD**
  - GitHub Actions workflow
  - Automated testing
  - Security scanning (Gosec, Trivy)
  - Docker image building and pushing
  - Automated deployment to staging/production

- **Documentation**
  - Swagger/OpenAPI documentation
  - Quick start guide
  - Deployment guide
  - Comprehensive README

### Technical Stack

#### Backend
- Go 1.23 with Gin framework
- PostgreSQL 13 with UUID extension
- Redis 6 for caching and session management
- Gorilla WebSocket for real-time features
- JWT for authentication
- Swagger for API documentation

#### Frontend
- Next.js 14 with App Router
- TypeScript 5
- TailwindCSS for styling
- shadcn/ui components
- Recharts for data visualization
- React Hook Form for forms
- Zod for validation
- Axios for API calls

### Database Schema

**Tables:**
- `users` - User accounts with role-based access
- `products` - Inventory items with stock tracking
- `categories` - Product categorization
- `stock_movements` - Audit trail for inventory changes
- `notifications` - User notifications
- `audit_logs` - System-wide audit trail
- `system_settings` - Configurable system parameters

**Features:**
- UUID primary keys for distributed system compatibility
- JSONB fields for flexible data storage
- Comprehensive indexing for performance
- Automatic updated_at triggers

### Configuration

**Environment Variables:**
- Backend: `.env` file with database, Redis, JWT, and email configuration
- Frontend: `.env.local` with API URL configuration
- Production: `.env.production` with secure secrets

**Default Accounts:**
- Admin: `admin@example.com` / `admin123`
- Staff: `staff@example.com` / `staff123`

### API Endpoints

**Public:**
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password

**Protected:**
- `GET /api/v1/profile` - Get current user profile
- `PUT /api/v1/profile` - Update profile
- `GET /api/v1/products` - List products (with filtering)
- `POST /api/v1/products` - Create product
- `PUT /api/v1/products/:id` - Update product
- `DELETE /api/v1/products/:id` - Delete product
- `POST /api/v1/products/:id/stock` - Update stock level
- `GET /api/v1/stock-movements` - List stock movements
- `GET /api/v1/categories` - List categories
- `GET /api/v1/dashboard/stats` - Get dashboard statistics
- `GET /api/v1/dashboard/alerts` - Get low stock alerts
- `GET /api/v1/notifications` - Get notifications
- `PUT /api/v1/notifications/:id/read` - Mark notification as read

**Admin Only:**
- `GET /api/v1/admin/users` - List all users
- `POST /api/v1/admin/users` - Create user
- `PUT /api/v1/admin/users/:id` - Update user
- `DELETE /api/v1/admin/users/:id` - Delete user
- `GET /api/v1/admin/categories` - Manage categories
- `POST /api/v1/admin/categories` - Create category
- `PUT /api/v1/admin/categories/:id` - Update category
- `DELETE /api/v1/admin/categories/:id` - Delete category
- `GET /api/v1/admin/reports/:type` - Generate reports
- `GET /api/v1/admin/settings` - Get system settings
- `PUT /api/v1/admin/settings` - Update settings

### Known Issues

None at this time.

### Future Enhancements

- [ ] Email notification service integration
- [ ] Two-factor authentication
- [ ] Advanced analytics dashboard
- [ ] Mobile application
- [ ] Barcode/QR code scanning
- [ ] Multi-warehouse support
- [ ] Supplier management
- [ ] Purchase order management
- [ ] Sales order management
- [ ] Integration with e-commerce platforms

---

## [Unreleased]

### Planned
- Initial release with core inventory management features
