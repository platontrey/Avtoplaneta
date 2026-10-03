/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';

import type { LoginCredentials, AuthState } from '../types';
import { authApi } from '../api/authApi';
import { logUserActivity } from '@/features/admin/api/adminApi';
import { AUTH_ME_QUERY_KEY, useCurrentUser } from './useCurrentUser';

export interface UseAuthReturn extends AuthState {
  loading: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  handleLogout: () => Promise<void>;
  checkAuthStatus: () => Promise<void>;
}

/**
 * Единый хук управления состоянием аутентификации пользователя.
 * Использует общий кэш React Query (AUTH_ME_QUERY_KEY) и фиксирует события в журнале аудита.
 */
export function useAuth(): UseAuthReturn {
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

      // Логируем успешный вход пользователя
      logUserActivity({
        action: 'login',
        resource_type: 'system',
        details: `Пользователь ${userData.name} (${userData.email}) вошел в систему`,
      }).catch(console.warn);
    },
    [queryClient],
  );

  const logout = useCallback(async () => {
    try {
      if (user) {
        await logUserActivity({
          action: 'logout',
          resource_type: 'system',
          details: `Пользователь ${user.name} (${user.email}) вышел из системы`,
        }).catch(console.warn);
      }
      await authApi.logout();
    } catch (error) {
      console.error('Logout failed:', error);
    } finally {
      queryClient.setQueryData(AUTH_ME_QUERY_KEY, null);
      localStorage.removeItem('userId');
      localStorage.removeItem('csrf_token');
      window.location.href = '/login';
    }
  }, [queryClient, user]);

  return {
    user,
    isLoading,
    loading: isLoading,
    isAuthenticated: !!user,
    login,
    logout,
    handleLogout: logout,
    checkAuthStatus,
  };
}
