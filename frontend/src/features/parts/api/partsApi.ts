/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { Part } from '../types';
import { getAuthHeaders } from '@/lib/csrf';
import { logUserActivity } from '../../admin/api/adminApi';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

export const partsApi = {
  // Get all parts with optional filters
  getAll: async (filters?: {
    search?: string;
    category?: string;
    brand?: string;
    model?: string;
    location?: string;
    address?: string;
    salesman?: string;
    status?: string;
    hasPhoto?: string;
    number?: string;
    oem_code?: string;
    vin?: string;
    body_brand?: string;
    engine_brand?: string;
    car_release_date?: string;
    car_release_period?: string;
    transmission?: string;
    drive?: string;
    condition?: string;
    manufacturer?: string;
    defect?: string;
    color?: string;
    min_price?: string;
    max_price?: string;
    min_quantity?: string;
    max_quantity?: string;
    front_rear?: string;
    left_right?: string;
    top_bottom?: string;
    manufacturer_code?: string;
    supplier_code?: string;
    transmission_model?: string;
    wear_percentage?: string;
    season?: string;
    diameter?: string;
    width?: string;
    profile?: string;
    tire_quantity?: string;
    drilling?: string;
    offset?: string;
    center_hole_diameter?: string;
    tire_model?: string;
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
    if (filters?.address) params.append('address', filters.address);
    if (filters?.salesman) params.append('salesman', filters.salesman);
    if (filters?.status) params.append('status', filters.status);
    if (filters?.hasPhoto) params.append('hasPhoto', filters.hasPhoto);
    if (filters?.number) params.append('number', filters.number);
    if (filters?.oem_code) params.append('oem_code', filters.oem_code);
    if (filters?.vin) params.append('vin', filters.vin);
    if (filters?.body_brand) params.append('body_brand', filters.body_brand);
    if (filters?.engine_brand) params.append('engine_brand', filters.engine_brand);
    if (filters?.car_release_date) params.append('car_release_date', filters.car_release_date);
    if (filters?.car_release_period) params.append('car_release_period', filters.car_release_period);
    if (filters?.transmission) params.append('transmission', filters.transmission);
    if (filters?.drive) params.append('drive', filters.drive);
    if (filters?.condition) params.append('condition', filters.condition);
    if (filters?.manufacturer) params.append('manufacturer', filters.manufacturer);
    if (filters?.defect) params.append('defect', filters.defect);
    if (filters?.color) params.append('color', filters.color);
    if (filters?.min_price) params.append('min_price', filters.min_price);
    if (filters?.max_price) params.append('max_price', filters.max_price);
    if (filters?.min_quantity) params.append('min_quantity', filters.min_quantity);
    if (filters?.max_quantity) params.append('max_quantity', filters.max_quantity);
    if (filters?.front_rear) params.append('front_rear', filters.front_rear);
    if (filters?.left_right) params.append('left_right', filters.left_right);
    if (filters?.top_bottom) params.append('top_bottom', filters.top_bottom);
    if (filters?.manufacturer_code) params.append('manufacturer_code', filters.manufacturer_code);
    if (filters?.supplier_code) params.append('supplier_code', filters.supplier_code);
    if (filters?.transmission_model) params.append('transmission_model', filters.transmission_model);
    if (filters?.wear_percentage) params.append('wear_percentage', filters.wear_percentage);
    if (filters?.season) params.append('season', filters.season);
    if (filters?.diameter) params.append('diameter', filters.diameter);
    if (filters?.width) params.append('width', filters.width);
    if (filters?.profile) params.append('profile', filters.profile);
    if (filters?.tire_quantity) params.append('tire_quantity', filters.tire_quantity);
    if (filters?.drilling) params.append('drilling', filters.drilling);
    if (filters?.offset) params.append('offset', filters.offset);
    if (filters?.center_hole_diameter) params.append('center_hole_diameter', filters.center_hole_diameter);
    if (filters?.tire_model) params.append('tire_model', filters.tire_model);
    if (filters?.limit && filters.limit > 0) {
      params.append('limit', filters.limit.toString());
    }
    if (filters?.page && filters.page > 0) {
      params.append('page', filters.page.toString());
    }

    const url = `${API_BASE_URL}/api/v1/inventory${params.toString() ? '?' + params.toString() : ''}`;
    console.log('partsApi.getAll: Fetching URL:', url);

    const response = await fetch(url, {
      credentials: 'include',
    });

    console.log('partsApi.getAll: Response status:', response.status);
    if (!response.ok) {
      throw new Error(`Failed to fetch parts: ${response.status}`);
    }

    const data = await response.json();
    // grpc-gateway returns { "parts": [...] }
    return data.parts || [];
  },

  // Add new part
  create: async (partData: Omit<Part, 'id'>): Promise<Part> => {
    console.log('partsApi.create: Attempting to create part');
    const headers = getAuthHeaders();
    console.log('partsApi.create: Headers being sent:', headers);
    const response = await fetch(`${API_BASE_URL}/api/v1/parts`, {
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
  update: async (id: number, partData: Partial<Part>): Promise<Part> => {
    const response = await fetch(`${API_BASE_URL}/api/v1/parts/${id}`, {
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

    return result;
  },

  // Delete part
  delete: async (id: number): Promise<void> => {
    console.log(`partsApi.delete: Deleting part with id ${id}`);
    const response = await fetch(`${API_BASE_URL}/api/v1/parts/${id}`, {
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

  // Upload part photo (добавляет в массив фото)
  uploadPhoto: async (id: number, photoFile: File): Promise<string[]> => {
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

    // Возвращаем массив всех фото запчасти (после добавления)
    return result.photos || [result.photo];
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
};
