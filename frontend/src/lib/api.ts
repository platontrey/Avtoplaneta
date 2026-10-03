/*
 *  Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { Part, StatisticsResponse } from './types.ts';
import { getAuthHeaders } from './csrf';

import { partsApi } from '@/features/parts/api/partsApi';

// Central configuration for API endpoints
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';
export const ORDERS_API_URL = import.meta.env.VITE_API_BASE_URL || '';

// Re-export unified partsApi
export { partsApi };

// Statistics API
export const statisticsApi = {
  get: async (): Promise<StatisticsResponse> => {
    const response = await fetch(`${API_BASE_URL}/api/v1/statistics`, {
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch statistics: ${response.status}`);
    }

    return response.json();
  },
};

// AI Agent API
export const aiAgentApi = {
  chat: async (message: string, context?: Record<string, unknown>) => {
    const response = await fetch(`${API_BASE_URL}/api/ai-agent/chat`, {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify({
        message,
        context: context || {},
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`AI Agent chat failed: ${errorText}`);
    }

    return response.json();
  },

  // CRUD operations for parts via AI
  addPart: async (partData: Omit<Part, 'id'>) => {
    console.log('aiAgentApi.addPart: Вызван с данными:', partData);
    await partsApi.create(partData);
    console.log('aiAgentApi.addPart: Завершен успешно');
  },

  updatePart: async (id: number, partData: Partial<Part>) => {
    return partsApi.update(id, partData);
  },

  deletePart: async (id: number) => {
    return partsApi.delete(id);
  },

  // Search parts for AI operations
  searchParts: async (query: string): Promise<Part[]> => {
    // Используем существующий API инвентаря с фильтром
    const allParts = await partsApi.getAll();
    if (!query) return allParts;

    // Простой поиск по названию и описанию
    const lowerQuery = query.toLowerCase();
    return allParts.filter(part =>
      part.name.toLowerCase().includes(lowerQuery) ||
      (part.description && part.description.toLowerCase().includes(lowerQuery)) ||
      (part.category && part.category.toLowerCase().includes(lowerQuery)) ||
      (part.brand && part.brand.toLowerCase().includes(lowerQuery)) ||
      (part.model && part.model.toLowerCase().includes(lowerQuery))
    );
  },
};

// Auth API
export const authApi = {
  login: async (credentials: { email: string; password: string }) => {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(credentials),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Login failed: ${errorText}`);
    }

    const data = await response.json();
    return data.user || data;
  },


  logout: async () => {
    const response = await fetch(`${API_BASE_URL}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Logout failed: ${response.status}`);
    }

    return response.json();
  },

  getCurrentUser: async () => {
    const response = await fetch(`${API_BASE_URL}/auth/me`, {
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to get user: ${response.status}`);
    }

    return response.json();
  },
};
