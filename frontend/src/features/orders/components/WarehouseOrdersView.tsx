/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState, useMemo } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
  MapPin,
  Camera,
  CheckCircle2,
  Printer,
  Package,
  Boxes,
  Search,
  AlertCircle,
  Truck,
} from 'lucide-react';
import type { Order } from '@/lib/types';
import { WAREHOUSE_OPTIONS, DELIVERY_OPTIONS } from '../constants';

interface WarehouseOrdersViewProps {
  orders: Order[];
  onUpdateStatus: (orderId: number, warehouseStatus: string) => void;
  isUpdating?: boolean;
}

export const WarehouseOrdersView: React.FC<WarehouseOrdersViewProps> = ({
  orders,
  onUpdateStatus,
  isUpdating = false,
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'inspecting' | 'picking' | 'packed'>('all');
  const [pickedItems, setPickedItems] = useState<Record<string, boolean>>({});

  // Фильтрация заданий для склада
  const filteredOrders = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    return orders.filter((order) => {
      const ws = order.warehouse_status || 'pending';
      if (statusFilter !== 'all') {
        if (statusFilter === 'picking' && (ws === 'picking' || ws === 'transfer')) {
          // match
        } else if (statusFilter === 'packed' && (ws === 'packed' || ws === 'ready')) {
          // match
        } else if (ws !== statusFilter) {
          return false;
        }
      }
      if (!q) return true;

      const inParts = (order.items || []).some(
        (it) =>
          (it.part_name || '').toLowerCase().includes(q) ||
          (it.part_name_snapshot || '').toLowerCase().includes(q) ||
          (it.location || '').toLowerCase().includes(q) ||
          String(it.part_id).includes(q)
      );

      return (
        String(order.id).includes(q) ||
        (order.order_number || '').toLowerCase().includes(q) ||
        (order.location || '').toLowerCase().includes(q) ||
        (order.part || '').toLowerCase().includes(q) ||
        (order.notes || '').toLowerCase().includes(q) ||
        inParts
      );
    });
  }, [orders, statusFilter, searchQuery]);

  const toggleItemPicked = (key: string) => {
    setPickedItems((prev) => ({
      ...prev,
      [key]: !prev[key],
    }));
  };

  const handlePrintAssemblySheet = (order: Order) => {
    const printWindow = window.open('', '_blank', 'width=800,height=700');
    if (!printWindow) return;

    const itemsHtml = (order.items && order.items.length > 0 ? order.items : [
      {
        id: 0,
        order_id: order.id,
        part_id: order.part_id,
        part_name: order.part,
        part_name_snapshot: order.part,
        location: order.location,
        quantity: 1,
        price: 0,
      },
    ])
      .map(
        (item, idx) => `
        <tr style="border-bottom: 1px solid #ddd;">
          <td style="padding: 10px; font-weight: bold;">#${idx + 1}</td>
          <td style="padding: 10px;">
            <div style="font-weight: bold; font-size: 14px;">${item.part_name_snapshot || item.part_name || order.part}</div>
            <div style="font-size: 12px; color: #555;">ID детали: ${item.part_id}</div>
          </td>
          <td style="padding: 10px; font-weight: bold; color: #b45309; font-size: 15px;">
            📍 ${item.location || order.location || 'Не указана'}
          </td>
          <td style="padding: 10px; text-align: center; font-size: 16px; font-weight: bold;">
            ${item.quantity || 1} шт.
          </td>
          <td style="padding: 10px; text-align: center;">
            <div style="width: 20px; height: 20px; border: 2px solid #333; margin: auto;"></div>
          </td>
        </tr>
      `
      )
      .join('');

    printWindow.document.write(`
      <!DOCTYPE html>
      <html>
        <head>
          <title>Сборочный лист — Заказ #${order.order_number || order.id}</title>
          <style>
            body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; padding: 25px; color: #111; }
            .header { border-bottom: 2px solid #000; padding-bottom: 12px; margin-bottom: 20px; display: flex; justify-content: space-between; align-items: flex-start; }
            .badge { display: inline-block; padding: 4px 8px; background: #eee; border-radius: 4px; font-size: 12px; font-weight: bold; }
            table { width: 100%; border-collapse: collapse; margin-top: 15px; }
            th { text-align: left; padding: 10px; background: #f4f4f4; border-bottom: 2px solid #bbb; }
            .footer { margin-top: 30px; border-top: 1px dashed #999; padding-top: 15px; display: flex; justify-content: space-between; font-size: 13px; color: #444; }
          </style>
        </head>
        <body>
          <div class="header">
            <div>
              <h1 style="margin: 0 0 5px 0; font-size: 24px;">СБОРОЧНЫЙ ЛИСТ СКЛАДА</h1>
              <div style="font-size: 16px; font-weight: bold;">Заказ № ${order.order_number || order.id} (ID: ${order.id})</div>
              <div style="font-size: 12px; color: #666; margin-top: 4px;">Дата: ${order.created_at_formatted || 'Сегодня'}</div>
            </div>
            <div style="text-align: right;">
              <span class="badge">Avtoplaneta WMS</span>
              <div style="margin-top: 8px; font-size: 13px;"><b>Доставка:</b> ${order.delivery_method || 'Самовывоз'}</div>
              ${order.transport_company ? `<div style="font-size: 12px;"><b>ТК:</b> ${order.transport_company}</div>` : ''}
            </div>
          </div>

          ${order.notes ? `<div style="background: #fffbeb; border: 1px solid #fde68a; padding: 10px; border-radius: 6px; margin-bottom: 15px; font-size: 13px;"><b>Заметка менеджера:</b> ${order.notes}</div>` : ''}

          <table>
            <thead>
              <tr>
                <th style="width: 40px;">№</th>
                <th>Наименование запчасти</th>
                <th style="width: 220px;">Ячейка хранения</th>
                <th style="width: 80px; text-align: center;">Кол-во</th>
                <th style="width: 60px; text-align: center;">Факт</th>
              </tr>
            </thead>
            <tbody>
              ${itemsHtml}
            </tbody>
          </table>

          <div class="footer">
            <div>Сборщик: _____________________ (подпись)</div>
            <div>Контролер упаковки: _____________________ (подпись)</div>
          </div>
          <script>
            window.onload = function() { window.print(); window.close(); }
          </script>
        </body>
      </html>
    `);
    printWindow.document.close();
  };

  return (
    <div className="space-y-4">
      {/* Панель фильтров склада */}
      <div className="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center justify-between">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Поиск по № заказа, детали, стеллажу или ячейке..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 bg-card"
          />
        </div>

        <div className="inline-flex rounded-lg border bg-muted p-1">
          <Button
            variant={statusFilter === 'all' ? 'default' : 'ghost'}
            size="sm"
            className="text-xs"
            onClick={() => setStatusFilter('all')}
          >
            Все ({orders.length})
          </Button>
          <Button
            variant={statusFilter === 'pending' ? 'default' : 'ghost'}
            size="sm"
            className="text-xs gap-1"
            onClick={() => setStatusFilter('pending')}
          >
            <Package className="w-3.5 h-3.5 text-slate-500" />
            <span>Очередь ({orders.filter((o) => (o.warehouse_status || 'pending') === 'pending').length})</span>
          </Button>
          <Button
            variant={statusFilter === 'inspecting' ? 'default' : 'ghost'}
            size="sm"
            className="text-xs gap-1"
            onClick={() => setStatusFilter('inspecting')}
          >
            <Camera className="w-3.5 h-3.5 text-cyan-500" />
            <span>Осмотр / Фото ({orders.filter((o) => o.warehouse_status === 'inspecting').length})</span>
          </Button>
          <Button
            variant={statusFilter === 'picking' ? 'default' : 'ghost'}
            size="sm"
            className="text-xs gap-1"
            onClick={() => setStatusFilter('picking')}
          >
            <Boxes className="w-3.5 h-3.5 text-indigo-500" />
            <span>В сборке ({orders.filter((o) => o.warehouse_status === 'picking' || o.warehouse_status === 'transfer').length})</span>
          </Button>
          <Button
            variant={statusFilter === 'packed' ? 'default' : 'ghost'}
            size="sm"
            className="text-xs gap-1"
            onClick={() => setStatusFilter('packed')}
          >
            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
            <span>Собраны ({orders.filter((o) => o.warehouse_status === 'packed' || o.warehouse_status === 'ready').length})</span>
          </Button>
        </div>
      </div>

      {/* Список сборочных заданий */}
      {filteredOrders.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            <Package className="w-12 h-12 mx-auto mb-3 opacity-30" />
            <p className="text-base font-medium">Нет заказов в выбранной очереди склада</p>
            <p className="text-xs mt-1">Все позиции проверены или измените условия поиска</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filteredOrders.map((order) => {
            const currentStatus = order.warehouse_status || 'inspecting';
            const deliveryLabel =
              DELIVERY_OPTIONS.find((d) => d.value === order.delivery_method)?.label || 'Самовывоз';

            const items =
              order.items && order.items.length > 0
                ? order.items
                : [
                    {
                      id: 0,
                      order_id: order.id,
                      part_id: order.part_id,
                      part_name: order.part,
                      part_name_snapshot: order.part,
                      location: order.location,
                      quantity: 1,
                      price: 0,
                    },
                  ];

            return (
              <Card
                key={order.id}
                className="overflow-hidden border-border bg-card shadow-xs hover:border-border/80 transition-all flex flex-col justify-between"
              >
                <div>
                  {/* Верхняя плашка карточки */}
                  <div className="p-4 border-b border-border bg-muted/30 flex items-start justify-between gap-2">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-lg font-bold">
                          Заказ №{order.order_number || order.id}
                        </span>
                        <Badge variant="outline" className="text-xs">
                          {order.created_at_formatted || 'Сегодня'}
                        </Badge>
                      </div>
                      <div className="text-xs text-muted-foreground flex items-center gap-2 mt-1">
                        <span className="flex items-center gap-1 font-medium text-foreground">
                          <Truck className="w-3.5 h-3.5 text-blue-500" />
                          {deliveryLabel}
                        </span>
                        {order.transport_company && (
                          <span>· ТК: {order.transport_company}</span>
                        )}
                        <span>· Менеджер: {order.seller || 'Не назначен'}</span>
                      </div>
                    </div>

                    <Button
                      variant="outline"
                      size="sm"
                      className="h-8 gap-1.5 text-xs shrink-0"
                      onClick={() => handlePrintAssemblySheet(order)}
                      title="Печать сборочного листа для склада"
                    >
                      <Printer className="w-3.5 h-3.5" />
                      <span>Печать</span>
                    </Button>
                  </div>

                  {/* Основное тело с позициями деталей и адресами ячеек */}
                  <div className="p-4 space-y-3">
                    {order.notes && (
                      <div className="p-2.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-xs text-amber-800 dark:text-amber-200 flex items-start gap-2">
                        <AlertCircle className="w-4 h-4 shrink-0 text-amber-600 mt-0.5" />
                        <div>
                          <span className="font-semibold">Указание менеджера:</span> {order.notes}
                        </div>
                      </div>
                    )}

                    <div className="space-y-2">
                      <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                        Позиции к сборке ({items.length}):
                      </p>

                      {items.map((item, idx) => {
                        const itemKey = `${order.id}-${item.id || idx}`;
                        const isPicked = Boolean(pickedItems[itemKey]);

                        return (
                          <div
                            key={itemKey}
                            onClick={() => toggleItemPicked(itemKey)}
                            className={`p-3 rounded-lg border transition-all cursor-pointer flex items-center justify-between gap-3 ${
                              isPicked
                                ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-950 dark:text-emerald-100'
                                : 'bg-background border-border hover:border-primary/40'
                            }`}
                          >
                            <div className="flex items-start gap-2.5 min-w-0">
                              <input
                                type="checkbox"
                                checked={isPicked}
                                onChange={() => toggleItemPicked(itemKey)}
                                className="mt-1 h-4 w-4 rounded border-border text-primary focus:ring-primary shrink-0"
                                onClick={(e) => e.stopPropagation()}
                              />
                              <div className="min-w-0">
                                <p
                                  className={`text-sm font-semibold truncate ${
                                    isPicked ? 'line-through opacity-75' : ''
                                  }`}
                                >
                                  {item.part_name_snapshot || item.part_name || order.part}
                                </p>
                                <p className="text-xs text-muted-foreground">
                                  ID: {item.part_id} · Количество: {item.quantity || 1} шт.
                                </p>
                              </div>
                            </div>

                            {/* Ячейка хранения */}
                            <div className="text-right shrink-0">
                              <Badge
                                variant="secondary"
                                className="text-xs font-bold gap-1 bg-amber-500/15 text-amber-800 dark:text-amber-300 border-amber-500/30"
                              >
                                <MapPin className="w-3 h-3 text-amber-600" />
                                <span>{item.location || order.location || 'Склад'}</span>
                              </Badge>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                </div>

                {/* Нижняя панель смены складского статуса */}
                <div className="p-3 border-t border-border bg-muted/20 flex flex-wrap items-center justify-between gap-2">
                  <div className="text-xs text-muted-foreground">
                    Текущий статус:{' '}
                    <span className="font-semibold text-foreground">
                      {WAREHOUSE_OPTIONS.find((w) => w.value === currentStatus)?.label || currentStatus}
                    </span>
                  </div>

                  <div className="flex items-center gap-1.5 flex-wrap">
                    {currentStatus !== 'inspecting' && (
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={isUpdating}
                        onClick={() => onUpdateStatus(order.id, 'inspecting')}
                        className="text-xs h-7 px-2"
                      >
                        <Camera className="w-3 h-3 mr-1 text-amber-500" />
                        Осмотр
                      </Button>
                    )}

                    {currentStatus !== 'transfer' && (
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={isUpdating}
                        onClick={() => onUpdateStatus(order.id, 'transfer')}
                        className="text-xs h-7 px-2"
                      >
                        <Boxes className="w-3 h-3 mr-1 text-blue-500" />
                        В сборку
                      </Button>
                    )}

                    {currentStatus !== 'ready' && (
                      <Button
                        variant="default"
                        size="sm"
                        disabled={isUpdating}
                        onClick={() => onUpdateStatus(order.id, 'ready')}
                        className="text-xs h-7 px-2.5 bg-emerald-600 hover:bg-emerald-700 text-white"
                      >
                        <CheckCircle2 className="w-3.5 h-3.5 mr-1" />
                        Собран / Готов
                      </Button>
                    )}
                  </div>
                </div>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
};
