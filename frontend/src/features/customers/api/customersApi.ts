import { getAuthHeaders } from '@/lib/csrf';
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

  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/customers?${query.toString()}`, {
    credentials: 'include',
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch customers: ${response.status}`);
  }

  return response.json();
};

export const getCustomer = async (id: number): Promise<CustomerDetails> => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/customers/${id}`, {
    credentials: 'include',
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch customer #${id}: ${response.status}`);
  }

  return response.json();
};

export const createCustomer = async (data: CreateCustomerInput): Promise<Customer> => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/customers`, {
    method: 'POST',
    credentials: 'include',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to create customer: ${errorText}`);
  }

  return response.json();
};

export const updateCustomer = async ({ id, data }: { id: number; data: UpdateCustomerInput }): Promise<Customer> => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/customers/${id}`, {
    method: 'PATCH',
    credentials: 'include',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to update customer: ${errorText}`);
  }

  return response.json();
};

export const deleteCustomer = async (id: number): Promise<{ message: string }> => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/customers/${id}`, {
    method: 'DELETE',
    credentials: 'include',
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to delete customer: ${errorText}`);
  }

  return response.json();
};
