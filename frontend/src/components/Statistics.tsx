/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState, useEffect, useCallback } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Area, AreaChart } from 'recharts';
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface Statistics {
   total_parts: number;
   total_value: number;
   total_earnings: number;
   categories: { name: string; count: number }[];
   monthly_sales: { month: string; sales: number }[];
}

function Statistics() {
  const [data, setData] = useState<Statistics | null>(null);
  const [loading, setLoading] = useState(true);

  // Функция форматирования цены для корректного отображения
  const formatPrice = (price: number) => {
    return price.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  };

  const fetchData = useCallback(async () => {
    try {
      const response = await fetch('http://localhost:8080/api/statistics', {
        credentials: 'include', // Include cookies in the request
      });
      if (!response.ok) {
        console.error(`HTTP error! status: ${response.status}`);
        return;
      }
      const result = await response.json();
      setData(result);
    } catch (error) {
      console.error('Error fetching statistics:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Expose refresh function to window for external calls
  useEffect(() => {
    (window as unknown as { refreshStatistics: () => void }).refreshStatistics = fetchData;
    return () => {
      delete (window as unknown as { refreshStatistics?: () => void }).refreshStatistics;
    };
  }, [fetchData]);

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-6 py-4">
        <div className="space-y-8">
          {/* Header skeleton */}
          <div className="space-y-2">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-4 w-96" />
          </div>

          {/* Summary cards skeleton */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <Card>
              <CardHeader className="pb-2">
                <Skeleton className="h-6 w-32" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-16" />
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <Skeleton className="h-6 w-40" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-24" />
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <Skeleton className="h-6 w-36" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-20" />
              </CardContent>
            </Card>
          </div>

          {/* Charts skeleton */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <Card>
              <CardHeader>
                <Skeleton className="h-6 w-48" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-64 w-full" />
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <Skeleton className="h-6 w-56" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-64 w-full" />
              </CardContent>
            </Card>
          </div>

          {/* Area chart skeleton */}
          <Card>
            <CardHeader>
              <Skeleton className="h-6 w-40" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-48 w-full" />
            </CardContent>
          </Card>

          {/* Categories list skeleton */}
          <Card>
            <CardHeader>
              <Skeleton className="h-6 w-64" />
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {Array.from({ length: 6 }).map((_, index) => (
                  <Card key={index}>
                    <CardContent className="p-4">
                      <Skeleton className="h-5 w-20 mb-2" />
                      <Skeleton className="h-7 w-12 mb-1" />
                      <Skeleton className="h-4 w-16" />
                    </CardContent>
                  </Card>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  // Prepare data for charts
  const categoryData = data?.categories?.filter(cat => cat && cat.name).map((cat, index) => ({
    name: cat.name,
    count: cat.count,
    fill: ['#6b7280', '#9ca3af', '#d1d5db', '#f3f4f6', '#e5e7eb'][index % 5]
  })) || [];

  const monthlySalesData = data?.monthly_sales?.map((item, index) => ({
     name: item.month,
     sales: item.sales,
     fill: ['#059669', '#10b981', '#34d399', '#6ee7b7', '#a7f3d0', '#34d399', '#10b981', '#059669', '#047857', '#065f46', '#064e3b', '#022c22'][index % 12]
  })) || [];

  const summaryData = [
    { name: 'Всего запчастей', value: data?.total_parts || 0, color: '#6b7280' },
    { name: 'Общая стоимость', value: data?.total_value || 0, color: '#374151' },
    { name: 'Общий заработок', value: data?.total_earnings || 0, color: '#059669' }
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4">
      <h2 className="text-2xl sm:text-3xl font-bold mb-2">Статистика продаж</h2>
      <p className="text-gray-600 mb-6 sm:mb-8 text-sm sm:text-base">Просматривайте аналитику продаж по автоматически удалённым и завершённым заказам</p>

      {data ? (
        <div className="space-y-8">
          {/* Summary Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-lg">Всего запчастей</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold">{data.total_parts}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-lg">Общая стоимость</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold">₽{formatPrice(data.total_value)}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-lg">Общий заработок</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-green-600">₽{formatPrice(data.total_earnings)}</p>
              </CardContent>
            </Card>
          </div>

          {/* Charts */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            {/* Bar Chart for Categories */}
            <Card>
              <CardHeader>
                <CardTitle>Распределение по категориям</CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={categoryData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis />
                    <Tooltip />
                    <Bar dataKey="count" />
                  </BarChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            {/* Pie Chart for Categories */}
            <Card>
              <CardHeader>
                <CardTitle>Категории (Круговая диаграмма)</CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <PieChart>
                    <Pie
                      data={categoryData}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      label={({ name, percent }) => `${name} ${((percent as number) * 100).toFixed(0)}%`}
                      outerRadius={80}
                      dataKey="count"
                    >
                      {categoryData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.fill} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>
          </div>

          {/* Monthly Sales Chart */}
          <Card>
            <CardHeader>
              <CardTitle>Продажи за год (статистика по месяцам)</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <AreaChart data={monthlySalesData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip formatter={(value) => [`₽${formatPrice(Number(value))}`, 'Продажи']} />
                  <Area type="monotone" dataKey="sales" stroke="#059669" fill="#059669" fillOpacity={0.6} />
                </AreaChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* Area Chart for Summary */}
          <Card>
            <CardHeader>
              <CardTitle>Обзор показателей</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={200}>
                <AreaChart data={summaryData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Area type="monotone" dataKey="value" stroke="#6b7280" fill="#6b7280" fillOpacity={0.1} />
                </AreaChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* Detailed Category List */}
          <Card>
            <CardHeader>
              <CardTitle>Детальная информация по категориям</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {data.categories && Array.isArray(data.categories) && data.categories.filter(cat => cat && cat.name).map((cat, index) => (
                  <Card key={index}>
                    <CardContent className="p-4">
                      <h4 className="font-semibold">{cat.name}</h4>
                      <p className="text-2xl font-bold">{cat.count}</p>
                      <p className="text-sm text-muted-foreground">запчастей</p>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Monthly Sales Details */}
          <Card>
            <CardHeader>
              <CardTitle>Детальная информация по продажам за год</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {data.monthly_sales && Array.isArray(data.monthly_sales) && data.monthly_sales.map((item, index) => (
                  <Card key={index}>
                    <CardContent className="p-4">
                      <h4 className="font-semibold">{item.month}</h4>
                      <p className="text-2xl font-bold text-green-600">₽{formatPrice(item.sales)}</p>
                      <p className="text-sm text-muted-foreground">продаж</p>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      ) : (
        <div className="text-center py-12">
          <p className="text-gray-500 text-lg">Данные недоступны. Убедитесь, что бэкенд запущен и подключен к базе данных.</p>
        </div>
      )}
    </div>
  );
}

export default Statistics;
