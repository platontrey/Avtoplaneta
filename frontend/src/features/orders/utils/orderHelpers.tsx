/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Link } from 'react-router-dom';
import { ExternalLink } from 'lucide-react';
import type { Order } from '@/lib/types';

export const calcOrderTotal = (order: Order): number => {
  if (typeof order.total_amount === 'number' && order.total_amount >= 0) {
    return order.total_amount;
  }
  const subtotal = (order.items || []).reduce(
    (acc, it) => acc + (Number(it.price) || 0) * (Number(it.quantity) || 1),
    0
  );
  return Math.max(0, subtotal - (Number(order.discount) || 0));
};

export const renderOrderParts = (order: Order, isCompleted: boolean): React.ReactNode => {
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
