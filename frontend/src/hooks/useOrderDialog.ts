/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useState, useEffect } from 'react';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { createOrder, getOrders, addOrderItem } from '../features/orders/api/ordersApi';
import type { Order } from '@/lib/types';

export interface OrderForm {
  customer_id?: number;
  order_number: string;
  source: string;
  part: string;
  buyer_number: string;
  quantity: number;
  price: number;
  discount?: number;
  payment_status: string;
  warehouse_status: string;
  delivery_method: string;
  transport_company: string;
  notes: string;
}

export type OrderChoice = 'quick' | 'new' | 'existing' | null;

export const useOrderDialog = (partName: string, defaultPrice: number = 0) => {
  const [isOrderDialogOpen, setIsOrderDialogOpen] = useState(false);
  const [orderChoice, setOrderChoice] = useState<OrderChoice>('new');
  const [selectedOrderId, setSelectedOrderId] = useState<number | null>(null);
  const [orderForm, setOrderForm] = useState<OrderForm>({
    order_number: '',
    source: 'drom',
    part: partName,
    buyer_number: '',
    quantity: 1,
    price: defaultPrice,
    payment_status: 'unpaid',
    warehouse_status: 'inspecting',
    delivery_method: 'tk',
    transport_company: '',
    notes: '',
  });

  useEffect(() => {
    setOrderForm((prev) => ({
      ...prev,
      part: partName,
      price: defaultPrice,
    }));
  }, [partName, defaultPrice]);

  const queryClient = useQueryClient();

  const { data: existingOrders } = useQuery<Order[]>({
    queryKey: ['orders'],
    queryFn: getOrders,
    enabled: isOrderDialogOpen,
  });

  const createOrderMutation = useMutation({
    mutationFn: createOrder,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['orders'] });
      void queryClient.invalidateQueries({ queryKey: ['orders', 'completed'] });
      void queryClient.invalidateQueries({ queryKey: ['parts'] });
      void queryClient.invalidateQueries({ queryKey: ['inventory'] });
      void queryClient.invalidateQueries({ queryKey: ['statistics'] });
      closeDialog();
    },
    onError: (error) => {
      console.error('Failed to create order:', error);
    },
  });

  const addToExistingOrderMutation = useMutation({
    mutationFn: async ({
      orderId,
      partId,
      quantity,
      price,
    }: {
      orderId: number;
      partId: number;
      quantity: number;
      price?: number;
    }) => {
      return addOrderItem(orderId, { part_id: partId, quantity, price });
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
    setOrderChoice('new');
    setSelectedOrderId(null);
    setOrderForm({
      order_number: '',
      source: 'drom',
      part: partName,
      buyer_number: '',
      quantity: 1,
      price: defaultPrice,
      payment_status: 'unpaid',
      warehouse_status: 'inspecting',
      delivery_method: 'tk',
      transport_company: '',
      notes: '',
    });
  };

  const selectChoice = (choice: OrderChoice) => setOrderChoice(choice);

  const updateForm = (field: keyof OrderForm, value: string | number) => {
    setOrderForm((prev) => ({ ...prev, [field]: value }));
  };

  const submitNewOrder = (partId: number, onDone?: () => void) => {
    if (!orderForm.buyer_number.trim()) return;

    createOrderMutation.mutate(
      {
        customer_id: orderForm.customer_id || 0,
        order_number: orderForm.order_number.trim(),
        source: orderForm.source,
        part: orderForm.part,
        part_id: partId,
        buyer_number: orderForm.buyer_number.trim(),
        payment_status: orderForm.payment_status,
        warehouse_status: orderForm.warehouse_status,
        delivery_method: orderForm.delivery_method,
        transport_company: orderForm.transport_company.trim(),
        notes: orderForm.notes.trim(),
        discount: Number(orderForm.discount) || 0,
        quick_sale: false,
        items: [
          {
            part_id: partId,
            quantity: Math.max(1, Number(orderForm.quantity) || 1),
            price: Number(orderForm.price) >= 0 ? Number(orderForm.price) : defaultPrice,
          },
        ],
      },
      {
        onSuccess: () => {
          onDone?.();
        },
      }
    );
  };

  const submitQuickSale = (partId: number, onDone?: () => void) => {
    createOrderMutation.mutate(
      {
        customer_id: 0,
        order_number: orderForm.order_number.trim(),
        source: 'pickup',
        part: orderForm.part,
        part_id: partId,
        buyer_number: orderForm.buyer_number.trim() || 'Самовывоз',
        payment_status: 'paid',
        warehouse_status: 'ready',
        delivery_method: 'pickup',
        notes: orderForm.notes.trim(),
        quick_sale: true,
        items: [
          {
            part_id: partId,
            quantity: Math.max(1, Number(orderForm.quantity) || 1),
            price: Number(orderForm.price) >= 0 ? Number(orderForm.price) : defaultPrice,
          },
        ],
      },
      {
        onSuccess: () => {
          onDone?.();
        },
      }
    );
  };

  const submitAddToExisting = (partId: number, onDone?: () => void) => {
    if (!selectedOrderId) return;
    addToExistingOrderMutation.mutate(
      {
        orderId: selectedOrderId,
        partId,
        quantity: Math.max(1, Number(orderForm.quantity) || 1),
        price: Number(orderForm.price) >= 0 ? Number(orderForm.price) : defaultPrice,
      },
      {
        onSuccess: () => {
          onDone?.();
        },
      }
    );
  };

  return {
    isOrderDialogOpen,
    orderChoice,
    selectedOrderId,
    orderForm,
    existingOrders,
    createOrderMutation,
    addToExistingOrderMutation,
    openDialog,
    closeDialog,
    selectChoice,
    updateForm,
    submitNewOrder,
    submitQuickSale,
    submitAddToExisting,
    setSelectedOrderId,
  };
};