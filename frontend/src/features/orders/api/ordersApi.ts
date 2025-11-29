
import { getAuthHeaders } from '@/lib/csrf';
import type { Order } from '@/lib/types';
import { logUserActivity } from '../../admin/api/adminApi';

// Orders service работает на отдельном порту
const ORDERS_API_URL = 'http://localhost:8082';

export const getOrders = async (): Promise<Order[]> => {
  console.log('ordersApi.getOrders: Fetching from', `${ORDERS_API_URL}/orders`);
  const response = await fetch(`${ORDERS_API_URL}/orders`, {
    credentials: 'include',
    headers: getAuthHeaders(),
  });

  console.log('ordersApi.getOrders: Response status', response.status);
  if (!response.ok) {
    throw new Error(`Failed to fetch orders: ${response.status}`);
  }

  const data = await response.json();
  console.log('ordersApi.getOrders: Received data', data);
  return data;
};

export const updateOrderStatus = async (orderId: number, status: string) => {
  const response = await fetch(`${ORDERS_API_URL}/admin/orders/${orderId}/status`, {
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

export const deleteOrder = async (orderId: number) => {
  const response = await fetch(`${ORDERS_API_URL}/admin/orders/${orderId}`, {
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
  const response = await fetch(`${ORDERS_API_URL}/admin/orders/${orderId}/complete`, {
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

export const createOrder = async (orderData: { customer_id: number; order_number: string; part: string; buyer_number: string; items: { part_id: number; quantity: number }[] }) => {
  console.log('ordersApi.createOrder: Creating order with data:', orderData);

  // Create the order directly - the backend will handle quantity updates
  const response = await fetch(`${ORDERS_API_URL}/orders`, {
    method: 'POST',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify(orderData),
  });

  if (!response.ok) {
    const errorText = await response.text();
    console.error('ordersApi.createOrder: Failed to create order:', errorText);
    throw new Error(`Failed to create order: ${errorText}`);
  }

  const result = await response.json();
  console.log('ordersApi.createOrder: Order created successfully:', result);

  // Логируем создание заказа
  logUserActivity({
    action: 'create_order',
    resource_type: 'order',
    resource_id: result.id,
    details: `Создан заказ ${result.id} для клиента ID: ${result.customer_id}`,
  }).catch(console.warn);

  return result;
};