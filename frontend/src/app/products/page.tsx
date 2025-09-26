'use client'

import { useEffect, useState } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Navigation } from '@/components/ui/navigation'
import { Product, ProductFilter, Category } from '@/types'
import { Search, Plus, Edit, Trash2, Package, AlertTriangle } from 'lucide-react'
import { ProductModal } from '@/components/ProductModal'
import { productsApi } from '@/lib/products'
import { categoriesApi } from '@/lib/categories'

export default function ProductsPage() {
  const { user } = useAuth()
  const [products, setProducts] = useState<Product[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('')
  const [showLowStockOnly, setShowLowStockOnly] = useState(false)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingProduct, setEditingProduct] = useState<Product | null>(null)

  useEffect(() => {
    fetchProducts()
    fetchCategories()
  }, [searchTerm, selectedCategory, showLowStockOnly])

  const fetchCategories = async () => {
    try {
      const categoriesData = await categoriesApi.getCategories()
      setCategories(categoriesData)
    } catch (error) {
      console.error('Failed to fetch categories:', error)
    }
  }

  const fetchProducts = async () => {
    try {
      setLoading(true)
      const filters: ProductFilter = {}
      if (searchTerm) filters.search = searchTerm
      if (selectedCategory) filters.category = selectedCategory
      if (showLowStockOnly) filters.low_stock_only = true

      const data = await productsApi.getProducts(filters)
      setProducts(data.data || [])
    } catch (error) {
      console.error('Failed to fetch products:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateProduct = () => {
    setEditingProduct(null)
    setIsModalOpen(true)
  }

  const handleEditProduct = (product: Product) => {
    setEditingProduct(product)
    setIsModalOpen(true)
  }

  const handleDeleteProduct = async (productId: string) => {
    if (!confirm('Are you sure you want to delete this product?')) return

    try {
      await productsApi.deleteProduct(productId)
      fetchProducts()
    } catch (error) {
      console.error('Failed to delete product:', error)
    }
  }

  const handleModalClose = () => {
    setIsModalOpen(false)
    setEditingProduct(null)
    fetchProducts()
  }

  if (!user) {
    return <div>Please log in to access products.</div>
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="lg:pl-64">
        <header className="bg-white border-b border-gray-200 shadow-sm">
          <div className="px-4 sm:px-6 lg:px-8">
            <div className="flex items-center justify-between py-6">
              <div>
                <h1 className="text-3xl font-bold text-gray-900">Products</h1>
                <p className="mt-1 text-sm text-gray-600">Manage your inventory</p>
              </div>
              <Button
                onClick={handleCreateProduct}
                className="flex items-center gap-2 px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700"
              >
                <Plus className="w-4 h-4" />
                Add Product
              </Button>
            </div>
          </div>
        </header>

        <main className="px-4 py-8 sm:px-6 lg:px-8">
          {/* Filters */}
          <Card className="mb-6">
            <CardHeader>
              <CardTitle className="text-lg">Filters</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                <div className="relative">
                  <Search className="absolute w-4 h-4 text-gray-400 transform -translate-y-1/2 left-3 top-1/2" />
                  <Input
                    placeholder="Search products..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                  />
                </div>
                <select
                  value={selectedCategory}
                  onChange={(e) => setSelectedCategory(e.target.value)}
                  className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="">All Categories</option>
                  {categories.map((category) => (
                    <option key={category.id} value={category.name}>{category.name}</option>
                  ))}
                </select>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={showLowStockOnly}
                    onChange={(e) => setShowLowStockOnly(e.target.checked)}
                    className="text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                  />
                  Show low stock only
                </label>
              </div>
            </CardContent>
          </Card>

          {/* Products Grid */}
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="w-8 h-8 border-b-2 border-blue-600 rounded-full animate-spin"></div>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {products.map((product) => (
                <Card key={product.id} className="transition-shadow duration-200 hover:shadow-lg">
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <CardTitle className="text-lg font-semibold text-gray-900 line-clamp-2">
                          {product.name}
                        </CardTitle>
                        <CardDescription className="mt-1 text-sm text-gray-600">
                          SKU: {product.sku}
                        </CardDescription>
                      </div>
                      <Badge
                        variant={product.stock <= product.minimum_threshold ? 'destructive' : 'secondary'}
                        className="ml-2"
                      >
                        {product.stock <= product.minimum_threshold && <AlertTriangle className="w-3 h-3 mr-1" />}
                        {product.stock}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-3">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Category:</span>
                        <Badge variant="outline">{product.category}</Badge>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Price:</span>
                        <span className="font-semibold text-gray-900">${product.price}</span>
                      </div>
                      {product.stock <= product.minimum_threshold && (
                        <div className="flex items-center gap-1 text-sm text-red-600">
                          <AlertTriangle className="w-4 h-4" />
                          Low stock warning
                        </div>
                      )}
                      <div className="flex gap-2 pt-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleEditProduct(product)}
                          className="flex-1"
                        >
                          <Edit className="w-4 h-4 mr-1" />
                          Edit
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleDeleteProduct(product.id)}
                          className="flex-1 text-red-600 hover:text-red-700 hover:bg-red-50"
                        >
                          <Trash2 className="w-4 h-4 mr-1" />
                          Delete
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}

          {!loading && products.length === 0 && (
            <div className="py-12 text-center">
              <Package className="w-12 h-12 mx-auto text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900">No products</h3>
              <p className="mt-1 text-sm text-gray-500">Get started by creating a new product.</p>
              <div className="mt-6">
                <Button onClick={handleCreateProduct}>
                  <Plus className="w-4 h-4 mr-2" />
                  Add Product
                </Button>
              </div>
            </div>
          )}
        </main>
      </div>

      <ProductModal
        isOpen={isModalOpen}
        onClose={handleModalClose}
        product={editingProduct}
      />
    </div>
  )
}