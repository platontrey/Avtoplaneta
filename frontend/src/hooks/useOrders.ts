import { useQuery } from '@tanstack/react-query';
import { getOrders, getCompletedOrders } from '../features/orders/api/ordersApi';
import { CACHE_TIERS } from '@/lib/queryClient';

export const useOrders = () => {
  return useQuery({
    queryKey: ['orders'],
    queryFn: getOrders,
    ...CACHE_TIERS.REALTIME,
  });
};

export const useCompletedOrders = (enabled = true) => {
  return useQuery({
    queryKey: ['orders', 'completed'],
    queryFn: getCompletedOrders,
    enabled,
    ...CACHE_TIERS.REALTIME,
  });
};