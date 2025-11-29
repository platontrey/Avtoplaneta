/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

// CSRF token management - centralized
let csrfToken: string | null = null;

const fetchCsrfToken = async (): Promise<void> => {
  try {
    console.log('csrf: Fetching CSRF token...');
    const response = await fetch(`${API_BASE_URL}/auth/csrf-token`, {
      credentials: 'include',
    });
    console.log('csrf: CSRF token fetch response status:', response.status);
    if (response.ok) {
      const data = await response.json();
      csrfToken = data.csrf_token;
      console.log('csrf: CSRF token fetched successfully:', csrfToken ? 'present' : 'null');
    } else {
      console.error('csrf: Failed to fetch CSRF token, status:', response.status);
    }
  } catch (error) {
    console.error('csrf: Failed to fetch CSRF token:', error);
  }
};

// Initialize CSRF token on app start
fetchCsrfToken();

// Helper function to get auth headers
export const getAuthHeaders = () => {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  // Add CSRF token if available
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
    console.log('csrf: Including CSRF token in headers');
  } else {
    console.warn('csrf: No CSRF token available for request');
  }

  return headers;
};

// Export for manual token refresh if needed
export const refreshCsrfToken = fetchCsrfToken;