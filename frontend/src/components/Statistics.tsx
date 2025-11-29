/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState, useEffect, useCallback } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Area, AreaChart } from 'recharts';
import { Skeleton } from "@/components/ui/skeleton";

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
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="bg-gray-50 border border-gray-200 p-6 rounded-lg shadow-sm">
              <Skeleton className="h-6 w-32 mb-2" />
              <Skeleton className="h-8 w-16" />
            </div>
            <div className="bg-gray-50 border border-gray-200 p-6 rounded-lg shadow-sm">
              <Skeleton className="h-6 w-40 mb-2" />
              <Skeleton className="h-8 w-24" />
            </div>
          </div>

          {/* Charts skeleton */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <div className="bg-white border border-gray-200 p-6 rounded-lg shadow-sm">
              <Skeleton className="h-6 w-48 mb-4" />
              <Skeleton className="h-64 w-full" />
            </div>
            <div className="bg-white border border-gray-200 p-6 rounded-lg shadow-sm">
              <Skeleton className="h-6 w-56 mb-4" />
              <Skeleton className="h-64 w-full" />
            </div>
          </div>

          {/* Area chart skeleton */}
          <div className="bg-white border border-gray-200 p-6 rounded-lg shadow-sm">
            <Skeleton className="h-6 w-40 mb-4" />
            <Skeleton className="h-48 w-full" />
          </div>

          {/* Categories list skeleton */}
          <div className="bg-white border border-gray-200 p-6 rounded-lg shadow-sm">
            <Skeleton className="h-6 w-64 mb-4" />
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {Array.from({ length: 6 }).map((_, index) => (
                <div key={index} className="bg-gray-50 border border-gray-200 p-4 rounded-lg">
                  <Skeleton className="h-5 w-20 mb-2" />
                  <Skeleton className="h-7 w-12 mb-1" />
                  <Skeleton className="h-4 w-16" />
                </div>
              ))}
            </div>
          </div>
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
    <div className="max-w-7xl mx-auto px-6 py-4">
      <h2 className="text-3xl font-bold mb-2">Статистика продаж</h2>
      <p className="text-gray-600 mb-8">Просматривайте аналитику продаж по автоматически удалённым и завершённым заказам</p>

      {data ? (
        <div className="space-y-8">
          {/* Summary Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="bg-gray-50 border border-gray-200 text-gray-900 p-6 rounded-lg shadow-sm">
              <h3 className="text-lg font-semibold mb-2">Всего запчастей</h3>
              <p className="text-3xl font-bold text-gray-700">{data.total_parts}</p>
            </div>
            <div className="bg-gray-50 border border-gray-200 text-gray-900 p-6 rounded-lg shadow-sm">
              <h3 className="text-lg font-semibold mb-2">Общая стоимость</h3>
              <p className="text-3xl font-bold text-gray-700">₽{formatPrice(data.total_value)}</p>
            </div>
            <div className="bg-green-50 border border-green-200 text-gray-900 p-6 rounded-lg shadow-sm">
              <h3 className="text-lg font-semibold mb-2">Общий заработок</h3>
              <p className="text-3xl font-bold text-green-700">₽{formatPrice(data.total_earnings)}</p>
            </div>
          </div>

          {/* Charts */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            {/* Bar Chart for Categories */}
            <div className="bg-white border border-gray-200 p-6 rounded-lg shadow-sm">
              <h3 className="text-xl font-semibold mb-4 text-gray-800">Распределение по категориям</h3>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={categoryData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="name" stroke="#6b7280" />
                  <YAxis stroke="#6b7280" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#f9fafb',
                      border: '1px solid #e5e7eb',
                      borderRadius: '6px'
                    }}
                  />
                  <Bar dataKey="count" fill="#6b7280" />
                </BarChart>
              </ResponsiveContainer>
            </div>

            {/* Pie Chart for Categories */}
            <div className="bg-white border border-gray-200 p-6 rounded-lg shadow-sm">
              <h3 className="text-xl font-semibold mb-4 text-gray-800">Категории (Круговая диаграмма)</h3>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={categoryData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percent }) => `${name} ${((percent as number) * 100).toFixed(0)}%`}
                    outerRadius={80}
                    fill="#6b7280"
                    dataKey="count"
                  >
                    {categoryData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.fill} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#f9fafb',
                      border: '1px solid #e5e7eb',
                      borderRadius: '6px'
                    }}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>

          {/* Monthly Sales Chart */}
          <div className="bg-white border border-gray-200 p-6 rounded-lg shadow-sm">
            <h3 className="text-xl font-semibold mb-4 text-gray-800">Продажи за год (статистика по месяцам)</h3>
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={monthlySalesData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="name" stroke="#6b7280" />
                <YAxis stroke="#6b7280" />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#f9fafb',
                    border: '1px solid #e5e7eb',
                    borderRadius: '6px'
                  }}
                  formatter={(value) => [`₽${formatPrice(Number(value))}`, 'Продажи']}
                />
                <Area type="monotone" dataKey="sales" stroke="#059669" fill="#059669" fillOpacity={0.6} />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          {/* Area Chart for Summary */}
          <div className="bg-white border border-gray-200 p-6 rounded-lg shadow-sm">
            <h3 className="text-xl font-semibold mb-4 text-gray-800">Обзор показателей</h3>
            <ResponsiveContainer width="100%" height={200}>
              <AreaChart data={summaryData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="name" stroke="#6b7280" />
                <YAxis stroke="#6b7280" />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#f9fafb',
                    border: '1px solid #e5e7eb',
                    borderRadius: '6px'
                  }}
                />
                <Area type="monotone" dataKey="value" stroke="#6b7280" fill="#6b7280" fillOpacity={0.1} />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          {/* Detailed Category List */}
          <div className="bg-white border border-gray-200 p-6 rounded-lg shadow-sm">
            <h3 className="text-xl font-semibold mb-4 text-gray-800">Детальная информация по категориям</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {data.categories && Array.isArray(data.categories) && data.categories.filter(cat => cat && cat.name).map((cat, index) => (
                <div key={index} className="bg-gray-50 border border-gray-200 p-4 rounded-lg">
                  <h4 className="font-semibold text-gray-800">{cat.name}</h4>
                  <p className="text-2xl font-bold text-gray-700">{cat.count}</p>
                  <p className="text-sm text-gray-600">запчастей</p>
                </div>
              ))}
            </div>
          </div>

          {/* Monthly Sales Details */}
          <div className="bg-white border border-gray-200 p-6 rounded-lg shadow-sm">
            <h3 className="text-xl font-semibold mb-4 text-gray-800">Детальная информация по продажам за год</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {data.monthly_sales && Array.isArray(data.monthly_sales) && data.monthly_sales.map((item, index) => (
                <div key={index} className="bg-green-50 border border-green-200 p-4 rounded-lg">
                  <h4 className="font-semibold text-gray-800">{item.month}</h4>
                  <p className="text-2xl font-bold text-green-700">₽{formatPrice(item.sales)}</p>
                  <p className="text-sm text-gray-600">продаж</p>
                </div>
              ))}
            </div>
          </div>
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
