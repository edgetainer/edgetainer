import { toast } from 'sonner'

/**
 * HTTP client for Edgetainer
 * 
 * IMPORTANT: Do not use this client directly in components.
 * Instead, use the hooks in src/hooks/use-api.ts which provide
 * proper React Query integration, caching, and error handling.
 * 
 * This client handles the low-level HTTP requests without any
 * knowledge of specific API routes or business logic.
 */

// Types for connection configuration
export interface HttpClientConfig {
  baseUrl: string
  apiKey?: string
}

// Load initial token from localStorage
const getInitialToken = (): string | undefined => {
  // Check if we're in a browser environment
  if (typeof window !== 'undefined' && window.localStorage) {
    return localStorage.getItem('edgetainer_token') || undefined
  }
  return undefined
}

// Default configuration
// In production, these would be environment variables
const defaultConfig: HttpClientConfig = {
  baseUrl: import.meta.env.VITE_BASE_URL || '',
  apiKey: getInitialToken() || import.meta.env.VITE_API_KEY,
}

/**
 * HTTP client class for making network requests
 */
class HttpClient {
  private config: HttpClientConfig

  constructor(config: HttpClientConfig = defaultConfig) {
    this.config = config
  }

  /**
   * Set the HTTP client configuration
   */
  configure(config: Partial<HttpClientConfig>) {
    this.config = { ...this.config, ...config }
  }

  /**
   * Get the current configuration
   */
  getConfig(): HttpClientConfig {
    return { ...this.config }
  }

  /**
   * Make a fetch request and automatically parse the result
   */
  async fetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.config.baseUrl}${endpoint}`
    const headers = new Headers(options.headers)

    // Add Authorization header if we have an API key
    if (this.config.apiKey) {
      headers.set('Authorization', `Bearer ${this.config.apiKey}`)
    }

    // Add Content-Type header if not present
    if (!headers.has('Content-Type') && options.method !== 'GET') {
      headers.set('Content-Type', 'application/json')
    }

    try {
      // Make the request
      const response = await fetch(url, {
        ...options,
        headers,
      })

      // Check if the response is OK
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        console.error('HTTP error:', response.status, response.statusText, errorData)
        throw new Error(
          errorData.message || `Request failed: ${response.statusText}`,
        )
      }

      // Check if the response is empty
      const contentLength = response.headers.get('Content-Length')
      if (contentLength === '0') {
        return {} as T
      }

      // Parse the response
      const contentType = response.headers.get('Content-Type') || ''
      if (contentType.includes('application/json')) {
        return await response.json()
      }

      return (await response.text()) as unknown as T
    } catch (error) {
      console.error('API request failed:', error)
      
      // Only show toast for actual network/server errors, not for auth errors or debug logging
      const isAuthError = error instanceof Error && 
        (error.message.includes('Unauthorized') || 
         error.message.includes('Forbidden') ||
         error.message.includes('401') || 
         error.message.includes('403'));
      
      const isDebugLog = error instanceof Error && error.message.includes('debug log');
      
      // Don't show toast for auth errors or debug logs
      if (!isAuthError && !isDebugLog) {
        toast.error(
          `API request failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
        )
      } else if (isAuthError) {
        // Just log auth errors to console without showing toast
        console.info('Authentication required:', error.message);
      }
      
      throw error
    }
  }

  /**
   * Make a fetch request and return the raw Response object
   * Useful for downloading files or getting headers directly
   */
  async fetchRaw(endpoint: string, options: RequestInit = {}): Promise<Response> {
    const url = `${this.config.baseUrl}${endpoint}`
    const headers = new Headers(options.headers)

    // Add Authorization header if we have an API key
    if (this.config.apiKey) {
      headers.set('Authorization', `Bearer ${this.config.apiKey}`)
    }

    try {
      // Make the request
      const response = await fetch(url, {
        ...options,
        headers,
      })

      // Only log errors but don't throw, let the caller handle the raw response
      if (!response.ok) {
        // Don't warn about auth errors, as they're expected in some flows
        const isAuthError = response.status === 401 || response.status === 403;
        
        if (isAuthError) {
          console.info(`Auth required (${url}):`, response.status, response.statusText);
        } else {
          console.warn(`HTTP warning (${url}):`, response.status, response.statusText);
        }
      }

      return response
    } catch (error) {
      console.error('Network error:', error)
      
      // Only show toast for actual network errors, not for auth errors
      const isAuthError = error instanceof Error && 
        (error.message.includes('Unauthorized') || 
         error.message.includes('Forbidden') ||
         error.message.includes('401') || 
         error.message.includes('403'));
          
      if (!isAuthError) {
        toast.error(
          `Network error: ${error instanceof Error ? error.message : 'Connection failed'}`,
        )
      } else {
        // Just log auth errors to console
        console.info('Authentication required:', error.message);
      }
      
      throw error
    }
  }

  /**
   * Convenience methods for HTTP verbs
   */
  get<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    return this.fetch<T>(endpoint, { ...options, method: 'GET' })
  }

  post<T>(endpoint: string, data: any, options: RequestInit = {}): Promise<T> {
    return this.fetch<T>(endpoint, {
      ...options,
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  put<T>(endpoint: string, data: any, options: RequestInit = {}): Promise<T> {
    return this.fetch<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  delete<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    return this.fetch<T>(endpoint, { ...options, method: 'DELETE' })
  }
}

// Export a singleton instance
export const httpClient = new HttpClient()

// Export the class for testing/mocking
export default HttpClient
