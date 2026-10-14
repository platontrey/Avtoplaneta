/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { SearchableSelect } from '@/components/ui/searchable-select';
import type { SelectOption } from '@/components/ui/searchable-select';
import { formatCarReleasePeriod } from '@/lib/utils';
import type { SearchFiltersState } from './searchTypes';

export interface BodyEngineFiltersTabProps {
  filters: SearchFiltersState;
  onChange: <K extends keyof SearchFiltersState>(key: K, val: SearchFiltersState[K]) => void;
  bodyOptions: SelectOption[];
  engineOptions: SelectOption[];
}

export const BodyEngineFiltersTab: React.FC<BodyEngineFiltersTabProps> = ({
  filters,
  onChange,
  bodyOptions,
  engineOptions,
}) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {/* Марка кузова */}
      <div>
        <Label htmlFor="body-brand-filter">Марка кузова</Label>
        <SearchableSelect
          value={filters.bodyBrand}
          onValueChange={(val) => onChange('bodyBrand', val)}
          options={bodyOptions}
          placeholder={filters.brand ? 'Выберите кузов' : 'ACV40, E90, W212...'}
          searchPlaceholder="Поиск кузова..."
          emptyMessage="Кузов не найден — можно ввести свой"
          allowCustom={true}
          className="mt-1.5 h-10"
        />
      </div>

      {/* Марка двигателя */}
      <div>
        <Label htmlFor="engine-brand-filter">Марка двигателя</Label>
        <SearchableSelect
          value={filters.engineBrand}
          onValueChange={(val) => onChange('engineBrand', val)}
          options={engineOptions}
          placeholder={filters.brand ? 'Выберите двигатель' : '2AZ-FE, N46, 1JZ...'}
          searchPlaceholder="Поиск двигателя..."
          emptyMessage="Двигатель не найден — можно ввести свой"
          allowCustom={true}
          className="mt-1.5 h-10"
        />
      </div>

      {/* VIN / Номер кузова */}
      <div>
        <Label htmlFor="vin-filter">VIN / Номер кузова</Label>
        <Input
          id="vin-filter"
          type="text"
          value={filters.vin}
          onChange={(e) => onChange('vin', e.target.value)}
          placeholder="WVWZZZ1JZ3W386549..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Период выпуска автомобиля */}
      <div>
        <Label htmlFor="car-release-period-filter">Период выпуска автомобиля</Label>
        <Input
          id="car-release-period-filter"
          type="text"
          value={filters.carReleasePeriod}
          onChange={(e) => onChange('carReleasePeriod', formatCarReleasePeriod(e.target.value))}
          placeholder="2001-2007..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Номер детали */}
      <div>
        <Label htmlFor="number-filter">Номер детали</Label>
        <Input
          id="number-filter"
          type="text"
          value={filters.number}
          onChange={(e) => onChange('number', e.target.value)}
          placeholder="52119-33939..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* OEM код */}
      <div>
        <Label htmlFor="oem-filter">OEM код</Label>
        <Input
          id="oem-filter"
          type="text"
          value={filters.oemCode}
          onChange={(e) => onChange('oemCode', e.target.value)}
          placeholder="89661-06G80..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Дефект */}
      <div>
        <Label htmlFor="defect-filter">Дефект</Label>
        <Input
          id="defect-filter"
          type="text"
          value={filters.defect}
          onChange={(e) => onChange('defect', e.target.value)}
          placeholder="Царапина, скол, трещина..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Цвет */}
      <div>
        <Label htmlFor="color-filter">Цвет</Label>
        <Input
          id="color-filter"
          type="text"
          value={filters.color}
          onChange={(e) => onChange('color', e.target.value)}
          placeholder="Черный, белый, серебристый..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>
    </div>
  );
};
