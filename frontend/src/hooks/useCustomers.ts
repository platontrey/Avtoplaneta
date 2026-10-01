import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getCustomers,
  getCustomer,
  createCustomer,
  updateCustomer,
  deleteCustomer,
} from '../features/customers/api/customersApi';
import type { CreateCustomerInput, UpdateCustomerInput } from '@/lib/types';

export const useCustomers = (params?: { category?: string; q?: string; limit?: number; offset?: number }) => {
  return useQuery({
    queryKey: ['customers', params?.category || 'all', params?.q || ''],
    queryFn: () => getCustomers(params),
  });
};

export const useCustomer = (id?: number | null) => {
  return useQuery({
    queryKey: ['customer', id],
    queryFn: () => (id ? getCustomer(id) : null),
    enabled: Boolean(id && id > 0),
  });
};

export const useCreateCustomer = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateCustomerInput) => createCustomer(data),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['customers'] });
    },
  });
};

export const useUpdateCustomer = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateCustomerInput }) => updateCustomer({ id, data }),
    onSuccess: (_, variables) => {
      void queryClient.invalidateQueries({ queryKey: ['customers'] });
      void queryClient.invalidateQueries({ queryKey: ['customer', variables.id] });
    },
  });
};

export const useDeleteCustomer = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteCustomer(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['customers'] });
    },
  });
};
