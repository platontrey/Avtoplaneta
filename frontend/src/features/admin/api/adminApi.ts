import { getAuthHeaders } from '@/lib/csrf';
import type { UserActivityLog, UserActivityAction, UserActivityResourceType } from '@/lib/types';

// Admin service работает через gateway
const ADMIN_API_URL = import.meta.env.VITE_API_BASE_URL || '';

export const getUserActivityLogs = async (params?: {
  user_id?: number;
  action?: UserActivityAction;
  resource_type?: UserActivityResourceType;
  limit?: number;
  offset?: number;
  start_date?: string;
  end_date?: string;
  // Скрывает навигационный шум, оставляя только полезные действия (мутации + вход/выход).
  useful_only?: boolean;
}): Promise<UserActivityLog[]> => {
  const url = new URL(`${ADMIN_API_URL}/admin/user-activity-logs`);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        url.searchParams.append(key, value.toString());
      }
    });
  }

  console.log('Fetching user activity logs from:', url.toString());

  const response = await fetch(url.toString(), {
    credentials: 'include',
    headers: getAuthHeaders(),
  });

  console.log('Response status:', response.status);

  if (!response.ok) {
    console.error('Failed to fetch user activity logs:', response.status);
    throw new Error(`Failed to fetch user activity logs: ${response.status}`);
  }

  const data = await response.json();
  console.log('Received activity logs data:', data);
  return data.logs || [];
};

export const logUserActivity = async (activity: {
  action: UserActivityAction;
  resource_type: UserActivityResourceType;
  resource_id?: number;
  details?: string;
}): Promise<void> => {
  const response = await fetch(`${ADMIN_API_URL}/admin/user-activity-logs`, {
    method: 'POST',
    headers: getAuthHeaders(),
    credentials: 'include',
    body: JSON.stringify(activity),
  });

  if (!response.ok) {
    console.warn('Failed to log user activity:', response.status);
    // Don't throw error to avoid breaking user flow
  }
};