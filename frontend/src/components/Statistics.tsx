/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { Package, Boxes, DollarSign, TrendingUp, TrendingDown, ChartNoAxesCombined, RefreshCw, Calendar, Warehouse, Receipt } from 'lucide-react';
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useTheme } from '@/contexts/ThemeContext';
import { useStatistics } from '@/hooks/useStatistics';
import type { StatisticsResponse } from '@/lib/types';

type StatisticsPayload = Partial<StatisticsResponse> & {
  total_parts?: unknown;
  total_quantity?: unknown;
  total_value?: unknown;
  total_earnings?: unknown;
  parts_growth?: unknown;
  quantity_growth?: unknown;
  value_growth?: unknown;
  earnings_growth?: unknown;
  monthly_sales?: unknown;
};

const toFiniteNumber = (value: unknown): number => {
  const number = typeof value === 'number' ? value : Number(value);
  return Number.isFinite(number) ? number : 0;
};

const normalizeStatistics = (value: unknown): StatisticsResponse => {
  const payload = value && typeof value === 'object' ? value as StatisticsPayload : {};
  const categories = Array.isArray(payload.categories)
    ? payload.categories
      .filter((category): category is { name: string; count: number } => Boolean(category && typeof category.name === 'string'))
      .map((category) => ({ name: category.name, count: toFiniteNumber(category.count) }))
    : [];
  const monthlySalesSource = Array.isArray(payload.monthlySales)
    ? payload.monthlySales
    : Array.isArray(payload.monthly_sales)
      ? payload.monthly_sales
      : [];
  const monthlySales = monthlySalesSource
    .filter((item): item is { month: string; sales: number } => Boolean(item && typeof item === 'object' && 'month' in item && typeof item.month === 'string'))
    .map((item) => ({ month: item.month, sales: toFiniteNumber(item.sales) }));

  const partsGrowth = payload.partsGrowth !== undefined || payload.parts_growth !== undefined
    ? toFiniteNumber(payload.partsGrowth ?? payload.parts_growth)
    : undefined;
  const quantityGrowth = payload.quantityGrowth !== undefined || payload.quantity_growth !== undefined
    ? toFiniteNumber(payload.quantityGrowth ?? payload.quantity_growth)
    : undefined;
  const valueGrowth = payload.valueGrowth !== undefined || payload.value_growth !== undefined
    ? toFiniteNumber(payload.valueGrowth ?? payload.value_growth)
    : undefined;
  const earningsGrowth = payload.earningsGrowth !== undefined || payload.earnings_growth !== undefined
    ? toFiniteNumber(payload.earningsGrowth ?? payload.earnings_growth)
    : undefined;

  return {
    totalParts: toFiniteNumber(payload.totalParts ?? payload.total_parts),
    totalQuantity: toFiniteNumber(payload.totalQuantity ?? payload.total_quantity),
    totalValue: toFiniteNumber(payload.totalValue ?? payload.total_value),
    totalEarnings: toFiniteNumber(payload.totalEarnings ?? payload.total_earnings),
    partsGrowth,
    quantityGrowth,
    valueGrowth,
    earningsGrowth,
    categories,
    monthlySales,
  };
};

function Statistics() {
  const { data: rawData, isLoading, isFetching, refetch } = useStatistics();
  const [activeTab, setActiveTab] = useState<'sales' | 'inventory'>('sales');
  const { resolvedTheme } = useTheme();

  const data = rawData ? normalizeStatistics(rawData) : null;

  const GrowthIndicator = ({ growth }: { growth?: number }) => {
    if (growth === undefined) return null;
    const isZero = Math.abs(growth) < 0.05;
    const Icon = growth >= 0 ? TrendingUp : TrendingDown;
    const color = isZero
      ? 'text-muted-foreground'
      : growth > 0
        ? 'text-emerald-600 dark:text-emerald-400'
        : 'text-rose-600 dark:text-rose-400';
    const sign = growth > 0 ? '+' : '';
    return (
      <div className="absolute top-3 right-3 bg-muted/80 backdrop-blur-sm rounded-md px-2 py-1 flex items-center gap-1">
        {!isZero && <Icon className={`h-3.5 w-3.5 ${color}`} />}
        <span className={`text-xs font-semibold ${color}`}>{sign}{growth.toFixed(1)}%</span>
      </div>
    );
  };

  const formatPrice = (price: number) => {
    return toFiniteNumber(price).toLocaleString('ru-RU', { minimumFractionDigits: 0, maximumFractionDigits: 0 });
  };

  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 py-6 space-y-8">
        <div className="space-y-2">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-96" />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} className="p-4 space-y-3">
              <Skeleton className="h-8 w-8 rounded-lg" />
              <Skeleton className="h-4 w-28" />
              <Skeleton className="h-8 w-32" />
            </Card>
          ))}
        </div>
        <Card className="p-6">
          <Skeleton className="h-6 w-48 mb-4" />
          <Skeleton className="h-72 w-full" />
        </Card>
      </div>
    );
  }

  const tooltipContentStyle = {
    backgroundColor: 'var(--card)',
    borderColor: 'var(--border)',
    color: 'var(--card-foreground)',
    borderRadius: '0.5rem',
    boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)',
  };
  const tooltipItemStyle = {
    color: 'var(--card-foreground)',
  };
  const axisTickStyle = {
    fill: 'var(--muted-foreground)',
    fontSize: 12,
  };
  const chartPalette = resolvedTheme === 'dark'
    ? ['#38bdf8', '#818cf8', '#c084fc', '#34d399', '#fbbf24', '#f472b6', '#a78bfa']
    : ['#2563eb', '#4f46e5', '#7c3aed', '#059669', '#d97706', '#db2777', '#6366f1'];

  const categoryData = data?.categories?.filter(cat => cat && cat.name).map((cat, index) => ({
    name: cat.name,
    count: cat.count,
    fill: chartPalette[index % chartPalette.length]
  })) || [];

  const monthlySalesData = data?.monthlySales?.map((item) => ({
    name: item.month,
    sales: item.sales,
  })) || [];

  // Анализ продаж и расчет метрик роста
  const totalPeriodSales = data?.monthlySales?.reduce((acc, curr) => acc + curr.sales, 0) || 0;
  const recentMonthSales = data?.monthlySales && data.monthlySales.length > 0
    ? data.monthlySales[0].sales
    : 0;
  const prevMonthSales = data?.monthlySales && data.monthlySales.length > 1
    ? data.monthlySales[1].sales
    : undefined;

  // 1. Месячный рост (MoM) для карточки "Продажи за последний месяц"
  const recentMonthGrowth = prevMonthSales !== undefined
    ? (prevMonthSales > 0
        ? ((recentMonthSales - prevMonthSales) / prevMonthSales) * 100
        : (recentMonthSales > 0 ? 100 : 0))
    : data?.earningsGrowth;

  // 2. Среднемесячная выручка и сравнение текущего месяца со средним
  const activeMonthsCount = data?.monthlySales?.length || 1;
  const averageMonthlySales = totalPeriodSales / activeMonthsCount;
  const vsAverageGrowth = averageMonthlySales > 0
    ? ((recentMonthSales - averageMonthlySales) / averageMonthlySales) * 100
    : 0;

  // 3. Динамика продаж за период (сравнение с предшествующим аналогичным периодом)
  let periodGrowth: number | undefined = undefined;
  if (data?.monthlySales && data.monthlySales.length >= 24) {
    const curYear = data.monthlySales.slice(0, 12).reduce((a, b) => a + b.sales, 0);
    const prevYear = data.monthlySales.slice(12, 24).reduce((a, b) => a + b.sales, 0);
    if (prevYear > 0) {
      periodGrowth = ((curYear - prevYear) / prevYear) * 100;
    }
  } else if (data?.monthlySales && data.monthlySales.length >= 2) {
    const half = Math.floor(data.monthlySales.length / 2);
    const recentHalf = data.monthlySales.slice(0, half).reduce((a, b) => a + b.sales, 0) / half;
    const olderHalf = data.monthlySales.slice(half).reduce((a, b) => a + b.sales, 0) / (data.monthlySales.length - half);
    if (olderHalf > 0) {
      periodGrowth = ((recentHalf - olderHalf) / olderHalf) * 100;
    }
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-6 space-y-6">
      {/* Шапка страницы */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">Аналитика и статистика</h1>
          <p className="text-muted-foreground text-sm">
            Объективные финансовые показатели продаж и текущее состояние складских запасов
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => refetch()}
          disabled={isFetching}
          className="self-start sm:self-auto gap-2"
        >
          <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {isFetching ? 'Обновление...' : 'Обновить'}
        </Button>
      </div>

      {data ? (
        <Tabs value={activeTab} onValueChange={(val) => setActiveTab(val as 'sales' | 'inventory')} className="space-y-6">
          <TabsList className="grid w-full sm:w-auto grid-cols-2">
            <TabsTrigger value="sales" className="gap-2">
              <ChartNoAxesCombined className="h-4 w-4" />
              Продажи и выручка
            </TabsTrigger>
            <TabsTrigger value="inventory" className="gap-2">
              <Warehouse className="h-4 w-4" />
              Склад и запасы
            </TabsTrigger>
          </TabsList>

          {/* ВКЛАДКА 1: ПРОДАЖИ И ВЫРУЧКА */}
          <TabsContent value="sales" className="space-y-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              <Card className="relative overflow-hidden">
                <CardHeader className="pb-2">
                  <GrowthIndicator growth={data.earningsGrowth} />
                  <div className="p-2 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 rounded-lg w-fit">
                    <DollarSign className="h-5 w-5" />
                  </div>
                  <CardTitle className="text-sm font-medium text-muted-foreground mt-2">Общая выручка</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl sm:text-3xl font-bold">₽{formatPrice(data.totalEarnings)}</div>
                  <p className="text-xs text-muted-foreground mt-1">Все завершённые заказы</p>
                </CardContent>
              </Card>

              <Card className="relative overflow-hidden">
                <CardHeader className="pb-2">
                  <GrowthIndicator growth={recentMonthGrowth} />
                  <div className="p-2 bg-blue-500/10 text-blue-600 dark:text-blue-400 rounded-lg w-fit">
                    <Calendar className="h-5 w-5" />
                  </div>
                  <CardTitle className="text-sm font-medium text-muted-foreground mt-2">Продажи за последний месяц</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl sm:text-3xl font-bold">₽{formatPrice(recentMonthSales)}</div>
                  <p className="text-xs text-muted-foreground mt-1">
                    {data.monthlySales.length > 0 ? data.monthlySales[0].month : 'Текущий период'} (динамика MoM)
                  </p>
                </CardContent>
              </Card>

              <Card className="relative overflow-hidden">
                <CardHeader className="pb-2">
                  <GrowthIndicator growth={periodGrowth} />
                  <div className="p-2 bg-violet-500/10 text-violet-600 dark:text-violet-400 rounded-lg w-fit">
                    <ChartNoAxesCombined className="h-5 w-5" />
                  </div>
                  <CardTitle className="text-sm font-medium text-muted-foreground mt-2">Продажи за год</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl sm:text-3xl font-bold">₽{formatPrice(totalPeriodSales)}</div>
                  <p className="text-xs text-muted-foreground mt-1">Агрегировано за период</p>
                </CardContent>
              </Card>

              <Card className="relative overflow-hidden">
                <CardHeader className="pb-2">
                  <GrowthIndicator growth={vsAverageGrowth} />
                  <div className="p-2 bg-amber-500/10 text-amber-600 dark:text-amber-400 rounded-lg w-fit">
                    <Receipt className="h-5 w-5" />
                  </div>
                  <CardTitle className="text-sm font-medium text-muted-foreground mt-2">Среднемесячная выручка</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl sm:text-3xl font-bold">₽{formatPrice(averageMonthlySales)}</div>
                  <p className="text-xs text-muted-foreground mt-1">В среднем за {data.monthlySales.length} мес.</p>
                </CardContent>
              </Card>
            </div>

            {/* График динамики продаж: столбчатая диаграмма точно как в Склад и запасы */}
            <Card>
              <CardHeader>
                <CardTitle>Динамика реализации по месяцам</CardTitle>
                <CardDescription>
                  Суммарная стоимость закрытых и оплаченных заказов по данным Orders Service
                </CardDescription>
              </CardHeader>
              <CardContent>
                {monthlySalesData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={320}>
                    <BarChart data={[...monthlySalesData].reverse()}>
                      <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
                      <XAxis dataKey="name" stroke="var(--border)" tick={axisTickStyle} />
                      <YAxis
                        stroke="var(--border)"
                        tick={axisTickStyle}
                        tickFormatter={(value) => `₽${(value / 1000).toFixed(0)}k`}
                      />
                      <Tooltip
                        contentStyle={tooltipContentStyle}
                        itemStyle={tooltipItemStyle}
                        formatter={(value) => [`₽${formatPrice(Number(value))}`, 'Выручка']}
                      />
                      <Bar
                        dataKey="sales"
                        fill="#10b981"
                        radius={[6, 6, 0, 0]}
                        maxBarSize={60}
                      />
                    </BarChart>
                  </ResponsiveContainer>
                ) : (
                  <div className="py-12 text-center text-muted-foreground text-sm">
                    Нет данных о завершённых продажах за выбранный период
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Сетка помесячной детализации */}
            {data.monthlySales.length > 0 && (
              <div className="space-y-3">
                <h3 className="text-lg font-semibold tracking-tight">Помесячная детализация выручки</h3>
                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
                  {data.monthlySales.map((item, index) => (
                    <Card key={index} className="p-3">
                      <div className="text-xs font-medium text-muted-foreground">{item.month}</div>
                      <div className="text-lg font-bold text-emerald-600 dark:text-emerald-400 mt-1">
                        ₽{formatPrice(item.sales)}
                      </div>
                    </Card>
                  ))}
                </div>
              </div>
            )}
          </TabsContent>

          {/* ВКЛАДКА 2: СКЛАД И ЗАПАСЫ */}
          <TabsContent value="inventory" className="space-y-6">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <Card className="relative overflow-hidden">
                <CardHeader className="pb-2">
                  <GrowthIndicator growth={data.partsGrowth} />
                  <div className="p-2 bg-slate-500/10 text-slate-700 dark:text-slate-300 rounded-lg w-fit">
                    <Package className="h-5 w-5" />
                  </div>
                  <CardTitle className="text-sm font-medium text-muted-foreground mt-2">Уникальных наименований</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl sm:text-3xl font-bold">{data.totalParts.toLocaleString('ru-RU')}</div>
                  <p className="text-xs text-muted-foreground mt-1">Позиций в наличии на складе</p>
                </CardContent>
              </Card>

              <Card className="relative overflow-hidden">
                <CardHeader className="pb-2">
                  <GrowthIndicator growth={data.quantityGrowth} />
                  <div className="p-2 bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 rounded-lg w-fit">
                    <Boxes className="h-5 w-5" />
                  </div>
                  <CardTitle className="text-sm font-medium text-muted-foreground mt-2">Общее количество штук</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl sm:text-3xl font-bold">{data.totalQuantity.toLocaleString('ru-RU')} шт.</div>
                  <p className="text-xs text-muted-foreground mt-1">Физический объем складских единиц</p>
                </CardContent>
              </Card>

              <Card className="relative overflow-hidden">
                <CardHeader className="pb-2">
                  <GrowthIndicator growth={data.valueGrowth} />
                  <div className="p-2 bg-amber-500/10 text-amber-600 dark:text-amber-400 rounded-lg w-fit">
                    <DollarSign className="h-5 w-5" />
                  </div>
                  <CardTitle className="text-sm font-medium text-muted-foreground mt-2">Оценочная стоимость склада</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl sm:text-3xl font-bold">₽{formatPrice(data.totalValue)}</div>
                  <p className="text-xs text-muted-foreground mt-1">Оборотный капитал в наличии</p>
                </CardContent>
              </Card>
            </div>

            {/* Распределение по категориям: Гистограмма и Круговая диаграмма */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle>Топ категорий по количеству</CardTitle>
                  <CardDescription>Количество запчастей в основных товарных категориях</CardDescription>
                </CardHeader>
                <CardContent>
                  {categoryData.length > 0 ? (
                    <ResponsiveContainer width="100%" height={300}>
                      <BarChart data={categoryData.slice(0, 8)}>
                        <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
                        <XAxis dataKey="name" stroke="var(--border)" tick={axisTickStyle} />
                        <YAxis stroke="var(--border)" tick={axisTickStyle} />
                        <Tooltip contentStyle={tooltipContentStyle} itemStyle={tooltipItemStyle} />
                        <Bar dataKey="count" fill="var(--primary)" radius={[4, 4, 0, 0]} />
                      </BarChart>
                    </ResponsiveContainer>
                  ) : (
                    <div className="py-12 text-center text-muted-foreground text-sm">Нет данных по категориям</div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Доли категорий в инвентаре</CardTitle>
                  <CardDescription>Процентное соотношение категорий на складе</CardDescription>
                </CardHeader>
                <CardContent>
                  {categoryData.length > 0 ? (
                    <ResponsiveContainer width="100%" height={300}>
                      <PieChart>
                        <Pie
                          data={categoryData.slice(0, 6)}
                          cx="50%"
                          cy="50%"
                          labelLine={{ stroke: 'var(--muted-foreground)' }}
                          label={({ name, percent }) => `${name} ${((percent as number) * 100).toFixed(0)}%`}
                          outerRadius={90}
                          dataKey="count"
                        >
                          {categoryData.slice(0, 6).map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={entry.fill} />
                          ))}
                        </Pie>
                        <Tooltip contentStyle={tooltipContentStyle} itemStyle={tooltipItemStyle} />
                      </PieChart>
                    </ResponsiveContainer>
                  ) : (
                    <div className="py-12 text-center text-muted-foreground text-sm">Нет данных по категориям</div>
                  )}
                </CardContent>
              </Card>
            </div>

            {/* Список всех категорий */}
            {data.categories.length > 0 && (
              <Card>
                <CardHeader>
                  <CardTitle>Все категории на складе</CardTitle>
                  <CardDescription>Полный перечень учтенных групп деталей</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
                    {data.categories.map((cat, index) => (
                      <div key={index} className="p-3 rounded-lg border bg-card">
                        <div className="text-xs text-muted-foreground truncate" title={cat.name}>{cat.name}</div>
                        <div className="text-lg font-bold mt-1">{cat.count.toLocaleString('ru-RU')}</div>
                        <div className="text-[10px] text-muted-foreground">запчастей</div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}
          </TabsContent>
        </Tabs>
      ) : (
        <Card className="text-center py-12">
          <CardContent>
            <p className="text-muted-foreground text-base">Данные статистики временно недоступны.</p>
            <Button variant="outline" size="sm" onClick={() => refetch()} className="mt-4">
              Попробовать снова
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

export default Statistics;
