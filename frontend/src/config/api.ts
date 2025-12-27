// API configuration for different environments
const getBaseURL = () => {
  // In Docker/production, use relative paths to leverage nginx proxy
  // In development, use environment variable if available, fallback to localhost
  if (import.meta.env.PROD) {
    return '/api/v1';
  }
  return import.meta.env.VITE_SERVER_URL || 'http://localhost:8080/api/v1';
};

export const API_BASE_URL = getBaseURL();