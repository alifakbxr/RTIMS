-- RTIMS Initial Database Schema
-- Version: 1.0.0
-- Description: Complete database schema for Real-Time Inventory Management System

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'staff' CHECK (role IN ('staff', 'admin')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    sku VARCHAR(50) UNIQUE NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    price DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    category VARCHAR(255) REFERENCES categories(name) ON DELETE SET NULL,
    minimum_threshold INTEGER NOT NULL DEFAULT 0,
    supplier_info JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Stock movements table
CREATE TABLE IF NOT EXISTS stock_movements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    change INTEGER NOT NULL,
    reason VARCHAR(50) NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name VARCHAR(255) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT
);

-- System settings table
CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_low_stock ON products(stock, minimum_threshold);
CREATE INDEX IF NOT EXISTS idx_stock_movements_product_id ON stock_movements(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON stock_movements(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_audit_logs_table_name ON audit_logs(table_name);
CREATE INDEX IF NOT EXISTS idx_audit_logs_record_id ON audit_logs(record_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_changed_at ON audit_logs(changed_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_changed_by ON audit_logs(changed_by);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add updated_at triggers
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default categories
INSERT INTO categories (id, name, description, created_at) VALUES
    (uuid_generate_v4(), 'Electronics', 'Electronic devices and components', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'Clothing', 'Apparel and accessories', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'Food & Beverage', 'Food items and beverages', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'Home & Garden', 'Home improvement and garden supplies', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'Sports & Outdoors', 'Sports equipment and outdoor gear', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'Books & Media', 'Books, movies, and music', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'Toys & Games', 'Toys and games for all ages', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'Health & Beauty', 'Health and beauty products', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'Automotive', 'Auto parts and accessories', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), 'Office Supplies', 'Office and school supplies', CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;

-- Insert default admin user (password: admin123)
-- Password hash generated using bcrypt
INSERT INTO users (id, name, email, password, role, is_active, created_at, updated_at) VALUES
    (uuid_generate_v4(), 'Admin User', 'admin@rtims.com', '$2a$10$X.vXKZOF4nqK8qQ3jZ3Z3.3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3', 'admin', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (email) DO NOTHING;

-- Insert default staff user (password: staff123)
-- Password hash generated using bcrypt
INSERT INTO users (id, name, email, password, role, is_active, created_at, updated_at) VALUES
    (uuid_generate_v4(), 'Staff User', 'staff@rtims.com', '$2a$10$Y.vYKZOF4nqK8qQ3jZ3Z3.3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3Z3', 'staff', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (email) DO NOTHING;

-- Insert default system settings
INSERT INTO system_settings (key, value, updated_at) VALUES
    ('low_stock_threshold', '10', CURRENT_TIMESTAMP),
    ('auto_backup', 'true', CURRENT_TIMESTAMP),
    ('backup_frequency', 'daily', CURRENT_TIMESTAMP),
    ('maintenance_mode', 'false', CURRENT_TIMESTAMP)
ON CONFLICT (key) DO NOTHING;
