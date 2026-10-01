import { useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useOrders, useCompletedOrders } from '../hooks/useOrders';
import {
  updateOrderDetails,
  updateOrderItem,
  deleteOrderItem,
  deleteOrder,
  completeOrder,
} from '../features/orders/api/ordersApi';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { Order, OrderItem } from '../lib/types';
import { Card, CardContent, CardHeader } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Badge } from './ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table';
import { Skeleton } from './ui/skeleton';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './ui/dialog';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from './ui/alert-dialog';
import {
  ExternalLink,
  Plus,
  Search,
  Edit,
  Trash2,
  CheckCircle2,
  Truck,
  MessageSquare,
  History,
  ClipboardList,
  Users,
} from 'lucide-react';
import { CustomersTab } from '../features/customers/components/CustomersTab';

const SOURCE_LABELS: Record<string, string> = {
  drom: 'Дром',
  avito: 'Авито',
  messenger: 'Мессенджер',
  pickup: 'Самовывоз',
};

const PAYMENT_OPTIONS = [
  { value: 'unpaid', label: '💳 Не оплачен', badgeClass: 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30' },
  { value: 'prepaid', label: '💵 Предоплата', badgeClass: 'bg-blue-500/15 text-blue-700 dark:text-blue-300 border-blue-500/30' },
  { value: 'paid', label: '✅ Оплачен', badgeClass: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-500/30' },
] as const;

const WAREHOUSE_OPTIONS = [
  { value: 'inspecting', label: '📷 Проверка / Фото' },
  { value: 'transfer', label: '🔄 Перемещение' },
  { value: 'ready', label: '📦 Готов к выдаче' },
] as const;

const DELIVERY_OPTIONS = [
  { value: 'pickup', label: '🏃 Самовывоз' },
  { value: 'tk', label: '🚚 Отправка ТК' },
  { value: 'city', label: '🚕 По городу' },
] as const;

const calcOrderTotal = (order: Order): number => {
  if (typeof order.total_amount === 'number' && order.total_amount >= 0) {
    return order.total_amount;
  }
  const subtotal = (order.items || []).reduce(
    (acc, it) => acc + (Number(it.price) || 0) * (Number(it.quantity) || 1),
    0
  );
  return Math.max(0, subtotal - (Number(order.discount) || 0));
};

const renderOrderParts = (order: Order, isCompleted: boolean) => {
  if (order.items && order.items.length > 0) {
    return (
      <div className="space-y-1">
        {order.items.map((item) => {
          const targetPartId = item.part_id || order.part_id;
          const label =
            item.part_name ||
            item.part_name_snapshot ||
            (order.items.length === 1 ? order.part : '') ||
            (targetPartId ? `Запчасть #${targetPartId}` : 'Запчасть');
          const priceStr = item.price ? ` · ₽${Number(item.price).toLocaleString('ru-RU')}` : '';
          const content = `${item.quantity}× ${label}`;

          return (
            <div key={item.id || `${order.id}-${item.part_id}`} className="text-sm flex items-center gap-1.5 flex-wrap">
              {!isCompleted && targetPartId && targetPartId > 0 ? (
                <Link
                  to={`/inventory?partId=${targetPartId}`}
                  className="text-primary hover:underline inline-flex items-center gap-1 font-medium transition-colors"
                  title={`Открыть запчасть #${targetPartId} в инвентаре`}
                >
                  <span>{content}</span>
                  <ExternalLink className="h-3.5 w-3.5 shrink-0 opacity-75" />
                </Link>
              ) : (
                <span className="font-medium">{content}</span>
              )}
              {priceStr && <span className="text-xs text-muted-foreground">{priceStr}</span>}
            </div>
          );
        })}
      </div>
    );
  }

  const fallbackLabel = order.part || 'Нет деталей';
  if (!isCompleted && order.part_id && order.part_id > 0) {
    return (
      <Link
        to={`/inventory?partId=${order.part_id}`}
        className="text-sm text-primary hover:underline inline-flex items-center gap-1 font-medium transition-colors"
      >
        <span>{fallbackLabel}</span>
        <ExternalLink className="h-3.5 w-3.5 shrink-0 opacity-75" />
      </Link>
    );
  }

  return <div className="text-sm font-medium">{fallbackLabel}</div>;
};

const Orders = () => {
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

  // Состояние формы редактирования заказа
  const [editBuyer, setEditBuyer] = useState('');
  const [editOrderNumber, setEditOrderNumber] = useState('');
  const [editSource, setEditSource] = useState('drom');
  const [editPayment, setEditPayment] = useState('unpaid');
  const [editWarehouse, setEditWarehouse] = useState('inspecting');
  const [editDelivery, setEditDelivery] = useState('pickup');
  const [editTK, setEditTK] = useState('');
  const [editTracking, setEditTracking] = useState('');
  const [editNotes, setEditNotes] = useState('');
  const [editDiscount, setEditDiscount] = useState(0);
  const [editItems, setEditItems] = useState<OrderItem[]>([]);

  const invalidateAll = () => {
    void queryClient.invalidateQueries({ queryKey: ['orders'] });
    void queryClient.invalidateQueries({ queryKey: ['orders', 'completed'] });
  };

  const quickUpdateMutation = useMutation({
    mutationFn: ({ orderId, payload }: { orderId: number; payload: Parameters<typeof updateOrderDetails>[1] }) =>
      updateOrderDetails(orderId, payload),
    onSuccess: invalidateAll,
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
    mutationFn: async () => {
      if (!editingOrder) return;
      await updateOrderDetails(editingOrder.id, {
        buyer_number: editBuyer,
        order_number: editOrderNumber,
        source: editSource,
        payment_status: editPayment,
        warehouse_status: editWarehouse,
        delivery_method: editDelivery,
        transport_company: editTK,
        tracking_number: editTracking,
        notes: editNotes,
        discount: Number(editDiscount) || 0,
      });

      for (const item of editItems) {
        await updateOrderItem(editingOrder.id, item.id, {
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
    onSuccess: (_, { itemId }) => {
      setEditItems((prev) => prev.filter((it) => it.id !== itemId));
      invalidateAll();
    },
  });

  const openEditModal = (order: Order) => {
    setEditingOrder(order);
    setEditBuyer(order.buyer_number || '');
    setEditOrderNumber(order.order_number || '');
    setEditSource(order.source || 'drom');
    setEditPayment(order.payment_status || 'unpaid');
    setEditWarehouse(order.warehouse_status || 'inspecting');
    setEditDelivery(order.delivery_method || 'pickup');
    setEditTK(order.transport_company || '');
    setEditTracking(order.tracking_number || '');
    setEditNotes(order.notes || '');
    setEditDiscount(order.discount || 0);
    setEditItems(order.items ? order.items.map((it) => ({ ...it })) : []);
  };

  const currentList = activeTab === 'active' ? activeOrders || [] : completedOrders || [];

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
    const q = searchQuery.trim().toLowerCase();
    return currentList.filter((order) => {
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
  }, [currentList, activeTab, searchQuery, paymentFilter, warehouseFilter, locationFilter]);

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
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5">
            <div className="relative">
              <Search className="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Поиск: телефон, № сделки, деталь, трек..."
                className="pl-9 h-9 text-sm"
              />
            </div>

            {activeTab === 'active' && (
              <>
                <Select value={paymentFilter} onValueChange={setPaymentFilter}>
                  <SelectTrigger className="h-9 text-xs sm:text-sm">
                    <SelectValue placeholder="Статус оплаты" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">💰 Все статусы оплаты</SelectItem>
                    {PAYMENT_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select value={warehouseFilter} onValueChange={setWarehouseFilter}>
                  <SelectTrigger className="h-9 text-xs sm:text-sm">
                    <SelectValue placeholder="Этап склада" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">🏭 Все этапы склада</SelectItem>
                    {WAREHOUSE_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select value={locationFilter} onValueChange={setLocationFilter}>
                  <SelectTrigger className="h-9 text-xs sm:text-sm">
                    <SelectValue placeholder="Склад" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">📍 Все склады</SelectItem>
                    {uniqueLocations.map((loc) => (
                      <SelectItem key={loc} value={loc}>
                        {loc}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </>
            )}
          </div>
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
                {filteredOrders.map((order) => {
                  const total = calcOrderTotal(order);
                  const isCompleted = activeTab === 'completed';
                  return (
                    <div key={order.id} className="p-4 space-y-3">
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <div className="flex items-center gap-2 flex-wrap">
                            <span className="font-bold text-base">Заказ #{order.id}</span>
                            {order.source && (
                              <Badge variant="outline" className="text-xs">
                                {SOURCE_LABELS[order.source] || order.source}
                              </Badge>
                            )}
                            {order.order_number && (
                              <Badge variant="secondary" className="text-xs">
                                № {order.order_number}
                              </Badge>
                            )}
                          </div>
                          <div className="text-xs text-muted-foreground mt-0.5">
                            {isCompleted && order.completed_at_formatted
                              ? `Завершён: ${order.completed_at_formatted}`
                              : `${order.created_at_formatted} · ${order.time_ago}`}
                          </div>
                        </div>
                        <div className="text-right">
                          <div className="font-bold text-base">₽{total.toLocaleString('ru-RU')}</div>
                          {Number(order.discount) > 0 && (
                            <div className="text-[11px] text-amber-600">
                              скидка ₽{Number(order.discount).toLocaleString('ru-RU')}
                            </div>
                          )}
                        </div>
                      </div>

                      <div>{renderOrderParts(order, isCompleted)}</div>

                      <div className="grid grid-cols-2 gap-x-3 gap-y-1 text-xs sm:text-sm bg-muted/40 rounded-lg p-2.5">
                        <div className="text-muted-foreground">Покупатель:</div>
                        <div className="font-medium">{order.buyer_number}</div>
                        <div className="text-muted-foreground">Склад:</div>
                        <div className="font-medium">{order.location || 'Неизвестно'}</div>
                        <div className="text-muted-foreground">Продавец:</div>
                        <div className="font-medium">{order.seller}</div>
                        {(order.transport_company || order.tracking_number) && (
                          <>
                            <div className="text-muted-foreground">ТК / Трек:</div>
                            <div className="font-medium">
                              {[order.transport_company, order.tracking_number].filter(Boolean).join(' · ')}
                            </div>
                          </>
                        )}
                      </div>

                      {order.notes && (
                        <div className="flex items-start gap-1.5 text-xs bg-amber-500/10 border border-amber-500/20 text-amber-800 dark:text-amber-200 rounded-md p-2">
                          <MessageSquare className="h-3.5 w-3.5 shrink-0 mt-0.5" />
                          <span>{order.notes}</span>
                        </div>
                      )}

                      {!isCompleted && (
                        <>
                          <div className="grid grid-cols-3 gap-1.5">
                            <Select
                              value={order.payment_status || 'unpaid'}
                              onValueChange={(value) =>
                                quickUpdateMutation.mutate({ orderId: order.id, payload: { payment_status: value } })
                              }
                            >
                              <SelectTrigger className="h-8 text-xs px-2">
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                {PAYMENT_OPTIONS.map((opt) => (
                                  <SelectItem key={opt.value} value={opt.value}>
                                    {opt.label}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>

                            <Select
                              value={order.warehouse_status || 'inspecting'}
                              onValueChange={(value) =>
                                quickUpdateMutation.mutate({ orderId: order.id, payload: { warehouse_status: value } })
                              }
                            >
                              <SelectTrigger className="h-8 text-xs px-2">
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                {WAREHOUSE_OPTIONS.map((opt) => (
                                  <SelectItem key={opt.value} value={opt.value}>
                                    {opt.label}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>

                            <Select
                              value={order.delivery_method || 'pickup'}
                              onValueChange={(value) =>
                                quickUpdateMutation.mutate({ orderId: order.id, payload: { delivery_method: value } })
                              }
                            >
                              <SelectTrigger className="h-8 text-xs px-2">
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                {DELIVERY_OPTIONS.map((opt) => (
                                  <SelectItem key={opt.value} value={opt.value}>
                                    {opt.label}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                          </div>

                          <div className="flex gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => openEditModal(order)}
                              className="gap-1"
                            >
                              <Edit className="h-3.5 w-3.5" />
                              <span>Изменить</span>
                            </Button>
                            <Button
                              variant="default"
                              size="sm"
                              className="flex-1 gap-1"
                              onClick={() => setOrderToComplete(order)}
                            >
                              <CheckCircle2 className="h-3.5 w-3.5" />
                              <span>Завершить</span>
                            </Button>
                            <Button
                              variant="destructive"
                              size="sm"
                              onClick={() => setOrderToDelete(order)}
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        </>
                      )}
                    </div>
                  );
                })}
              </div>

              {/* Десктопный вид: таблица */}
              <div className="hidden lg:block rounded-md border overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-24">№ / Дата</TableHead>
                      <TableHead>Покупатель / Сделка</TableHead>
                      <TableHead>Состав заказа и Склад</TableHead>
                      <TableHead className="w-32">Сумма</TableHead>
                      {activeTab === 'active' ? (
                        <>
                          <TableHead className="w-44">💰 Оплата</TableHead>
                          <TableHead className="w-48">🏭 Склад</TableHead>
                          <TableHead className="w-48">🚚 Доставка / Трек</TableHead>
                          <TableHead className="text-right">Действия</TableHead>
                        </>
                      ) : (
                        <>
                          <TableHead>Логистика / Заметки</TableHead>
                          <TableHead>Продавец</TableHead>
                          <TableHead className="text-right">Действия</TableHead>
                        </>
                      )}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredOrders.map((order) => {
                      const total = calcOrderTotal(order);
                      const isCompleted = activeTab === 'completed';
                      return (
                        <TableRow key={order.id}>
                          <TableCell className="align-top">
                            <div className="font-bold">#{order.id}</div>
                            <div className="text-xs text-muted-foreground">
                              {isCompleted && order.completed_at_formatted
                                ? order.completed_at_formatted
                                : order.created_at_formatted}
                            </div>
                            {!isCompleted && (
                              <div className="text-[11px] text-muted-foreground">{order.time_ago}</div>
                            )}
                          </TableCell>

                          <TableCell className="align-top">
                            <div className="font-medium">{order.buyer_number}</div>
                            <div className="flex items-center gap-1.5 flex-wrap mt-1">
                              {order.source && (
                                <Badge variant="outline" className="text-[11px] px-1.5 py-0">
                                  {SOURCE_LABELS[order.source] || order.source}
                                </Badge>
                              )}
                              {order.order_number && (
                                <Badge variant="secondary" className="text-[11px] px-1.5 py-0">
                                  № {order.order_number}
                                </Badge>
                              )}
                            </div>
                            <div className="text-xs text-muted-foreground mt-1">
                              Продавец: {order.seller}
                            </div>
                          </TableCell>

                          <TableCell className="align-top">
                            {renderOrderParts(order, isCompleted)}
                            <div className="text-xs text-muted-foreground mt-1">
                              📍 Склад: <span className="font-medium text-foreground">{order.location || 'Неизвестно'}</span>
                            </div>
                            {order.notes && (
                              <div className="mt-1.5 inline-flex items-center gap-1 text-xs bg-amber-500/10 border border-amber-500/20 text-amber-800 dark:text-amber-200 px-2 py-0.5 rounded">
                                <MessageSquare className="h-3 w-3 shrink-0" />
                                <span>{order.notes}</span>
                              </div>
                            )}
                          </TableCell>

                          <TableCell className="align-top">
                            <div className="font-bold text-base">₽{total.toLocaleString('ru-RU')}</div>
                            {Number(order.discount) > 0 && (
                              <div className="text-xs text-amber-600">
                                −₽{Number(order.discount).toLocaleString('ru-RU')} скидка
                              </div>
                            )}
                          </TableCell>

                          {!isCompleted ? (
                            <>
                              <TableCell className="align-top">
                                <Select
                                  value={order.payment_status || 'unpaid'}
                                  onValueChange={(value) =>
                                    quickUpdateMutation.mutate({
                                      orderId: order.id,
                                      payload: { payment_status: value },
                                    })
                                  }
                                >
                                  <SelectTrigger className="h-8 text-xs">
                                    <SelectValue />
                                  </SelectTrigger>
                                  <SelectContent>
                                    {PAYMENT_OPTIONS.map((opt) => (
                                      <SelectItem key={opt.value} value={opt.value}>
                                        {opt.label}
                                      </SelectItem>
                                    ))}
                                  </SelectContent>
                                </Select>
                              </TableCell>

                              <TableCell className="align-top">
                                <Select
                                  value={order.warehouse_status || 'inspecting'}
                                  onValueChange={(value) =>
                                    quickUpdateMutation.mutate({
                                      orderId: order.id,
                                      payload: { warehouse_status: value },
                                    })
                                  }
                                >
                                  <SelectTrigger className="h-8 text-xs">
                                    <SelectValue />
                                  </SelectTrigger>
                                  <SelectContent>
                                    {WAREHOUSE_OPTIONS.map((opt) => (
                                      <SelectItem key={opt.value} value={opt.value}>
                                        {opt.label}
                                      </SelectItem>
                                    ))}
                                  </SelectContent>
                                </Select>
                              </TableCell>

                              <TableCell className="align-top">
                                <Select
                                  value={order.delivery_method || 'pickup'}
                                  onValueChange={(value) =>
                                    quickUpdateMutation.mutate({
                                      orderId: order.id,
                                      payload: { delivery_method: value },
                                    })
                                  }
                                >
                                  <SelectTrigger className="h-8 text-xs">
                                    <SelectValue />
                                  </SelectTrigger>
                                  <SelectContent>
                                    {DELIVERY_OPTIONS.map((opt) => (
                                      <SelectItem key={opt.value} value={opt.value}>
                                        {opt.label}
                                      </SelectItem>
                                    ))}
                                  </SelectContent>
                                </Select>
                                {(order.transport_company || order.tracking_number) && (
                                  <div className="mt-1 text-xs text-muted-foreground flex items-center gap-1">
                                    <Truck className="h-3 w-3 shrink-0" />
                                    <span className="truncate max-w-[160px]">
                                      {[order.transport_company, order.tracking_number]
                                        .filter(Boolean)
                                        .join(' · ')}
                                    </span>
                                  </div>
                                )}
                              </TableCell>

                              <TableCell className="align-top text-right">
                                <div className="flex items-center justify-end gap-1.5">
                                  <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => openEditModal(order)}
                                    title="Редактировать заказ, цену, ТК и трек-номер"
                                  >
                                    <Edit className="h-3.5 w-3.5" />
                                  </Button>
                                  <Button
                                    variant="default"
                                    size="sm"
                                    onClick={() => setOrderToComplete(order)}
                                    className="gap-1"
                                  >
                                    <CheckCircle2 className="h-3.5 w-3.5" />
                                    <span>Завершить</span>
                                  </Button>
                                  <Button
                                    variant="destructive"
                                    size="sm"
                                    onClick={() => setOrderToDelete(order)}
                                    title="Удалить заказ"
                                  >
                                    <Trash2 className="h-3.5 w-3.5" />
                                  </Button>
                                </div>
                              </TableCell>
                            </>
                          ) : (
                            <>
                              <TableCell className="align-top text-xs">
                                <div>
                                  {DELIVERY_OPTIONS.find((d) => d.value === order.delivery_method)?.label ||
                                    'Самовывоз'}
                                </div>
                                {(order.transport_company || order.tracking_number) && (
                                  <div className="text-muted-foreground mt-0.5">
                                    {[order.transport_company, order.tracking_number]
                                      .filter(Boolean)
                                      .join(' · ')}
                                  </div>
                                )}
                              </TableCell>
                              <TableCell className="align-top text-sm">{order.seller}</TableCell>
                              <TableCell className="align-top text-right">
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={() => openEditModal(order)}
                                  title="Изменить заметку или трек-номер"
                                >
                                  <Edit className="h-3.5 w-3.5 mr-1" />
                                  Детали
                                </Button>
                              </TableCell>
                            </>
                          )}
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </div>
            </>
          )}
        </CardContent>
      </Card>
      )}

      {/* Диалог редактирования заказа */}
      <Dialog open={Boolean(editingOrder)} onOpenChange={(open) => !open && setEditingOrder(null)}>
        <DialogContent className="sm:max-w-[620px] max-h-[92vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Редактирование заказа #{editingOrder?.id}</DialogTitle>
            <DialogDescription>
              Изменение реквизитов покупателя, логистики (ТК и трек-номера), скидки и цен позиций.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div className="space-y-1.5">
                <Label>Покупатель / Телефон</Label>
                <Input value={editBuyer} onChange={(e) => setEditBuyer(e.target.value)} />
              </div>
              <div className="space-y-1.5">
                <Label>№ сделки (Дром / Авито)</Label>
                <Input value={editOrderNumber} onChange={(e) => setEditOrderNumber(e.target.value)} />
              </div>
              <div className="space-y-1.5">
                <Label>Источник</Label>
                <Select value={editSource} onValueChange={setEditSource}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="drom">Дром</SelectItem>
                    <SelectItem value="avito">Авито</SelectItem>
                    <SelectItem value="messenger">Мессенджер</SelectItem>
                    <SelectItem value="pickup">Самовывоз</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div className="space-y-1.5">
                <Label>💰 Оплата</Label>
                <Select value={editPayment} onValueChange={setEditPayment}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {PAYMENT_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>🏭 Склад</Label>
                <Select value={editWarehouse} onValueChange={setEditWarehouse}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {WAREHOUSE_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>🚚 Способ выдачи</Label>
                <Select value={editDelivery} onValueChange={setEditDelivery}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {DELIVERY_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Транспортная компания / Город</Label>
                <Input
                  value={editTK}
                  onChange={(e) => setEditTK(e.target.value)}
                  placeholder="Например: СДЭК, Энергия..."
                />
              </div>
              <div className="space-y-1.5">
                <Label>Трек-номер (накладная ТК)</Label>
                <Input
                  value={editTracking}
                  onChange={(e) => setEditTracking(e.target.value)}
                  placeholder="Введите номер отслеживания"
                />
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div className="sm:col-span-2 space-y-1.5">
                <Label>Комментарий / Заметка для склада</Label>
                <Input
                  value={editNotes}
                  onChange={(e) => setEditNotes(e.target.value)}
                  placeholder="Примечания к сборке, упаковке или отправке"
                />
              </div>
              <div className="space-y-1.5">
                <Label>Скидка на заказ (₽)</Label>
                <Input
                  type="number"
                  min={0}
                  value={editDiscount}
                  onChange={(e) => setEditDiscount(Math.max(0, Number(e.target.value) || 0))}
                />
              </div>
            </div>

            {/* Позиции заказа */}
            {editItems.length > 0 && (
              <div className="border-t pt-3 space-y-2">
                <Label className="text-sm font-medium block">Позиции заказа (торг / количество):</Label>
                {editItems.map((item, idx) => (
                  <div
                    key={item.id}
                    className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 p-2.5 rounded-lg border bg-muted/40"
                  >
                    <div className="text-sm font-medium truncate flex-1">
                      {item.part_name || item.part_name_snapshot || `Запчасть #${item.part_id}`}
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="flex items-center gap-1">
                        <span className="text-xs text-muted-foreground">Кол-во:</span>
                        <Input
                          type="number"
                          min={1}
                          value={item.quantity}
                          onChange={(e) => {
                            const val = Math.max(1, Number(e.target.value) || 1);
                            setEditItems((prev) =>
                              prev.map((it, i) => (i === idx ? { ...it, quantity: val } : it))
                            );
                          }}
                          className="w-16 h-8 text-xs"
                        />
                      </div>
                      <div className="flex items-center gap-1">
                        <span className="text-xs text-muted-foreground">Цена ₽:</span>
                        <Input
                          type="number"
                          min={0}
                          value={item.price || 0}
                          onChange={(e) => {
                            const val = Math.max(0, Number(e.target.value) || 0);
                            setEditItems((prev) =>
                              prev.map((it, i) => (i === idx ? { ...it, price: val } : it))
                            );
                          }}
                          className="w-24 h-8 text-xs"
                        />
                      </div>
                      {editItems.length > 1 && editingOrder && !editingOrder.auto_deleted && (
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          className="text-destructive h-8 px-2"
                          onClick={() =>
                            removeOrderItemMutation.mutate({
                              orderId: editingOrder.id,
                              itemId: item.id,
                            })
                          }
                          title="Убрать позицию из заказа"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditingOrder(null)}>
              Отмена
            </Button>
            <Button
              onClick={() => saveOrderDetailsMutation.mutate()}
              disabled={saveOrderDetailsMutation.isPending}
            >
              {saveOrderDetailsMutation.isPending ? 'Сохранение...' : 'Сохранить изменения'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Диалог подтверждения завершения */}
      <AlertDialog
        open={Boolean(orderToComplete)}
        onOpenChange={(open) => !open && setOrderToComplete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Завершить сделку по заказу #{orderToComplete?.id}?</AlertDialogTitle>
            <AlertDialogDescription>
              Заказанное количество будет списано со склада (если остаток запчасти исчерпан — карточка запчасти и её фото будут удалены из инвентаря), а сам заказ перейдёт во вкладку «История продаж» и учтётся в статистике на сумму{' '}
              <strong>₽{orderToComplete ? calcOrderTotal(orderToComplete).toLocaleString('ru-RU') : 0}</strong>.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (orderToComplete) {
                  completeOrderMutation.mutate(orderToComplete.id);
                }
              }}
            >
              Завершить продажу
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Диалог подтверждения удаления */}
      <AlertDialog
        open={Boolean(orderToDelete)}
        onOpenChange={(open) => !open && setOrderToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Отменить и удалить заказ #{orderToDelete?.id}?</AlertDialogTitle>
            <AlertDialogDescription>
              Заказ #{orderToDelete?.id} ({orderToDelete?.buyer_number}) будет удалён из активных заказов, а запчасти останутся на складе.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Назад</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (orderToDelete) {
                  deleteOrderMutation.mutate(orderToDelete.id);
                }
              }}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Удалить заказ
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};

export default Orders;