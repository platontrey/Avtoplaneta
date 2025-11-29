/*
 *  Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { Part, StatisticsResponse } from './types.ts';
import { getAuthHeaders } from './csrf';

// Central configuration for API endpoints
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
export const ORDERS_API_URL = 'http://localhost:8082';

// Parts API
export const partsApi = {
  // Get all parts
  getAll: async (): Promise<Part[]> => {
    const response = await fetch(`${API_BASE_URL}/api/inventory`, {
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch parts: ${response.status}`);
    }

    const data = await response.json();
    return data.parts || [];
  },

  // Add new part
  create: async (partData: Omit<Part, 'id'>): Promise<Part> => {
    console.log('partsApi.create: Отправка запроса на', `${API_BASE_URL}/api/addpart`, 'с данными:', partData);
    const response = await fetch(`${API_BASE_URL}/api/addpart`, {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(partData),
    });

    console.log('partsApi.create: Получен ответ, статус:', response.status);
    if (!response.ok) {
      const errorText = await response.text();
      console.error('partsApi.create: Ошибка ответа:', errorText);
      throw new Error(`Failed to create part: ${errorText}`);
    }

    const result = await response.json();
    console.log('partsApi.create: Успешный ответ:', result);
    return result;
  },

  // Update part
  update: async (id: number, partData: Partial<Part>): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/updatepart/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(partData),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to update part: ${errorText}`);
    }
  },

  // Delete part
  delete: async (id: number): Promise<void> => {
    console.log(`partsApi.delete: Deleting part with id ${id}`);
    const response = await fetch(`${API_BASE_URL}/api/deletepart/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
      credentials: 'include',
    });

    console.log(`partsApi.delete: Response status: ${response.status}`);
    if (!response.ok) {
      const errorText = await response.text();
      console.error(`partsApi.delete: Failed to delete part: ${errorText}`);
      throw new Error(`Failed to delete part: ${errorText}`);
    }
    console.log(`partsApi.delete: Successfully deleted part ${id}`);
  },

  // Upload part photo
  uploadPhoto: async (id: number, file: File): Promise<{ photo: string; filename: string; size: number }> => {
    const formData = new FormData();
    formData.append('photo', file);

    const response = await fetch(`${API_BASE_URL}/api/uploadpartphoto/${id}`, {
      method: 'POST',
      headers: {
        'X-CSRF-Token': getAuthHeaders()['X-CSRF-Token'] || '',
      },
      credentials: 'include',
      body: formData,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to upload photo: ${errorText}`);
    }

    const result = await response.json();
    if (!result.photo) {
      throw new Error('Upload response missing photo URL');
    }

    return {
      photo: result.photo,
      filename: result.filename,
      size: result.size,
    };
  },
};

// Statistics API
export const statisticsApi = {
  get: async (): Promise<StatisticsResponse> => {
    const response = await fetch(`${API_BASE_URL}/api/statistics`, {
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
  chat: async (message: string, context?: Record<string, any>) => {
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

    return response.json();
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