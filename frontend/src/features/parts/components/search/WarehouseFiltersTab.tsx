/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { SelectItem } from '@/components/ui/select';
import { ClearableSelect } from '@/components/ClearableSelect';
import { SearchableSelect } from '@/components/ui/searchable-select';
import { useShelves } from '@/hooks/useShelves';
import type { SearchFiltersState } from './searchTypes';

const WAREHOUSE_ADDRESSES = [
  'Томск, Профсоюзная 2/11',
  'г. Томск, пос. Кузовлево, ул. Заречная, 36',
];

export interface WarehouseFiltersTabProps {
  filters: SearchFiltersState;
  onChange: <K extends keyof SearchFiltersState>(key: K, val: SearchFiltersState[K]) => void;
}

export const WarehouseFiltersTab: React.FC<WarehouseFiltersTabProps> = ({
  filters,
  onChange,
}) => {
  const { shelves } = useShelves();

  const shelfOptions = shelves.map((s) => ({ value: s.name, label: s.name }));
  if (filters.location && !shelfOptions.some((o) => o.value.toLowerCase() === filters.location.toLowerCase())) {
    shelfOptions.unshift({ value: filters.location, label: filters.location });
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {/* Местоположение */}
      <div>
        <Label htmlFor="location-filter">Местоположение (полка/стеллаж)</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.location}
            onValueChange={(val) => onChange('location', val)}
            options={shelfOptions}
            placeholder="Все полки"
            searchPlaceholder="Поиск или ввод полки..."
            allowCustom={true}
            className="w-full bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Адрес склада */}
      <div>
        <Label htmlFor="address-filter">Адрес склада</Label>
        <div className="mt-1.5">
          <ClearableSelect
            id="address-filter"
            value={filters.address}
            onValueChange={(val) => onChange('address', val)}
            placeholder="Все склады"
            className="w-full bg-transparent border border-gray-300"
          >
            {WAREHOUSE_ADDRESSES.map((addr) => (
              <SelectItem key={addr} value={addr}>
                {addr}
              </SelectItem>
            ))}
          </ClearableSelect>
        </div>
      </div>

      {/* Продавец */}
      <div>
        <Label htmlFor="salesman-filter">Продавец</Label>
        <Input
          id="salesman-filter"
          type="text"
          value={filters.salesman}
          onChange={(e) => onChange('salesman', e.target.value)}
          placeholder="Имя продавца..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Производитель */}
      <div>
        <Label htmlFor="manufacturer-filter">Производитель</Label>
        <Input
          id="manufacturer-filter"
          type="text"
          value={filters.manufacturer}
          onChange={(e) => onChange('manufacturer', e.target.value)}
          placeholder="Denso, Bosch, Brembo..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Код производителя */}
      <div>
        <Label htmlFor="manufacturer-code-filter">Код производителя</Label>
        <Input
          id="manufacturer-code-filter"
          type="text"
          value={filters.manufacturerCode}
          onChange={(e) => onChange('manufacturerCode', e.target.value)}
          placeholder="Код производителя..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Код поставщика */}
      <div>
        <Label htmlFor="supplier-code-filter">Код поставщика</Label>
        <Input
          id="supplier-code-filter"
          type="text"
          value={filters.supplierCode}
          onChange={(e) => onChange('supplierCode', e.target.value)}
          placeholder="Код поставщика..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Процент износа */}
      <div>
        <Label htmlFor="wear-percentage-filter">Процент износа</Label>
        <Input
          id="wear-percentage-filter"
          type="text"
          value={filters.wearPercentage}
          onChange={(e) => onChange('wearPercentage', e.target.value)}
          placeholder="5%, 10%..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>
    </div>
  );
};
