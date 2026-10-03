/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Edit, Trash2, CheckCircle2, MessageSquare } from 'lucide-react';
import type { Order } from '@/lib/types';
import { SOURCE_LABELS, PAYMENT_OPTIONS, WAREHOUSE_OPTIONS, DELIVERY_OPTIONS } from '../constants';
import { calcOrderTotal, renderOrderParts } from '../utils/orderHelpers';

interface OrderMobileCardProps {
  order: Order;
  isCompleted: boolean;
  onQuickUpdate: (orderId: number, payload: { payment_status?: string; warehouse_status?: string; delivery_method?: string }) => void;
  onOpenEdit: (order: Order) => void;
  onComplete: (order: Order) => void;
  onDelete: (order: Order) => void;
}

export const OrderMobileCard: React.FC<OrderMobileCardProps> = ({
  order,
  isCompleted,
  onQuickUpdate,
  onOpenEdit,
  onComplete,
  onDelete,
}) => {
  const total = calcOrderTotal(order);

  return (
    <div className="p-4 space-y-3">
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
              value={order.warehouse_status || 'inspecting'}
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
            <Button
              variant="destructive"
              size="sm"
              onClick={() => onDelete(order)}
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        </>
      )}
    </div>
  );
};
