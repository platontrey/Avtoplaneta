/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { Part } from '../types';
import type { PartFilters } from '@/hooks/useParts';
import { getAuthHeaders } from '@/lib/csrf';
import { safeFetchJson } from '@/lib/api';
import { logUserActivity } from '../../admin/api/adminApi';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

export const partsApi = {
  // Get all parts with optional filters
  getAll: async (filters?: PartFilters): Promise<Part[]> => {
    const params = new URLSearchParams();
    if (filters) {
      for (const [key, value] of Object.entries(filters)) {
        if (value === undefined || value === null || value === '' || (key === 'hasPhoto' && value === 'all')) {
          continue;
        }
        if ((key === 'limit' || key === 'page') && typeof value === 'number' && value <= 0) {
          continue;
        }
        params.append(key, String(value));
      }
    }

    const queryString = params.toString();
    const url = `${API_BASE_URL}/api/v1/inventory${queryString ? '?' + queryString : ''}`;

    const data = await safeFetchJson<{ parts?: Part[] } | Part[]>(url, {
      credentials: 'include',
    });

    return Array.isArray(data) ? data : data.parts || [];
  },

  // Get single part by ID
  getById: async (id: number): Promise<Part> => {
    return safeFetchJson<Part>(`${API_BASE_URL}/api/v1/parts/item/${id}`, {
      credentials: 'include',
      headers: getAuthHeaders(),
    });
  },

  // Add new part
  create: async (partData: Omit<Part, 'id'>): Promise<Part> => {
    const result = await safeFetchJson<Part>(`${API_BASE_URL}/api/v1/parts`, {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(partData),
    });

    logUserActivity({
      action: 'create_part',
      resource_type: 'part',
      resource_id: result.id,
      details: `Создана запчасть "${result.name}" (ID: ${result.id})`,
    }).catch(console.warn);

    return result;
  },

  // Update part
  update: async (id: number, partData: Partial<Part>): Promise<Part> => {
    const result = await safeFetchJson<Part & { error?: string }>(`${API_BASE_URL}/api/v1/parts/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(partData),
    });

    if (result.error) {
      throw new Error(result.error);
    }

    logUserActivity({
      action: 'update_part',
      resource_type: 'part',
      resource_id: id,
      details: `Обновлена запчасть ID: ${id}`,
    }).catch(console.warn);

    return result;
  },

  // Delete part
  delete: async (id: number): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/v1/parts/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
      credentials: 'include',
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to delete part: ${errorText}`);
    }
  },

  // Upload part photo (добавляет в массив фото)
  uploadPhoto: async (id: number, photoFile: File): Promise<string[]> => {
    const formData = new FormData();
    formData.append('photo', photoFile);

    const headers = getAuthHeaders();
    delete headers['Content-Type'];

    const result = await safeFetchJson<{ photos?: string[]; photo?: string }>(
      `${API_BASE_URL}/api/v1/uploadpartphoto/${id}`,
      {
        method: 'POST',
        headers,
        credentials: 'include',
        body: formData,
      }
    );

    logUserActivity({
      action: 'upload_photo',
      resource_type: 'photo',
      resource_id: id,
      details: `Загружено фото для запчасти ID: ${id}`,
    }).catch(console.warn);

    return result.photos || (result.photo ? [result.photo] : []);
  },

  // Upload multiple photos
  uploadPhotos: async (id: number, photoFiles: File[]): Promise<string[]> => {
    const allPhotos: string[] = [];
    for (const file of photoFiles) {
      const photos = await partsApi.uploadPhoto(id, file);
      allPhotos.push(...photos);
    }
    return allPhotos;
  },

  // Duplicate part (creates an exact copy without photos)
  duplicate: async (id: number): Promise<Part> => {
    const result = await safeFetchJson<Part>(`${API_BASE_URL}/api/v1/parts/${id}/duplicate`, {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include',
    });

    logUserActivity({
      action: 'duplicate_part',
      resource_type: 'part',
      resource_id: result.id,
      details: `Дублирована запчасть ID: ${id} -> новая ID: ${result.id} "${result.name}"`,
    }).catch(console.warn);

    return result;
  },
};
