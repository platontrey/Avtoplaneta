/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { getAuthHeaders } from '@/lib/csrf';
import { safeFetchJson } from '@/lib/api';
import type { Order } from '@/lib/types';
import { logUserActivity } from '../../admin/api/adminApi';

const ORDERS_API_URL = import.meta.env.VITE_API_BASE_URL || '';

export interface CreateOrderPayload {
  customer_id?: number;
  order_number?: string;
  source?: string;
  part: string;
  part_id?: number;
  buyer_number: string;
  payment_status?: string;
  warehouse_status?: string;
  delivery_method?: string;
  transport_company?: string;
  tracking_number?: string;
  notes?: string;
  discount?: number;
  quick_sale?: boolean;
  items: {
    part_id: number;
    quantity: number;
    price?: number;
  }[];
}

export interface UpdateOrderDetailsPayload {
  buyer_number?: string;
  order_number?: string;
  source?: string;
  status?: string;
  payment_status?: string;
  warehouse_status?: string;
  delivery_method?: string;
  transport_company?: string;
  tracking_number?: string;
  notes?: string;
  discount?: number;
}

export interface AddOrderItemPayload {
  part_id: number;
  quantity: number;
  price?: number;
}

export const getOrders = async (): Promise<Order[]> => {
  return safeFetchJson<Order[]>(`${ORDERS_API_URL}/api/v1/orders`, {
    credentials: 'include',
    headers: getAuthHeaders(),
  });
};

export const getCompletedOrders = async (): Promise<Order[]> => {
  return safeFetchJson<Order[]>(`${ORDERS_API_URL}/api/v1/orders?state=completed`, {
    credentials: 'include',
    headers: getAuthHeaders(),
  });
};

export const updateOrderStatus = async (orderId: number, status: string) => {
  return safeFetchJson(`${ORDERS_API_URL}/api/v1/orders/${orderId}/status`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify({ status }),
  });
};

export const updateOrderDetails = async (orderId: number, payload: UpdateOrderDetailsPayload): Promise<Order> => {
  return safeFetchJson<Order>(`${ORDERS_API_URL}/api/v1/orders/${orderId}`, {
    method: 'PATCH',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify(payload),
  });
};

export const addOrderItem = async (
  orderId: number,
  payload: AddOrderItemPayload
) => {
  return safeFetchJson(`${ORDERS_API_URL}/api/v1/orders/${orderId}/items`, {
    method: 'POST',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify(payload),
  });
};

export const updateOrderItem = async (
  orderId: number,
  itemId: number,
  payload: { quantity?: number; price?: number }
) => {
  return safeFetchJson(`${ORDERS_API_URL}/api/v1/orders/${orderId}/items/${itemId}`, {
    method: 'PATCH',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify(payload),
  });
};

export const deleteOrderItem = async (orderId: number, itemId: number) => {
  return safeFetchJson(`${ORDERS_API_URL}/api/v1/orders/${orderId}/items/${itemId}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
    credentials: 'include',
  });
};

export const deleteOrder = async (orderId: number) => {
  return safeFetchJson(`${ORDERS_API_URL}/api/v1/orders/${orderId}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
    credentials: 'include',
  });
};

export const completeOrder = async (orderId: number) => {
  return safeFetchJson(`${ORDERS_API_URL}/api/v1/orders/${orderId}/complete`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    credentials: 'include',
  });
};

export const createOrder = async (orderData: CreateOrderPayload): Promise<Order> => {
  const result = await safeFetchJson<Order>(`${ORDERS_API_URL}/api/v1/orders`, {
    method: 'POST',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify({
      customer_id: orderData.customer_id || 0,
      ...orderData,
    }),
  });

  logUserActivity({
    action: 'create_order',
    resource_type: 'order',
    resource_id: result.id,
    details: `${orderData.quick_sale ? 'Быстрая продажа' : 'Создан заказ'} #${result.id} (${result.buyer_number})`,
  }).catch(console.warn);

  return result;
};