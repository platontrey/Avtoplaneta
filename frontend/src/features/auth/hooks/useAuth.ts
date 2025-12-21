/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState, useEffect, useCallback } from 'react';
import type { User, LoginCredentials, AuthState } from '../types';
import { authApi } from '../api/authApi';

export function useAuth(): AuthState & {
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  checkAuthStatus: () => Promise<void>;
} {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const checkAuthStatus = useCallback(async () => {
    try {
      const userData = await authApi.getCurrentUser();
      setUser({
        ...userData,
        initials: userData.name ? userData.name.split(' ').map(n => n[0]).join('').toUpperCase() : '',
      });
    } catch {
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const login = useCallback(async (credentials: LoginCredentials) => {
    const userData = await authApi.login(credentials);
    console.log('userData from login:', userData);
    const userWithInitials = {
      ...userData,
      initials: userData.name ? userData.name.split(' ').map(n => n[0]).join('').toUpperCase() : '',
    };
    setUser(userWithInitials);
    localStorage.setItem('userId', userData.id.toString());
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } finally {
      setUser(null);
      localStorage.removeItem('userId');
      // Force page reload to clear all client-side state
      // The CSRF token is managed in the API layer, no need to clear it here
      window.location.href = '/login';
    }
  }, []);

  useEffect(() => {
    checkAuthStatus();
  }, [checkAuthStatus]);

  return {
    user,
    isLoading,
    isAuthenticated: !!user,
    login,
    logout,
    checkAuthStatus,
  };
}