import { httpClient } from '../lib/api-client'
import { Device, Deployment, Fleet, Software } from '../lib/models'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

/**
 * This file contains all the React Query hooks for API interactions.
 * 
 * It centralizes all API access and prepends the '/api' prefix to all endpoints.
 * Always use these hooks instead of using the httpClient directly.
 */

// Keys for query caching
export const QueryKeys = {
  devices: 'devices',
  device: (id: string) => ['devices', id],
  fleets: 'fleets',
  fleet: (id: string) => ['fleets', id],
  software: 'software',
  softwareItem: (id: string) => ['software', id],
  currentUser: 'currentUser',
  deployments: 'deployments',
  deviceDeployments: (deviceId: string) => ['deployments', 'device', deviceId],
  fleetDeployments: (fleetId: string) => ['deployments', 'fleet', fleetId],
  softwareDeployments: (softwareId: string) => ['deployments', 'software', softwareId],
  deploymentCounts: 'deploymentCounts',
}

// ============ DEVICES ============

export function useDevices() {
  // Check for authentication
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: [QueryKeys.devices],
    queryFn: () => httpClient.get<Device[]>('/api/devices'),
    enabled: hasToken,
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

export function useDevice(deviceId: string) {
  // Check for authentication
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: QueryKeys.device(deviceId),
    queryFn: () => httpClient.get<Device>(`/api/devices/${deviceId}`),
    enabled: !!deviceId && hasToken,
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

export function useCreateDevice() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (deviceData: Partial<Device>) =>
      httpClient.post<Device>('/api/devices', deviceData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.devices] })
    },
  })
}

export function useUpdateDevice(deviceId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (deviceData: Partial<Device>) =>
      httpClient.put<Device>(`/api/devices/${deviceId}`, deviceData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.devices] })
      queryClient.invalidateQueries({ queryKey: QueryKeys.device(deviceId) })
    },
  })
}

export function useDeleteDevice() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (deviceId: string) => httpClient.delete(`/api/devices/${deviceId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.devices] })
    },
  })
}

export function useRestartDevice() {
  return useMutation({
    mutationFn: (deviceId: string) =>
      httpClient.post(`/api/devices/${deviceId}/restart`, {}),
  })
}

// ============ FLEETS ============

export function useFleets() {
  // Check for authentication
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: [QueryKeys.fleets],
    queryFn: () => httpClient.get<Fleet[]>('/api/fleets'),
    enabled: hasToken,
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

export function useFleet(fleetId: string) {
  // Check for authentication
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: QueryKeys.fleet(fleetId),
    queryFn: () => httpClient.get<Fleet>(`/api/fleets/${fleetId}`),
    enabled: !!fleetId && hasToken,
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

export function useCreateFleet() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (fleetData: Partial<Fleet>) =>
      httpClient.post<Fleet>('/api/fleets', fleetData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.fleets] })
    },
  })
}

export function useUpdateFleet(fleetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (fleetData: Partial<Fleet>) =>
      httpClient.put<Fleet>(`/api/fleets/${fleetId}`, fleetData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.fleets] })
      queryClient.invalidateQueries({ queryKey: QueryKeys.fleet(fleetId) })
    },
  })
}

export function useDeleteFleet() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (fleetId: string) => httpClient.delete(`/api/fleets/${fleetId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.fleets] })
    },
  })
}

// ============ SOFTWARE ============

export function useSoftware() {
  // Check for authentication
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: [QueryKeys.software],
    queryFn: () => httpClient.get<Software[]>('/api/software'),
    enabled: hasToken,
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

// ============ DEPLOYMENTS ============

export function useDeploymentsBySoftware(softwareId: string) {
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: QueryKeys.softwareDeployments(softwareId),
    queryFn: () => httpClient.get<Deployment[]>(`/api/software/${softwareId}/deployments`),
    enabled: !!softwareId && hasToken,
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

export function useDeploymentsByDevice(deviceId: string) {
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: QueryKeys.deviceDeployments(deviceId),
    queryFn: () => httpClient.get<Deployment[]>(`/api/devices/${deviceId}/deployments`),
    enabled: !!deviceId && hasToken,
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

export function useDeploymentsByFleet(fleetId: string) {
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: QueryKeys.fleetDeployments(fleetId),
    queryFn: () => httpClient.get<Deployment[]>(`/api/fleets/${fleetId}/deployments`),
    enabled: !!fleetId && hasToken,
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

export function useDeploymentCounts() {
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: [QueryKeys.deploymentCounts],
    queryFn: () => 
      httpClient.get<Record<string, number>>('/api/software/deployment-counts'),
    enabled: hasToken,
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

// ============ METRICS ============

export const MetricsQueryKeys = {
  deviceMetrics: (deviceId: string) => ['metrics', 'device', deviceId],
}

export function useDeviceMetrics(deviceId: string) {
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: MetricsQueryKeys.deviceMetrics(deviceId),
    queryFn: () => httpClient.get(`/api/devices/${deviceId}/metrics`),
    enabled: !!deviceId && hasToken,
    refetchInterval: 30000, // Refresh every 30 seconds
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

export function useSoftwareItem(softwareId: string) {
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: QueryKeys.softwareItem(softwareId),
    queryFn: () => httpClient.get<Software>(`/api/software/${softwareId}`),
    enabled: !!softwareId && hasToken,
    // Don't retry auth failures
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('Unauthorized')) {
        return false
      }
      return failureCount < 2
    },
  })
}

export function useCreateSoftware() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (softwareData: Partial<Software>) =>
      httpClient.post<Software>('/api/software', softwareData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.software] })
    },
  })
}

export function useDeleteSoftware() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (softwareId: string) => httpClient.delete(`/api/software/${softwareId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.software] })
    },
  })
}

// ============ AUTH ============

// Authentication interfaces
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  user: any;
  token: string;
}

export function useCurrentUser() {
  // Only attempt to fetch the current user if we have a token
  const hasToken = !!localStorage.getItem('edgetainer_token')
  
  return useQuery({
    queryKey: [QueryKeys.currentUser],
    queryFn: async () => {
      // Double-check token existence before making the request
      // This prevents unnecessary 401 errors
      if (!localStorage.getItem('edgetainer_token')) {
        return null
      }
      
      return await httpClient.get('/api/auth/me')
    },
    // Only run this query if we have a token 
    enabled: hasToken,
    // Retry once in case of network glitches, but don't retry auth failures
    retry: (failureCount, error) => {
      // If we get a 401/403, don't retry
      if (error instanceof Error && 
          (error.message.includes('401') || 
           error.message.includes('403') ||
           error.message.includes('Unauthorized') ||
           error.message.includes('Forbidden'))) {
        return false
      }
      // Otherwise allow one retry for network issues
      return failureCount < 1
    },
    // Silence errors for auth failures
    staleTime: 0,
    refetchOnWindowFocus: false,
  })
}

export function useLogin() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (credentials: LoginRequest) => 
      httpClient.post<LoginResponse>('/api/auth/login', credentials),
    onSuccess: (data) => {
      const { token } = data
      
      // Store the token for future API calls
      httpClient.configure({ apiKey: token })
      
      // Invalidate current user query
      queryClient.invalidateQueries({ queryKey: [QueryKeys.currentUser] })
    }
  })
}

export function useLogout() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: () => httpClient.post('/api/auth/logout', {}),
    onSettled: () => {
      // Clear auth token from HTTP client
      httpClient.configure({ apiKey: undefined })
      
      // Invalidate current user query
      queryClient.invalidateQueries({ queryKey: [QueryKeys.currentUser] })
    }
  })
}

// ============ PROVISIONING ============

// Device provisioning interface
export interface DeviceProvisionRequest {
  name: string
  fleet_id?: string
  description?: string
}

export function useDeviceProvisioning() {
  return useMutation({
    mutationFn: async (request: DeviceProvisionRequest) => {
      try {
        console.log('Creating device provisioning for:', request.name)
        
        // Make a direct request to get the Ignition configuration
        const response = await httpClient.fetchRaw('/api/provision/device', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          },
          body: JSON.stringify(request),
        })
        
        if (!response.ok) {
          // We'll handle the error in a way that doesn't trigger the global error toast
          const errorText = await response.text();
          console.error('Provisioning error:', response.status, errorText.substring(0, 200));
          
          // If we got HTML and status is 200, it might be a routing issue (SPA handler)
          if (response.status === 200 && errorText.includes('<!doctype html>')) {
            throw new Error('Server returned HTML instead of JSON - check your API route');
          } else {
            throw new Error(`Provisioning failed: ${response.statusText}`);
          }
        }
        
        const contentType = response.headers.get('Content-Type') || '';
        
        // Check for proper ignition configuration format
        if (contentType.includes('text/html')) {
          const text = await response.text();
          console.error('Received HTML instead of JSON:', text.substring(0, 200));
          throw new Error('Server returned HTML instead of JSON');
        }
        
        // Get the blob for download
        const blob = await response.blob();
        
        // Quickly validate it's valid JSON
        try {
          const validationText = await new Response(blob.slice(0)).text();
          JSON.parse(validationText); // Will throw if invalid JSON
        } catch (err) {
          console.error('Invalid Ignition configuration format:', err);
          throw new Error('Invalid Ignition configuration format');
        }
        
        return blob;
      } catch (error) {
        // Make sure we log the error properly
        console.error('Device provisioning error:', error);
        throw error;
      }
    },
    // Add proper onError handling
    onError: (error) => {
      toast.error(`Provisioning failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  });
}
