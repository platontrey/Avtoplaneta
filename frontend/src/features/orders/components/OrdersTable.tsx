/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Edit, Trash2, CheckCircle2, MessageSquare, Truck } from 'lucide-react';
import type { Order } from '@/lib/types';
import { SOURCE_LABELS, PAYMENT_OPTIONS, WAREHOUSE_OPTIONS, DELIVERY_OPTIONS } from '../constants';
import { calcOrderTotal, renderOrderParts } from '../utils/orderHelpers';

interface OrdersTableProps {
  orders: Order[];
  activeTab: 'active' | 'completed' | 'customers';
  onQuickUpdate: (orderId: number, payload: { payment_status?: string; warehouse_status?: string; delivery_method?: string }) => void;
  onOpenEdit: (order: Order) => void;
  onComplete: (order: Order) => void;
  onDelete: (order: Order) => void;
}

export const OrdersTable: React.FC<OrdersTableProps> = ({
  orders,
  activeTab,
  onQuickUpdate,
  onOpenEdit,
  onComplete,
  onDelete,
}) => {
  const isCompleted = activeTab === 'completed';

  return (
    <div className="hidden lg:block rounded-md border overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-24">№ / Дата</TableHead>
            <TableHead>Покупатель / Сделка</TableHead>
            <TableHead>Состав заказа и Склад</TableHead>
            <TableHead className="w-32">Сумма</TableHead>
            {!isCompleted ? (
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
          {orders.map((order) => {
            const total = calcOrderTotal(order);
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
                          onQuickUpdate(order.id, { payment_status: value })
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
                          onQuickUpdate(order.id, { warehouse_status: value })
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
                          onQuickUpdate(order.id, { delivery_method: value })
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
                          onClick={() => onOpenEdit(order)}
                          title="Редактировать заказ, цену, ТК и трек-номер"
                        >
                          <Edit className="h-3.5 w-3.5" />
                        </Button>
                        <Button
                          variant="default"
                          size="sm"
                          onClick={() => onComplete(order)}
                          className="gap-1"
                        >
                          <CheckCircle2 className="h-3.5 w-3.5" />
                          <span>Завершить</span>
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => onDelete(order)}
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
                        onClick={() => onOpenEdit(order)}
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
  );
};
