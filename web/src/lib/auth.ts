import { createContext, useContext } from 'react'

// User type definition
export interface User {
  id: string
  username: string
  email?: string
  role: 'admin' | 'operator' | 'viewer'
}

// Auth context interface
export interface AuthContextProps {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  error: string | null
}

// Create context with default values
export const AuthContext = createContext<AuthContextProps>({
  user: null,
  isLoading: false,
  isAuthenticated: false,
  login: async () => {},
  logout: () => {},
  error: null,
})

// Hook for accessing the auth context
export const useAuth = () => useContext(AuthContext)
