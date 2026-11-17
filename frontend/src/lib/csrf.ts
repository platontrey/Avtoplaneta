/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

// CSRF token management - centralized
let csrfToken: string | null = typeof window !== 'undefined' ? localStorage.getItem('csrf_token') : null;
let pendingFetch: Promise<string | null> | null = null;

export const setCsrfToken = (token: string | null): void => {
  csrfToken = token;
  if (typeof window !== 'undefined') {
    if (token) {
      localStorage.setItem('csrf_token', token);
    } else {
      localStorage.removeItem('csrf_token');
    }
  }
};

export const fetchCsrfToken = async (): Promise<string | null> => {
  if (pendingFetch) {
    return pendingFetch;
  }

  pendingFetch = (async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/csrf-token`, {
        credentials: 'include',
      });
      if (response.ok) {
        const data = await response.json();
        const token = data.csrf_token || null;
        setCsrfToken(token);
        return token;
      }
      console.error('csrf: Failed to fetch CSRF token, status:', response.status);
    } catch (error) {
      console.error('csrf: Failed to fetch CSRF token:', error);
    } finally {
      pendingFetch = null;
    }
    return csrfToken;
  })();

  return pendingFetch;
};

// Гарантирует наличие CSRF токена перед отправкой мутирующих запросов
export const ensureCsrfToken = async (): Promise<string> => {
  if (csrfToken) {
    return csrfToken;
  }
  const token = await fetchCsrfToken();
  return token || '';
};

// Initialize CSRF token on app start
if (typeof window !== 'undefined') {
  void fetchCsrfToken();
}

// Helper function to get auth headers (синхронно, с fallback на localStorage)
export const getAuthHeaders = (): Record<string, string> => {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  };

  const token = csrfToken || (typeof window !== 'undefined' ? localStorage.getItem('csrf_token') : null);
  if (token) {
    headers['X-CSRF-Token'] = token;
  }

  return headers;
};

// Асинхронная версия с гарантией получения токена
export const getAuthHeadersAsync = async (): Promise<Record<string, string>> => {
  const token = await ensureCsrfToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  };
  if (token) {
    headers['X-CSRF-Token'] = token;
  }
  return headers;
};

// Export for manual token refresh if needed
export const refreshCsrfToken = fetchCsrfToken;
