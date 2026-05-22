import { useQuery } from '@tanstack/react-query';
import { getOrders } from '../features/orders/api/ordersApi';

export const useOrders = () => {
  return useQuery({
    queryKey: ['orders'],
    queryFn: getOrders,
  });
};