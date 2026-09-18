/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';

import type { LoginCredentials, AuthState } from '../types';
import { authApi } from '../api/authApi';
import { AUTH_ME_QUERY_KEY, useCurrentUser } from './useCurrentUser';

/**
 * Тот же общий запрос текущего пользователя, что и в hooks/useAuth: обе версии
 * хука делят один ключ react-query, поэтому на страницу приходится ровно один
 * запрос /auth/me независимо от того, сколько компонентов спросили о нём.
 */
export function useAuth(): AuthState & {
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  checkAuthStatus: () => Promise<void>;
} {
  const queryClient = useQueryClient();
  const { data: user = null, isLoading } = useCurrentUser();

  const checkAuthStatus = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: AUTH_ME_QUERY_KEY });
  }, [queryClient]);

  const login = useCallback(
    async (credentials: LoginCredentials) => {
      const userData = await authApi.login(credentials);
      queryClient.setQueryData(AUTH_ME_QUERY_KEY, {
        ...userData,
        initials: userData.name
          ? userData.name
              .split(' ')
              .map((n) => n[0])
              .join('')
              .toUpperCase()
          : '',
      });
      localStorage.setItem('userId', userData.id.toString());
    },
    [queryClient],
  );

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } finally {
      queryClient.setQueryData(AUTH_ME_QUERY_KEY, null);
      localStorage.removeItem('userId');
      // Force page reload to clear all client-side state
      // The CSRF token is managed in the API layer, no need to clear it here
      window.location.href = '/login';
    }
  }, [queryClient]);

  return {
    user,
    isLoading,
    isAuthenticated: !!user,
    login,
    logout,
    checkAuthStatus,
  };
}
