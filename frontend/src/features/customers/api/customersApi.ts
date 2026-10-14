/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { getAuthHeaders } from '@/lib/csrf';
import { safeFetchJson } from '@/lib/api';
import type { Customer, CustomerWithStats, CustomerDetails, CreateCustomerInput, UpdateCustomerInput } from '@/lib/types';

const ORDERS_API_URL = import.meta.env.VITE_API_BASE_URL || '';

export interface CustomersResponse {
  customers: CustomerWithStats[];
  total: number;
}

export const getCustomers = async (params?: { category?: string; q?: string; limit?: number; offset?: number }): Promise<CustomersResponse> => {
  const query = new URLSearchParams();
  if (params?.category) query.set('category', params.category);
  if (params?.q) query.set('q', params.q);
  if (params?.limit) query.set('limit', params.limit.toString());
  if (params?.offset) query.set('offset', params.offset.toString());

  const queryString = query.toString();
  return safeFetchJson<CustomersResponse>(
    `${ORDERS_API_URL}/api/v1/orders/customers${queryString ? '?' + queryString : ''}`,
    {
      credentials: 'include',
      headers: getAuthHeaders(),
    }
  );
};

export const getCustomer = async (id: number): Promise<CustomerDetails> => {
  return safeFetchJson<CustomerDetails>(`${ORDERS_API_URL}/api/v1/orders/customers/${id}`, {
    credentials: 'include',
    headers: getAuthHeaders(),
  });
};

export const createCustomer = async (data: CreateCustomerInput): Promise<Customer> => {
  return safeFetchJson<Customer>(`${ORDERS_API_URL}/api/v1/orders/customers`, {
    method: 'POST',
    credentials: 'include',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
};

export const updateCustomer = async ({ id, data }: { id: number; data: UpdateCustomerInput }): Promise<Customer> => {
  return safeFetchJson<Customer>(`${ORDERS_API_URL}/api/v1/orders/customers/${id}`, {
    method: 'PATCH',
    credentials: 'include',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
};

export const deleteCustomer = async (id: number): Promise<{ message: string }> => {
  return safeFetchJson<{ message: string }>(`${ORDERS_API_URL}/api/v1/orders/customers/${id}`, {
    method: 'DELETE',
    credentials: 'include',
    headers: getAuthHeaders(),
  });
};
