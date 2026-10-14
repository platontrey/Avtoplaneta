/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { SearchableSelect } from '@/components/ui/searchable-select';
import type { SelectOption } from '@/components/ui/searchable-select';
import type { SearchFiltersState } from './searchTypes';

export interface WheelsFiltersTabProps {
  filters: SearchFiltersState;
  onChange: <K extends keyof SearchFiltersState>(key: K, val: SearchFiltersState[K]) => void;
  seasonOptions: SelectOption[];
}

export const WheelsFiltersTab: React.FC<WheelsFiltersTabProps> = ({
  filters,
  onChange,
  seasonOptions,
}) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {/* Сезонность */}
      <div>
        <Label htmlFor="season-filter">Сезонность шин</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.season}
            onValueChange={(val) => onChange('season', val)}
            options={seasonOptions}
            placeholder="Все сезоны"
            searchPlaceholder="Поиск сезона..."
            emptyMessage="Сезон не найден"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Модель шины */}
      <div>
        <Label htmlFor="tire-model-filter">Модель шины</Label>
        <Input
          id="tire-model-filter"
          type="text"
          value={filters.tireModel}
          onChange={(e) => onChange('tireModel', e.target.value)}
          placeholder="Hakkapeliitta 8, Ice Cruiser..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Диаметр */}
      <div>
        <Label htmlFor="diameter-filter">Диаметр</Label>
        <Input
          id="diameter-filter"
          type="text"
          value={filters.diameter}
          onChange={(e) => onChange('diameter', e.target.value)}
          placeholder="R15, R16, R17..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Ширина шины */}
      <div>
        <Label htmlFor="width-filter">Ширина шины</Label>
        <Input
          id="width-filter"
          type="text"
          value={filters.width}
          onChange={(e) => onChange('width', e.target.value)}
          placeholder="205, 215, 225..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Профиль */}
      <div>
        <Label htmlFor="profile-filter">Профиль шины</Label>
        <Input
          id="profile-filter"
          type="text"
          value={filters.profile}
          onChange={(e) => onChange('profile', e.target.value)}
          placeholder="55, 60, 65..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Количество шин */}
      <div>
        <Label htmlFor="tire-quantity-filter">Количество шин</Label>
        <Input
          id="tire-quantity-filter"
          type="text"
          value={filters.tireQuantity}
          onChange={(e) => onChange('tireQuantity', e.target.value)}
          placeholder="4 шт, пара..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Сверловка (PCD) */}
      <div>
        <Label htmlFor="drilling-filter">Сверловка (PCD)</Label>
        <Input
          id="drilling-filter"
          type="text"
          value={filters.drilling}
          onChange={(e) => onChange('drilling', e.target.value)}
          placeholder="5x114.3, 4x100..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Вылет (ET) */}
      <div>
        <Label htmlFor="offset-filter">Вылет (ET)</Label>
        <Input
          id="offset-filter"
          type="text"
          value={filters.offset}
          onChange={(e) => onChange('offset', e.target.value)}
          placeholder="ET45, ET38..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Диаметр ЦО (DIA) */}
      <div>
        <Label htmlFor="center-hole-diameter-filter">Диаметр ЦО (DIA)</Label>
        <Input
          id="center-hole-diameter-filter"
          type="text"
          value={filters.centerHoleDiameter}
          onChange={(e) => onChange('centerHoleDiameter', e.target.value)}
          placeholder="60.1, 67.1..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>
    </div>
  );
};
