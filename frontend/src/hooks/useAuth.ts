import { useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';

import type { User } from '../features/auth/types';
import { logUserActivity } from '../features/admin/api/adminApi';
import { AUTH_ME_QUERY_KEY, useCurrentUser } from '../features/auth/hooks/useCurrentUser';
import { API_BASE_URL } from '@/lib/api';

/**
 * Пользователь берётся из общего запроса react-query, а не из локального
 * состояния хука. Это важно: useAuth вызывается в том числе из PartBlock,
 * который рисуется на каждую строку инвентаря.
 */
export const useAuth = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: user = null, isLoading: loading } = useCurrentUser();

  const setUser = (value: User | null) => {
    queryClient.setQueryData(AUTH_ME_QUERY_KEY, value);
  };

  const checkAuthStatus = async () => {
    await queryClient.invalidateQueries({ queryKey: AUTH_ME_QUERY_KEY });
  };

  const handleLogin = (userData: User) => {
    setUser(userData);
    // Логируем вход пользователя
    logUserActivity({
      action: 'login',
      resource_type: 'system',
      details: `Пользователь ${userData.name} (${userData.email}) вошел в систему`,
    }).catch(console.warn);
    navigate('/');
  };

  const handleLogout = async () => {
    try {
      // Логируем выход пользователя перед logout
      if (user) {
        await logUserActivity({
          action: 'logout',
          resource_type: 'system',
          details: `Пользователь ${user.name} (${user.email}) вышел из системы`,
        }).catch(console.warn);
      }

      await fetch(`${API_BASE_URL}/auth/logout`, {
        method: 'POST',
        credentials: 'include',
      });
      setUser(null);
      if (window.localStorage) {
        localStorage.removeItem('csrf_token');
      }
      window.location.href = '/login';
    } catch (error) {
      console.error('Logout failed:', error);
      setUser(null);
      window.location.href = '/login';
    }
  };

  return { user, loading, handleLogin, handleLogout, checkAuthStatus };
};
