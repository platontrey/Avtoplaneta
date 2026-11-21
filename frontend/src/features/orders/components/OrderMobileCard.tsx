/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Edit, Trash2, CheckCircle2 } from 'lucide-react';
import type { Order } from '@/lib/types';
import { SOURCE_LABELS, PAYMENT_OPTIONS, WAREHOUSE_OPTIONS, DELIVERY_OPTIONS, normalizeWarehouseStatus } from '../constants';
import { calcOrderTotal, renderOrderParts } from '../utils/orderHelpers';

interface OrderMobileCardProps {
  order: Order;
  isCompleted: boolean;
  onQuickUpdate: (orderId: number, payload: { payment_status?: string; warehouse_status?: string; delivery_method?: string }) => void;
  onOpenEdit: (order: Order) => void;
  onComplete: (order: Order) => void;
  onDelete: (order: Order) => void;
  userRole?: string;
}

export const OrderMobileCard: React.FC<OrderMobileCardProps> = ({
  order,
  isCompleted,
  onQuickUpdate,
  onOpenEdit,
  onComplete,
  onDelete,
  userRole,
}) => {
  const total = calcOrderTotal(order);

  return (
    <div className="p-4 space-y-3">
      <div className="flex items-start justify-between gap-2">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-bold text-base">#{order.id}</span>
            {order.order_number && (
              <Badge variant="secondary" className="text-[11px] px-1.5 py-0">
                № {order.order_number}
              </Badge>
            )}
            {order.source && (
              <Badge variant="outline" className="text-[11px] px-1.5 py-0">
                {SOURCE_LABELS[order.source] || order.source}
              </Badge>
            )}
          </div>
          <div className="text-xs text-muted-foreground mt-0.5">
            {isCompleted && order.completed_at_formatted
              ? order.completed_at_formatted
              : order.created_at_formatted}
            {!isCompleted && ` (${order.time_ago})`}
          </div>
        </div>

        <div className="text-right">
          <div className="text-base font-bold text-emerald-600 dark:text-emerald-400">
            ₽{total.toLocaleString('ru-RU')}
          </div>
          {Boolean(order.discount && order.discount > 0) && (
            <div className="text-[11px] text-muted-foreground line-through">
              Скидка: ₽{(order.discount || 0).toLocaleString('ru-RU')}
            </div>
          )}
        </div>
      </div>

      <div>
        <div className="text-sm font-medium">{order.buyer_number}</div>
        <div className="text-xs text-muted-foreground">Продавец: {order.seller}</div>
      </div>

      <div className="rounded-md bg-muted/40 p-2.5 text-xs space-y-1">
        {renderOrderParts(order, isCompleted)}
        <div className="text-muted-foreground pt-1 border-t border-border/50">
          📍 Склад: <span className="font-medium text-foreground">{order.location || 'Неизвестно'}</span>
        </div>
        {order.notes && (
          <div className="text-muted-foreground italic pt-0.5">💬 {order.notes}</div>
        )}
      </div>

      {!isCompleted && (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
            <Select
              value={order.payment_status || 'unpaid'}
              onValueChange={(value) =>
                onQuickUpdate(order.id, { payment_status: value })
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
              value={normalizeWarehouseStatus(order.warehouse_status)}
              onValueChange={(value) =>
                onQuickUpdate(order.id, { warehouse_status: value })
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
                onQuickUpdate(order.id, { delivery_method: value })
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
              onClick={() => onOpenEdit(order)}
              className="gap-1"
            >
              <Edit className="h-3.5 w-3.5" />
              <span>Изменить</span>
            </Button>
            <Button
              variant="default"
              size="sm"
              className="flex-1 gap-1"
              onClick={() => onComplete(order)}
            >
              <CheckCircle2 className="h-3.5 w-3.5" />
              <span>Завершить</span>
            </Button>
            {userRole === 'admin' && (
              <Button
                variant="destructive"
                size="sm"
                onClick={() => onDelete(order)}
                title="Удалить заказ"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            )}
          </div>
        </>
      )}
    </div>
  );
};
