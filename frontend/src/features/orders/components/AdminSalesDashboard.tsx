/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent } from '@/components/ui/card';
import { DollarSign, Clock, AlertTriangle, PackageCheck } from 'lucide-react';
import { getMonthlySales } from '../api/ordersApi';
import type { Order } from '@/lib/types';

interface AdminSalesDashboardProps {
  orders: Order[];
}

export const AdminSalesDashboard: React.FC<AdminSalesDashboardProps> = ({ orders }) => {
  const { data: salesData } = useQuery({
    queryKey: ['orders', 'monthly-sales'],
    queryFn: getMonthlySales,
    staleTime: 60 * 1000,
  });

  // Текущий месяц в формате YYYY-MM
  const currentMonth = new Date().toISOString().slice(0, 7);
  const currentMonthSales = salesData?.find((e) => e.month === currentMonth)?.sales ?? 0;

  // Количество заказов, требующих внимания склада (в очереди / осмотр / сборка)
  const warehouseAttentionCount = orders.filter(
    (o) =>
      o.warehouse_status === 'pending' ||
      o.warehouse_status === 'inspecting' ||
      o.warehouse_status === 'picking' ||
      o.warehouse_status === 'transfer'
  ).length;

  // Неоплаченные активные заказы
  const unpaidCount = orders.filter((o) => (o.payment_status || 'unpaid') === 'unpaid').length;

  // Заказы, готовые к выдаче
  const readyCount = orders.filter(
    (o) => o.warehouse_status === 'packed' || o.warehouse_status === 'ready'
  ).length;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
      <Card className="bg-card/70 border-border shadow-xs">
        <CardContent className="p-3.5 flex items-center justify-between">
          <div>
            <p className="text-xs font-medium text-muted-foreground">Выручка (тек. месяц)</p>
            <p className="text-lg font-bold text-foreground">
              ₽{currentMonthSales.toLocaleString('ru-RU')}
            </p>
          </div>
          <div className="h-9 w-9 rounded-full bg-emerald-500/10 flex items-center justify-center text-emerald-600 dark:text-emerald-400">
            <DollarSign className="h-4 w-4" />
          </div>
        </CardContent>
      </Card>

      <Card className="bg-card/70 border-border shadow-xs">
        <CardContent className="p-3.5 flex items-center justify-between">
          <div>
            <p className="text-xs font-medium text-muted-foreground">На складе в работе</p>
            <p className="text-lg font-bold text-foreground">{warehouseAttentionCount}</p>
          </div>
          <div className="h-9 w-9 rounded-full bg-cyan-500/10 flex items-center justify-center text-cyan-600 dark:text-cyan-400">
            <Clock className="h-4 w-4" />
          </div>
        </CardContent>
      </Card>

      <Card className="bg-card/70 border-border shadow-xs">
        <CardContent className="p-3.5 flex items-center justify-between">
          <div>
            <p className="text-xs font-medium text-muted-foreground">Ожидают оплаты</p>
            <p className="text-lg font-bold text-foreground">{unpaidCount}</p>
          </div>
          <div className="h-9 w-9 rounded-full bg-amber-500/10 flex items-center justify-center text-amber-600 dark:text-amber-400">
            <AlertTriangle className="h-4 w-4" />
          </div>
        </CardContent>
      </Card>

      <Card className="bg-card/70 border-border shadow-xs">
        <CardContent className="p-3.5 flex items-center justify-between">
          <div>
            <p className="text-xs font-medium text-muted-foreground">Готовы к выдаче / отправке</p>
            <p className="text-lg font-bold text-foreground">{readyCount}</p>
          </div>
          <div className="h-9 w-9 rounded-full bg-blue-500/10 flex items-center justify-center text-blue-600 dark:text-blue-400">
            <PackageCheck className="h-4 w-4" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
