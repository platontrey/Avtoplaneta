/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { Part } from '../types';
import { getAuthHeaders } from '@/lib/csrf';
import { logUserActivity } from '../../admin/api/adminApi';

const API_BASE_URL = 'http://localhost:8080';

export const partsApi = {
  // Get all parts with optional filters
  getAll: async (filters?: {
    search?: string;
    category?: string;
    brand?: string;
    model?: string;
    location?: string;
    salesman?: string;
    status?: string;
    hasPhoto?: string;
    limit?: number;
    page?: number;
  }): Promise<Part[]> => {
    console.log('partsApi.getAll called with filters:', filters);
    const params = new URLSearchParams();
    if (filters?.search) params.append('search', filters.search);
    if (filters?.category) params.append('category', filters.category);
    if (filters?.brand) params.append('brand', filters.brand);
    if (filters?.model) params.append('model', filters.model);
    if (filters?.location) params.append('location', filters.location);
    if (filters?.salesman) params.append('salesman', filters.salesman);
    if (filters?.status) params.append('status', filters.status);
    if (filters?.hasPhoto) params.append('hasPhoto', filters.hasPhoto);
    if (filters?.limit && filters.limit > 0) {
      params.append('limit', filters.limit.toString());
    }
    if (filters?.page && filters.page > 0) {
      params.append('page', filters.page.toString());
    }

    const url = `${API_BASE_URL}/api/inventory${params.toString() ? '?' + params.toString() : ''}`;
    console.log('partsApi.getAll: Fetching URL:', url);

    const response = await fetch(url, {
      credentials: 'include',
    });

    console.log('partsApi.getAll: Response status:', response.status);
    if (!response.ok) {
      throw new Error(`Failed to fetch parts: ${response.status}`);
    }

    const data = await response.json();
    // Backend returns array directly, not wrapped in {parts: [...]}
    return Array.isArray(data) ? data : [];
  },

  // Add new part
  create: async (partData: Omit<Part, 'id'>): Promise<Part> => {
    console.log('partsApi.create: Attempting to create part');
    const headers = getAuthHeaders();
    console.log('partsApi.create: Headers being sent:', headers);
    const response = await fetch(`${API_BASE_URL}/api/addpart`, {
      method: 'POST',
      headers,
      credentials: 'include',
      body: JSON.stringify(partData),
    });

    console.log('partsApi.create: Response status:', response.status);
    if (!response.ok) {
      const errorText = await response.text();
      console.error('partsApi.create: Failed to create part:', errorText);
      throw new Error(`Failed to create part: ${errorText}`);
    }

    const result = await response.json();
    console.log('partsApi.create: Part created successfully');

    // Логируем создание запчасти
    logUserActivity({
      action: 'create_part',
      resource_type: 'part',
      resource_id: result.id,
      details: `Создана запчасть "${result.name}" (ID: ${result.id})`,
    }).catch(console.warn);

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
      console.error('partsApi.update: Failed to update part, status:', response.status, 'error:', errorText);
      throw new Error(`Failed to update part: ${errorText}`);
    }

    // Check for error in response body even if status is ok
    const result = await response.json();
    if (result.error) {
      console.error('partsApi.update: Error in response body:', result.error);
      throw new Error(result.error);
    }

    // Логируем обновление запчасти
    logUserActivity({
      action: 'update_part',
      resource_type: 'part',
      resource_id: id,
      details: `Обновлена запчасть ID: ${id}`,
    }).catch(console.warn);
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
  uploadPhoto: async (id: number, photoFile: File): Promise<string> => {
    const formData = new FormData();
    formData.append('photo', photoFile);

    const headers = getAuthHeaders();
    // Remove Content-Type for FormData
    delete headers['Content-Type'];

    const response = await fetch(`${API_BASE_URL}/api/uploadpartphoto/${id}`, {
      method: 'POST',
      headers,
      credentials: 'include',
      body: formData,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to upload photo: ${errorText}`);
    }

    const result = await response.json();

    // Логируем загрузку фото
    logUserActivity({
      action: 'upload_photo',
      resource_type: 'photo',
      resource_id: id,
      details: `Загружено фото для запчасти ID: ${id}`,
    }).catch(console.warn);

    return result.photo;
  },
};
