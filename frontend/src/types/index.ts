// User types
export type UserRole = 'staff' | 'admin'

export interface User {
  id: string
  name: string
  email: string
  role: UserRole
  created_at: string
  updated_at: string
  is_active: boolean
}

export interface CreateUserRequest {
  name: string
  email: string
  password: string
  role: UserRole
}

export interface UpdateUserRequest {
  name?: string
  email?: string
  role?: UserRole
  is_active?: boolean
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
}

export interface AuthResponse {
  user: User
  access_token: string
  token_type: string
  expires_in: number
}

export interface RefreshTokenRequest {
  refresh_token: string
}

// Product types
export interface Product {
  id: string
  name: string
  sku: string
  stock: number
  price: number
  category: string
  minimum_threshold: number
  supplier_info?: string
  created_at: string
  updated_at: string
}

export interface CreateProductRequest {
  name: string
  sku: string
  stock: number
  price: number
  category: string
  minimum_threshold: number
  supplier_info?: string
}

export interface UpdateProductRequest {
  name?: string
  sku?: string
  stock?: number
  price?: number
  category?: string
  minimum_threshold?: number
  supplier_info?: string
}

export interface ProductFilter {
  search?: string
  category?: string
  min_stock?: number
  max_stock?: number
  min_price?: number
  max_price?: number
  low_stock_only?: boolean
  page?: number
  limit?: number
  sort_by?: string
  sort_order?: string
}

// Stock movement types
export type MovementReason = 'purchase' | 'sale' | 'adjustment' | 'return' | 'damage' | 'transfer'

export interface StockMovement {
  id: string
  product_id: string
  change: number
  reason: MovementReason
  created_by: string
  created_at: string
  notes?: string
}

export interface CreateStockMovementRequest {
  product_id: string
  change: number
  reason: MovementReason
  notes?: string
}

export interface StockMovementFilter {
  product_id?: string
  reason?: MovementReason
  start_date?: string
  end_date?: string
  page?: number
  limit?: number
  sort_by?: string
  sort_order?: string
}

// Category types
export interface Category {
  id: string
  name: string
  description?: string
  created_at: string
}

export interface CreateCategoryRequest {
  name: string
  description?: string
}

export interface UpdateCategoryRequest {
  name?: string
  description?: string
}

// Notification types
export type NotificationType = 'low_stock' | 'system' | 'user'

export interface Notification {
  id: string
  user_id: string
  message: string
  type: NotificationType
  is_read: boolean
  created_at: string
}

export interface CreateNotificationRequest {
  user_id: string
  message: string
  type: NotificationType
}

export interface NotificationFilter {
  user_id?: string
  type?: NotificationType
  is_read?: boolean
  page?: number
  limit?: number
  sort_by?: string
  sort_order?: string
}

// Audit log types
export type AuditAction = 'create' | 'update' | 'delete' | 'login' | 'logout' | 'view'

export interface AuditLog {
  id: string
  table_name: string
  record_id: string
  action: AuditAction
  old_values?: Record<string, unknown>
  new_values?: Record<string, unknown>
  changed_by: string
  changed_at: string
  ip_address?: string
  user_agent?: string
}

export interface CreateAuditLogRequest {
  table_name: string
  record_id: string
  action: AuditAction
  old_values?: Record<string, unknown>
  new_values?: Record<string, unknown>
  changed_by: string
  ip_address?: string
  user_agent?: string
}

export interface AuditLogFilter {
  table_name?: string
  action?: AuditAction
  changed_by?: string
  start_date?: string
  end_date?: string
  page?: number
  limit?: number
  sort_by?: string
  sort_order?: string
}

// API Response types
export interface ApiResponse<T> {
  data?: T
  message?: string
  error?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  pagination: {
    page: number
    limit: number
    total: number
    pages: number
  }
}

// Dashboard types
export interface DashboardStats {
  total_products: number
  low_stock_count: number
  total_users: number
  total_movements_today: number
}

export interface StockAlert {
  id: string
  name: string
  sku: string
  stock: number
  minimum_threshold: number
  type: 'low_stock_alert'
}

export interface SystemStatus {
  total_products: number
  low_stock_count: number
  total_users: number
  server_time: string
}

// WebSocket message types
export interface WebSocketMessage {
  type: 'stock_change' | 'notification' | 'system_status' | 'stock_update'
  data?: Record<string, unknown>
  message?: string
  timestamp?: string
}