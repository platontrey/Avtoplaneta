import { useQuery } from '@tanstack/react-query';
import { getOrders, getCompletedOrders } from '../features/orders/api/ordersApi';

export const useOrders = () => {
  return useQuery({
    queryKey: ['orders'],
    queryFn: getOrders,
  });
};

export const useCompletedOrders = (enabled = true) => {
  return useQuery({
    queryKey: ['orders', 'completed'],
    queryFn: getCompletedOrders,
    enabled,
  });
};