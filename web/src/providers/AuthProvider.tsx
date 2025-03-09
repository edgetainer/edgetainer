import { useState, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { AuthContext } from '../lib/auth'
import { httpClient } from '../lib/api-client'
import { useLogin, useLogout, useCurrentUser, LoginRequest } from '../hooks/use-api'

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<any | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()
  
  // Use our centralized hooks for login/logout functionality
  const loginMutation = useLogin()
  const logoutMutation = useLogout()
  const { data: currentUser } = useCurrentUser()

  // Update user state when currentUser query changes
  useEffect(() => {
    if (currentUser) {
      setUser(currentUser)
    }
  }, [currentUser])
  
  // Check for existing session on initial load
  useEffect(() => {
    const checkAuth = async () => {
      try {
        const storedUser = localStorage.getItem('edgetainer_user')
        const storedToken = localStorage.getItem('edgetainer_token')
        
        if (storedUser && storedToken) {
          // Restore API token
          httpClient.configure({ apiKey: storedToken })
          
          // User data will be fetched by the useCurrentUser hook
          setUser(JSON.parse(storedUser))
        }
      } catch (error) {
        console.error('Failed to restore session:', error)
        localStorage.removeItem('edgetainer_user')
        localStorage.removeItem('edgetainer_token')
      } finally {
        setIsLoading(false)
      }
    }
    
    checkAuth()
  }, [])

  // Login function that also handles local storage
  const login = async (username: string, password: string) => {
    setIsLoading(true)
    try {
      const credentials: LoginRequest = { username, password };
      const response = await loginMutation.mutateAsync(credentials)
      
      // Store auth data in localStorage
      localStorage.setItem('edgetainer_user', JSON.stringify(response.user))
      localStorage.setItem('edgetainer_token', response.token)
      
      // Update state
      setUser(response.user)
      setError(null)
      
      // Navigate to dashboard
      navigate({ to: '/' })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Authentication failed')
      console.error('Login failed:', err)
    } finally {
      setIsLoading(false)
    }
  }

  // Logout function that also handles local storage and navigation
  const logout = () => {
    // Clear local storage
    localStorage.removeItem('edgetainer_user')
    localStorage.removeItem('edgetainer_token')
    
    // Execute logout mutation
    logoutMutation.mutate()
    
    // Update state and navigate
    setUser(null)
    navigate({ to: '/login' })
  }

  return (
    <AuthContext.Provider value={{
      user,
      isLoading: isLoading || loginMutation.isPending || logoutMutation.isPending,
      isAuthenticated: !!user,
      login,
      logout,
      error,
    }}>
      {children}
    </AuthContext.Provider>
  )
}
