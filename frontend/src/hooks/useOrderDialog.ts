/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState } from 'react';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { createOrder, getOrders } from '../features/orders/api/ordersApi';
import { getAuthHeaders } from '@/lib/csrf';
import { ORDERS_API_URL } from '@/lib/api';
import type { Order } from '@/lib/types';

export interface OrderForm {
  customer_id: string;
  order_number: string;
  part: string;
  buyer_number: string;
  quantity: number;
}

export type OrderChoice = 'new' | 'existing' | null;

export const useOrderDialog = (partName: string) => {
  const [isOrderDialogOpen, setIsOrderDialogOpen] = useState(false);
  const [orderChoice, setOrderChoice] = useState<OrderChoice>(null);
  const [selectedOrderId, setSelectedOrderId] = useState<number | null>(null);
  const [orderForm, setOrderForm] = useState<OrderForm>({
    customer_id: '',
    order_number: '',
    part: partName,
    buyer_number: '',
    quantity: 1
  });

  const queryClient = useQueryClient();

  // Query for existing orders
  const { data: existingOrders } = useQuery<Order[]>({
    queryKey: ['orders'],
    queryFn: getOrders,
    enabled: isOrderDialogOpen, // Always fetch orders when dialog is open
  });

  const createOrderMutation = useMutation({
    mutationFn: createOrder,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['orders'] });
      void queryClient.invalidateQueries({ queryKey: ['parts'] });
      closeDialog();
    },
    onError: (error) => {
      console.error('Failed to create order:', error);
    },
  });

  const addToExistingOrderMutation = useMutation({
    mutationFn: async ({ orderId, partId, quantity }: { orderId: number; partId: number; quantity: number }) => {
      const response = await fetch(`${ORDERS_API_URL}/orders/${orderId}/items`, {
        method: 'POST',
        headers: getAuthHeaders(),
        credentials: 'include',
        body: JSON.stringify({ part_id: partId, quantity }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to add item to order: ${errorText}`);
      }

      return response.json();
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['orders'] });
      void queryClient.invalidateQueries({ queryKey: ['parts'] });
      closeDialog();
    },
    onError: (error) => {
      console.error('Failed to add item to existing order:', error);
    },
  });

  const openDialog = () => setIsOrderDialogOpen(true);

  const closeDialog = () => {
    setIsOrderDialogOpen(false);
    setOrderChoice(null);
    setSelectedOrderId(null);
    setOrderForm({
      customer_id: '',
      order_number: '',
      part: partName,
      buyer_number: '',
      quantity: 1
    });
  };

  const selectChoice = (choice: OrderChoice) => setOrderChoice(choice);

  const updateForm = (field: keyof OrderForm, value: string | number) => {
    setOrderForm(prev => ({ ...prev, [field]: value }));
  };

  const submitNewOrder = (partId: number) => {
    if (!orderForm.customer_id || !orderForm.buyer_number) return;

    createOrderMutation.mutate({
      customer_id: parseInt(orderForm.customer_id),
      order_number: orderForm.order_number,
      part: orderForm.part,
      buyer_number: orderForm.buyer_number,
      items: [{ part_id: partId, quantity: orderForm.quantity }],
    });
  };

  const submitAddToExisting = (partId: number) => {
    if (!selectedOrderId) return;
    addToExistingOrderMutation.mutate({
      orderId: selectedOrderId,
      partId,
      quantity: orderForm.quantity,
    });
  };

  return {
    // State
    isOrderDialogOpen,
    orderChoice,
    selectedOrderId,
    orderForm,
    existingOrders,

    // Mutations
    createOrderMutation,
    addToExistingOrderMutation,

    // Actions
    openDialog,
    closeDialog,
    selectChoice,
    updateForm,
    submitNewOrder,
    submitAddToExisting,
    setSelectedOrderId,
  };
};