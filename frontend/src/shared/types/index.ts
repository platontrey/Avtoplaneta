/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

export interface ApiResponse<T> {
  data: T;
  message?: string;
  success: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export interface SelectOption {
  value: string;
  label: string;
}

export interface FormFieldProps {
  id: string;
  label: string;
  error?: string;
  required?: boolean;
}