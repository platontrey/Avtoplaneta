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

  // Количество заказов, требующих внимания склада (осмотр/перемещение)
  const warehouseAttentionCount = orders.filter(
    (o) => o.warehouse_status === 'inspecting' || o.warehouse_status === 'transfer'
  ).length;

  // Неоплаченные активные заказы
  const unpaidCount = orders.filter((o) => (o.payment_status || 'unpaid') === 'unpaid').length;

  // Заказы, готовые к выдаче
  const readyCount = orders.filter((o) => o.warehouse_status === 'ready').length;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
      <Card className="bg-card/70 border-border shadow-xs">
        <CardContent className="p-4 flex items-center gap-3.5">
          <div className="p-2.5 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            <DollarSign className="w-5 h-5" />
          </div>
          <div>
            <p className="text-xs text-muted-foreground font-medium">Выручка за {currentMonth}</p>
            <p className="text-xl font-bold tracking-tight text-foreground">
              ₽{currentMonthSales.toLocaleString('ru-RU')}
            </p>
          </div>
        </CardContent>
      </Card>

      <Card className="bg-card/70 border-border shadow-xs">
        <CardContent className="p-4 flex items-center gap-3.5">
          <div className="p-2.5 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400">
            <PackageCheck className="w-5 h-5" />
          </div>
          <div>
            <p className="text-xs text-muted-foreground font-medium">Готовы к выдаче / отгрузке</p>
            <p className="text-xl font-bold tracking-tight text-foreground">{readyCount}</p>
          </div>
        </CardContent>
      </Card>

      <Card className="bg-card/70 border-border shadow-xs">
        <CardContent className="p-4 flex items-center gap-3.5">
          <div className="p-2.5 rounded-xl bg-amber-500/10 text-amber-600 dark:text-amber-400">
            <AlertTriangle className="w-5 h-5" />
          </div>
          <div>
            <p className="text-xs text-muted-foreground font-medium">На складе (осмотр / фото)</p>
            <p className="text-xl font-bold tracking-tight text-foreground">{warehouseAttentionCount}</p>
          </div>
        </CardContent>
      </Card>

      <Card className="bg-card/70 border-border shadow-xs">
        <CardContent className="p-4 flex items-center gap-3.5">
          <div className="p-2.5 rounded-xl bg-rose-500/10 text-rose-600 dark:text-rose-400">
            <Clock className="w-5 h-5" />
          </div>
          <div>
            <p className="text-xs text-muted-foreground font-medium">Ожидают оплаты</p>
            <p className="text-xl font-bold tracking-tight text-foreground">{unpaidCount}</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
