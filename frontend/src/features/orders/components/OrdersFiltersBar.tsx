/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Search } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { PAYMENT_OPTIONS, WAREHOUSE_OPTIONS } from '../constants';

interface OrdersFiltersBarProps {
  searchQuery: string;
  onSearchChange: (value: string) => void;
  activeTab: 'active' | 'completed' | 'customers';
  paymentFilter: string;
  onPaymentFilterChange: (value: string) => void;
  warehouseFilter: string;
  onWarehouseFilterChange: (value: string) => void;
  locationFilter: string;
  onLocationFilterChange: (value: string) => void;
  uniqueLocations: string[];
}

export const OrdersFiltersBar: React.FC<OrdersFiltersBarProps> = ({
  searchQuery,
  onSearchChange,
  activeTab,
  paymentFilter,
  onPaymentFilterChange,
  warehouseFilter,
  onWarehouseFilterChange,
  locationFilter,
  onLocationFilterChange,
  uniqueLocations,
}) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5">
      <div className="relative">
        <Search className="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder="Поиск: телефон, № сделки, деталь, трек..."
          className="pl-9 h-9 text-sm"
        />
      </div>

      {activeTab === 'active' && (
        <>
          <Select value={paymentFilter} onValueChange={onPaymentFilterChange}>
            <SelectTrigger className="h-9 text-xs sm:text-sm">
              <SelectValue placeholder="Статус оплаты" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">💰 Все статусы оплаты</SelectItem>
              {PAYMENT_OPTIONS.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select value={warehouseFilter} onValueChange={onWarehouseFilterChange}>
            <SelectTrigger className="h-9 text-xs sm:text-sm">
              <SelectValue placeholder="Этап склада" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">🏭 Все этапы склада</SelectItem>
              {WAREHOUSE_OPTIONS.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select value={locationFilter} onValueChange={onLocationFilterChange}>
            <SelectTrigger className="h-9 text-xs sm:text-sm">
              <SelectValue placeholder="Склад" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">📍 Все склады</SelectItem>
              {uniqueLocations.map((loc) => (
                <SelectItem key={loc} value={loc}>
                  {loc}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </>
      )}
    </div>
  );
};
