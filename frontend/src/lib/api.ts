interface ApiResponse<T = unknown> {
  data: T;
  status: number;
}

class ApiError extends Error {
  status: number;
  
  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

class ApiClient {
  private baseURL: string;
  private refreshTokenFn?: () => Promise<boolean>;
  private getTokenFn?: () => string;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  // Set the refresh token function from AuthContext
  setRefreshTokenFn(refreshTokenFn: () => Promise<boolean>) {
    this.refreshTokenFn = refreshTokenFn;
  }

  // Set the get token function from AuthContext
  setGetTokenFn(getTokenFn: () => string) {
    this.getTokenFn = getTokenFn;
  }

  private async makeRequest<T>(
    endpoint: string,
    options: RequestInit = {},
    retry = true
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;
    
    // Add authorization header if token is available
    const token = this.getTokenFn?.();
    if (token) {
      options.headers = {
        ...options.headers,
        'Authorization': token,
      };
    }

    // Ensure credentials are included for refresh token cookies
    options.credentials = options.credentials || 'include';

    try {
      const response = await fetch(url, options);

      // If unauthorized and we haven't retried yet, try to refresh token
      if (response.status === 401 && retry && this.refreshTokenFn) {
        console.log('Access token expired, attempting refresh...');
        const refreshSuccess = await this.refreshTokenFn();
        
        if (refreshSuccess) {
          console.log('Token refresh successful, retrying request...');
          // Retry the request with the new token (retry = false to prevent infinite loop)
          return this.makeRequest<T>(endpoint, options, false);
        } else {
          throw new ApiError('Authentication failed', 401);
        }
      }

      if (!response.ok) {
        let errorMessage = `HTTP error! status: ${response.status}`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || errorData.message || errorMessage;
        } catch {
          // If parsing error response fails, use default message
        }
        throw new ApiError(errorMessage, response.status);
      }

      const data = await response.json();
      return { data, status: response.status };
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError(error instanceof Error ? error.message : 'Network error', 0);
    }
  }

  // GET request
  async get<T>(endpoint: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
    return this.makeRequest<T>(endpoint, { ...options, method: 'GET' });
  }

  // POST request
  async post<T>(
    endpoint: string, 
    data?: unknown, 
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    return this.makeRequest<T>(endpoint, {
      ...options,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  // PUT request
  async put<T>(
    endpoint: string, 
    data?: unknown, 
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    return this.makeRequest<T>(endpoint, {
      ...options,
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  // DELETE request
  async delete<T>(endpoint: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
    return this.makeRequest<T>(endpoint, { ...options, method: 'DELETE' });
  }
}

// Create a singleton instance
const apiClient = new ApiClient(import.meta.env.VITE_SERVER_URL + '/v1');

export { apiClient, ApiError };
export type { ApiResponse };
