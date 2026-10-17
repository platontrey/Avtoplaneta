/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { User, LoginCredentials } from '../types';
import { getAuthHeaders } from '@/lib/csrf';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

export const authApi = {
  login: async (credentials: LoginCredentials): Promise<User> => {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(credentials),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Не удалось войти: ${errorText}`);
    }

    const data = await response.json();
    return data.user || data;
  },

  adminLogin: async (password: string): Promise<User> => {
    const response = await fetch(`${API_BASE_URL}/auth/admin`, {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify({ password }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Admin login failed: ${errorText}`);
    }

    return response.json();
  },

  logout: async (): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Logout failed: ${response.status}`);
    }

    return response.json();
  },

  getCurrentUser: async (): Promise<User> => {
    const response = await fetch(`${API_BASE_URL}/auth/me`, {
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to get user: ${response.status}`);
    }

    return response.json();
  },

  updateProfile: async (data: { name: string; username?: string; inn?: string }): Promise<User> => {
    const response = await fetch(`${API_BASE_URL}/auth/profile`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const err = await response.json().catch(() => ({ error: 'Не удалось обновить профиль' }));
      throw new Error(err.error || 'Не удалось обновить профиль');
    }

    const res = await response.json();
    return res.user;
  },

  changePassword: async (data: { old_password: string; new_password: string }): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/auth/change-password`, {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const err = await response.json().catch(() => ({ error: 'Не удалось сменить пароль' }));
      throw new Error(err.error || 'Не удалось сменить пароль');
    }
  },

  uploadAvatar: async (file: File): Promise<{ avatar_url: string; user: User }> => {
    const formData = new FormData();
    formData.append('avatar', file);

    const headers: Record<string, string> = {};
    const csrfToken = localStorage.getItem('csrf_token');
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }

    const response = await fetch(`${API_BASE_URL}/auth/avatar`, {
      method: 'POST',
      headers,
      credentials: 'include',
      body: formData,
    });

    if (!response.ok) {
      const err = await response.json().catch(() => ({ error: 'Не удалось загрузить аватар' }));
      throw new Error(err.error || 'Не удалось загрузить аватар');
    }

    return response.json();
  },

  deleteAvatar: async (): Promise<User> => {
    const headers: Record<string, string> = {};
    const csrfToken = localStorage.getItem('csrf_token');
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }

    const response = await fetch(`${API_BASE_URL}/auth/avatar`, {
      method: 'DELETE',
      headers,
      credentials: 'include',
    });

    if (!response.ok) {
      const err = await response.json().catch(() => ({ error: 'Не удалось удалить аватар' }));
      throw new Error(err.error || 'Не удалось удалить аватар');
    }

    const res = await response.json();
    return res.user;
  },
};
