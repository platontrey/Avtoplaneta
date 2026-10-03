/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useOrders, useCompletedOrders } from '@/hooks/useOrders';
import {
  updateOrderDetails,
  updateOrderItem,
  deleteOrderItem,
  deleteOrder,
  completeOrder,
} from '../api/ordersApi';
import type { Order, OrderItem } from '@/lib/types';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Plus, History, ClipboardList, Users } from 'lucide-react';
import { CustomersTab } from '@/features/customers/components/CustomersTab';
import { calcOrderTotal } from '../utils/orderHelpers';
import { OrdersFiltersBar } from './OrdersFiltersBar';
import { OrderMobileCard } from './OrderMobileCard';
import { OrdersTable } from './OrdersTable';
import { EditOrderDialog } from './EditOrderDialog';
import { CompleteOrderDialog, DeleteOrderDialog } from './OrderActionDialogs';

export const Orders: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState<'active' | 'completed' | 'customers'>('active');
  const [searchQuery, setSearchQuery] = useState('');
  const [paymentFilter, setPaymentFilter] = useState<string>('all');
  const [warehouseFilter, setWarehouseFilter] = useState<string>('all');
  const [locationFilter, setLocationFilter] = useState<string>('all');

  const { data: activeOrders, isLoading: activeLoading, error: activeError, refetch: refetchActive } = useOrders();
  const { data: completedOrders, isLoading: completedLoading } = useCompletedOrders(activeTab === 'completed');

  const [orderToDelete, setOrderToDelete] = useState<Order | null>(null);
  const [orderToComplete, setOrderToComplete] = useState<Order | null>(null);
  const [editingOrder, setEditingOrder] = useState<Order | null>(null);

  const invalidateAll = () => {
    void queryClient.invalidateQueries({ queryKey: ['orders'] });
    void queryClient.invalidateQueries({ queryKey: ['orders', 'completed'] });
  };

  const quickUpdateMutation = useMutation({
    mutationFn: ({ orderId, payload }: { orderId: number; payload: Parameters<typeof updateOrderDetails>[1] }) =>
      updateOrderDetails(orderId, payload),
    onMutate: async ({ orderId, payload }) => {
      await queryClient.cancelQueries({ queryKey: ['orders'] });
      await queryClient.cancelQueries({ queryKey: ['orders', 'completed'] });

      const previousActive = queryClient.getQueryData<Order[]>(['orders']);
      const previousCompleted = queryClient.getQueryData<Order[]>(['orders', 'completed']);

      if (previousActive) {
        queryClient.setQueryData<Order[]>(['orders'], (old) =>
          old ? old.map((ord) => (ord.id === orderId ? { ...ord, ...payload } : ord)) : []
        );
      }

      if (previousCompleted) {
        queryClient.setQueryData<Order[]>(['orders', 'completed'], (old) =>
          old ? old.map((ord) => (ord.id === orderId ? { ...ord, ...payload } : ord)) : []
        );
      }

      return { previousActive, previousCompleted };
    },
    onError: (_err, _variables, context) => {
      if (context?.previousActive) {
        queryClient.setQueryData(['orders'], context.previousActive);
      }
      if (context?.previousCompleted) {
        queryClient.setQueryData(['orders', 'completed'], context.previousCompleted);
      }
    },
    onSettled: () => {
      invalidateAll();
    },
  });

  const deleteOrderMutation = useMutation({
    mutationFn: deleteOrder,
    onSuccess: () => {
      invalidateAll();
      setOrderToDelete(null);
    },
  });

  const completeOrderMutation = useMutation({
    mutationFn: completeOrder,
    onSuccess: () => {
      invalidateAll();
      void queryClient.invalidateQueries({ queryKey: ['parts'] });
      void queryClient.invalidateQueries({ queryKey: ['inventory'] });
      void queryClient.invalidateQueries({ queryKey: ['statistics'] });
      setOrderToComplete(null);
    },
  });

  const saveOrderDetailsMutation = useMutation({
    mutationFn: async ({
      orderId,
      details,
    }: {
      orderId: number;
      details: {
        buyer_number: string;
        order_number: string;
        source: string;
        payment_status: string;
        warehouse_status: string;
        delivery_method: string;
        transport_company: string;
        tracking_number: string;
        notes: string;
        discount: number;
        items: OrderItem[];
      };
    }) => {
      await updateOrderDetails(orderId, {
        buyer_number: details.buyer_number,
        order_number: details.order_number,
        source: details.source,
        payment_status: details.payment_status,
        warehouse_status: details.warehouse_status,
        delivery_method: details.delivery_method,
        transport_company: details.transport_company,
        tracking_number: details.tracking_number,
        notes: details.notes,
        discount: details.discount,
      });

      for (const item of details.items) {
        await updateOrderItem(orderId, item.id, {
          quantity: Math.max(1, Number(item.quantity) || 1),
          price: Math.max(0, Number(item.price) || 0),
        });
      }
    },
    onSuccess: () => {
      invalidateAll();
      setEditingOrder(null);
    },
  });

  const removeOrderItemMutation = useMutation({
    mutationFn: ({ orderId, itemId }: { orderId: number; itemId: number }) =>
      deleteOrderItem(orderId, itemId),
    onSuccess: () => {
      invalidateAll();
    },
  });

  const handleQuickUpdate = (orderId: number, payload: Parameters<typeof updateOrderDetails>[1]) => {
    quickUpdateMutation.mutate({ orderId, payload });
  };

  const uniqueLocations = useMemo(() => {
    const locs = new Set<string>();
    for (const o of activeOrders || []) {
      if (o.location && o.location !== 'Неизвестно' && o.location !== 'Нет деталей') {
        locs.add(o.location);
      }
    }
    return Array.from(locs);
  }, [activeOrders]);

  const filteredOrders = useMemo(() => {
    const list = activeTab === 'active' ? activeOrders || [] : completedOrders || [];
    const q = searchQuery.trim().toLowerCase();

    return list.filter((order) => {
      if (activeTab === 'active') {
        if (paymentFilter !== 'all' && (order.payment_status || 'unpaid') !== paymentFilter) {
          return false;
        }
        if (warehouseFilter !== 'all' && (order.warehouse_status || 'inspecting') !== warehouseFilter) {
          return false;
        }
        if (locationFilter !== 'all' && order.location !== locationFilter) {
          return false;
        }
      }

      if (!q) return true;

      const inItems = (order.items || []).some(
        (it) =>
          (it.part_name || '').toLowerCase().includes(q) ||
          (it.part_name_snapshot || '').toLowerCase().includes(q) ||
          String(it.part_id).includes(q)
      );

      return (
        String(order.id).includes(q) ||
        (order.buyer_number || '').toLowerCase().includes(q) ||
        (order.order_number || '').toLowerCase().includes(q) ||
        (order.part || '').toLowerCase().includes(q) ||
        (order.seller || '').toLowerCase().includes(q) ||
        (order.tracking_number || '').toLowerCase().includes(q) ||
        (order.transport_company || '').toLowerCase().includes(q) ||
        (order.notes || '').toLowerCase().includes(q) ||
        inItems
      );
    });
  }, [activeOrders, completedOrders, activeTab, searchQuery, paymentFilter, warehouseFilter, locationFilter]);

  const totalListSum = useMemo(
    () => filteredOrders.reduce((sum, o) => sum + calcOrderTotal(o), 0),
    [filteredOrders]
  );

  if (activeLoading && activeTab === 'active') {
    return (
      <div className="container mx-auto p-4 space-y-4">
        <Skeleton className="h-9 w-56" />
        <Skeleton className="h-20 w-full rounded-lg" />
        <Skeleton className="h-96 w-full rounded-lg" />
      </div>
    );
  }

  if (activeError) {
    return (
      <div className="container mx-auto p-4 space-y-4">
        <div className="p-4 bg-destructive/10 text-destructive border border-destructive/20 rounded-lg flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <p className="font-semibold text-base">Ошибка загрузки заказов</p>
            <p className="text-xs text-muted-foreground mt-1">
              {activeError instanceof Error ? activeError.message : 'Не удалось получить список заказов с сервера'}
            </p>
          </div>
          <Button variant="outline" size="sm" onClick={() => refetchActive()}>
            Повторить попытку
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-4 space-y-5">
      {/* Шапка и переключение вкладок */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">
            {activeTab === 'customers' ? 'База клиентов' : 'Управление заказами'}
          </h1>
          <p className="text-xs sm:text-sm text-muted-foreground">
            {activeTab === 'customers'
              ? 'Единая база покупателей, история заказов, реквизиты для ТК и категории'
              : `Найдено: ${filteredOrders.length} · Сумма: ₽${totalListSum.toLocaleString('ru-RU')}`}
          </p>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          <div className="inline-flex rounded-lg border bg-muted p-1">
            <Button
              variant={activeTab === 'active' ? 'default' : 'ghost'}
              size="sm"
              className="gap-1.5 text-xs sm:text-sm"
              onClick={() => setActiveTab('active')}
            >
              <ClipboardList className="h-4 w-4" />
              <span>В работе ({activeOrders?.length || 0})</span>
            </Button>
            <Button
              variant={activeTab === 'completed' ? 'default' : 'ghost'}
              size="sm"
              className="gap-1.5 text-xs sm:text-sm"
              onClick={() => setActiveTab('completed')}
            >
              <History className="h-4 w-4" />
              <span>История продаж</span>
            </Button>
            <Button
              variant={activeTab === 'customers' ? 'default' : 'ghost'}
              size="sm"
              className="gap-1.5 text-xs sm:text-sm"
              onClick={() => setActiveTab('customers')}
            >
              <Users className="h-4 w-4" />
              <span>База клиентов</span>
            </Button>
          </div>

          <Button size="sm" onClick={() => navigate('/inventory')} className="gap-1.5">
            <Plus className="h-4 w-4" />
            <span>Подобрать в инвентаре</span>
          </Button>
        </div>
      </div>

      {activeTab === 'customers' ? (
        <CustomersTab />
      ) : (
        <Card>
          <CardHeader className="p-3 sm:p-4">
            <OrdersFiltersBar
              searchQuery={searchQuery}
              onSearchChange={setSearchQuery}
              activeTab={activeTab}
              paymentFilter={paymentFilter}
              onPaymentFilterChange={setPaymentFilter}
              warehouseFilter={warehouseFilter}
              onWarehouseFilterChange={setWarehouseFilter}
              locationFilter={locationFilter}
              onLocationFilterChange={setLocationFilter}
              uniqueLocations={uniqueLocations}
            />
          </CardHeader>

          <CardContent className="p-0 sm:p-4 sm:pt-0">
            {completedLoading && activeTab === 'completed' ? (
              <div className="p-8 text-center text-muted-foreground">Загрузка истории продаж...</div>
            ) : filteredOrders.length === 0 ? (
              <div className="p-8 text-center text-muted-foreground">
                {activeTab === 'active'
                  ? 'Активных заказов по выбранным фильтрам нет.'
                  : 'История завершённых заказов пуста.'}
              </div>
            ) : (
              <>
                {/* Мобильный вид: карточки */}
                <div className="lg:hidden divide-y divide-border">
                  {filteredOrders.map((order) => (
                    <OrderMobileCard
                      key={order.id}
                      order={order}
                      isCompleted={activeTab === 'completed'}
                      onQuickUpdate={handleQuickUpdate}
                      onOpenEdit={(o) => setEditingOrder(o)}
                      onComplete={(o) => setOrderToComplete(o)}
                      onDelete={(o) => setOrderToDelete(o)}
                    />
                  ))}
                </div>

                {/* Десктопный вид: таблица */}
                <OrdersTable
                  orders={filteredOrders}
                  activeTab={activeTab}
                  onQuickUpdate={handleQuickUpdate}
                  onOpenEdit={(o) => setEditingOrder(o)}
                  onComplete={(o) => setOrderToComplete(o)}
                  onDelete={(o) => setOrderToDelete(o)}
                />
              </>
            )}
          </CardContent>
        </Card>
      )}

      {/* Диалог редактирования заказа */}
      <EditOrderDialog
        order={editingOrder}
        isOpen={Boolean(editingOrder)}
        onClose={() => setEditingOrder(null)}
        onSave={async (orderId, details) => {
          await saveOrderDetailsMutation.mutateAsync({ orderId, details });
        }}
        onDeleteItem={async (orderId, itemId) => {
          await removeOrderItemMutation.mutateAsync({ orderId, itemId });
        }}
        isSaving={saveOrderDetailsMutation.isPending}
      />

      {/* Диалог подтверждения завершения */}
      <CompleteOrderDialog
        order={orderToComplete}
        isOpen={Boolean(orderToComplete)}
        onClose={() => setOrderToComplete(null)}
        isPending={completeOrderMutation.isPending}
        onConfirm={(order) => {
          completeOrderMutation.mutate(order.id);
        }}
      />

      {/* Диалог подтверждения удаления */}
      <DeleteOrderDialog
        order={orderToDelete}
        isOpen={Boolean(orderToDelete)}
        onClose={() => setOrderToDelete(null)}
        isPending={deleteOrderMutation.isPending}
        onConfirm={(order) => {
          deleteOrderMutation.mutate(order.id);
        }}
      />
    </div>
  );
};

export default Orders;
