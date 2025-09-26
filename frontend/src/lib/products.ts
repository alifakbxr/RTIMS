import { Product, ProductFilter, CreateProductRequest, UpdateProductRequest, PaginatedResponse } from '@/types'
import { api } from './api'

export const productsApi = {
  // Get all products with optional filtering
  async getProducts(filters?: ProductFilter): Promise<PaginatedResponse<Product>> {
    const params = new URLSearchParams()

    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          params.append(key, value.toString())
        }
      })
    }

    const response = await api.get(`/products?${params}`)
    return response.data
  },

  // Get a single product by ID
  async getProduct(id: string): Promise<Product> {
    const response = await api.get(`/products/${id}`)
    return response.data
  },

  // Create a new product
  async createProduct(product: CreateProductRequest): Promise<Product> {
    const response = await api.post('/products', product)
    return response.data
  },

  // Update an existing product
  async updateProduct(id: string, product: UpdateProductRequest): Promise<Product> {
    const response = await api.put(`/products/${id}`, product)
    return response.data
  },

  // Delete a product
  async deleteProduct(id: string): Promise<void> {
    await api.delete(`/products/${id}`)
  },

  // Update product stock
  async updateStock(productId: string, change: number, reason: string, notes?: string): Promise<void> {
    await api.post(`/products/${productId}/stock`, {
      change,
      reason,
      notes
    })
  },

  // Get stock movements
  async getStockMovements(productId?: string): Promise<PaginatedResponse<unknown>> {
    const params = productId ? `?product_id=${productId}` : ''
    const response = await api.get(`/stock-movements${params}`)
    return response.data
  }
}