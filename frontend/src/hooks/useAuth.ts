import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

import type { User } from '../features/auth/types';
import { logUserActivity } from '../features/admin/api/adminApi';
import { API_BASE_URL } from '@/lib/api';

export const useAuth = () => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/me`, {
        credentials: 'include',
      });
      if (response.ok) {
        const userData = await response.json();
        setUser(userData);
      } else {
        setUser(null);
      }
    } catch (error) {
      console.log('Not authenticated', error);
      setUser(null);
    } finally {
      setLoading(false);
    }
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