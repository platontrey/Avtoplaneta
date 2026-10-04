import { getAuthHeaders } from '@/lib/csrf';
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

export const getOrders = async (): Promise<Order[]> => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders`, {
    credentials: 'include',
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch orders: ${response.status}`);
  }

  return response.json();
};

export const getCompletedOrders = async (): Promise<Order[]> => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders?state=completed`, {
    credentials: 'include',
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch completed orders: ${response.status}`);
  }

  return response.json();
};

export const updateOrderStatus = async (orderId: number, status: string) => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/${orderId}/status`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify({ status }),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to update order status: ${errorText}`);
  }

  return response.json();
};

export const updateOrderDetails = async (orderId: number, payload: UpdateOrderDetailsPayload): Promise<Order> => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/${orderId}`, {
    method: 'PATCH',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to update order details: ${errorText}`);
  }

  return response.json();
};

export interface AddOrderItemPayload {
  part_id: number;
  quantity: number;
  price?: number;
}

export const addOrderItem = async (
  orderId: number,
  payload: AddOrderItemPayload
) => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/${orderId}/items`, {
    method: 'POST',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to add item to order: ${errorText}`);
  }

  return response.json();
};

export const updateOrderItem = async (
  orderId: number,
  itemId: number,
  payload: { quantity?: number; price?: number }
) => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/${orderId}/items/${itemId}`, {
    method: 'PATCH',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to update order item: ${errorText}`);
  }

  return response.json();
};

export const deleteOrderItem = async (orderId: number, itemId: number) => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/${orderId}/items/${itemId}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
    credentials: 'include',
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to delete order item: ${errorText}`);
  }

  return response.json();
};

export const deleteOrder = async (orderId: number) => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/${orderId}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
    credentials: 'include',
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to delete order: ${errorText}`);
  }

  return response.json();
};

export const completeOrder = async (orderId: number) => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders/${orderId}/complete`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    credentials: 'include',
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to complete order: ${errorText}`);
  }

  return response.json();
};

export const createOrder = async (orderData: CreateOrderPayload): Promise<Order> => {
  const response = await fetch(`${ORDERS_API_URL}/api/v1/orders`, {
    method: 'POST',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify({
      customer_id: orderData.customer_id || 0,
      ...orderData,
    }),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to create order: ${errorText}`);
  }

  const result = await response.json();

  logUserActivity({
    action: 'create_order',
    resource_type: 'order',
    resource_id: result.id,
    details: `${orderData.quick_sale ? 'Быстрая продажа' : 'Создан заказ'} #${result.id} (${result.buyer_number})`,
  }).catch(console.warn);

  return result;
};