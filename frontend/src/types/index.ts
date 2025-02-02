export interface Product {
  id: string
  name: string
  description: string
  price: number
  imageUrl: string
  category: string
  stock: number
  createdAt: string
}

export interface User {
  id: string
  name: string
  email: string
  createdAt: string
}

export interface CartItem {
  product: Product
  quantity: number
}

export interface Order {
  id: string
  userId: string
  items: CartItem[]
  total: number
  status: 'pending' | 'processing' | 'shipped' | 'delivered'
  createdAt: string
}

export interface ApiResponse<T> {
  success: boolean
  data: T
  message?: string
} 