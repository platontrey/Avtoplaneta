/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { SearchableSelect } from '@/components/ui/searchable-select';
import type { SelectOption } from '@/components/ui/searchable-select';
import type { SearchFiltersState } from './searchTypes';

export interface TransmissionFiltersTabProps {
  filters: SearchFiltersState;
  onChange: <K extends keyof SearchFiltersState>(key: K, val: SearchFiltersState[K]) => void;
  transmissionOptions: SelectOption[];
  driveOptions: SelectOption[];
  frontRearOptions: SelectOption[];
  leftRightOptions: SelectOption[];
  topBottomOptions: SelectOption[];
}

export const TransmissionFiltersTab: React.FC<TransmissionFiltersTabProps> = ({
  filters,
  onChange,
  transmissionOptions,
  driveOptions,
  frontRearOptions,
  leftRightOptions,
  topBottomOptions,
}) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {/* Трансмиссия */}
      <div>
        <Label htmlFor="transmission-filter">Трансмиссия</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.transmission}
            onValueChange={(val) => onChange('transmission', val)}
            options={transmissionOptions}
            placeholder="Все типы КПП"
            searchPlaceholder="Поиск КПП..."
            emptyMessage="Тип КПП не найден"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Модель КПП */}
      <div>
        <Label htmlFor="transmission-model-filter">Модель КПП</Label>
        <Input
          id="transmission-model-filter"
          type="text"
          value={filters.transmissionModel}
          onChange={(e) => onChange('transmissionModel', e.target.value)}
          placeholder="U140F, RE4F04B..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Привод */}
      <div>
        <Label htmlFor="drive-filter">Привод</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.drive}
            onValueChange={(val) => onChange('drive', val)}
            options={driveOptions}
            placeholder="Все приводы"
            searchPlaceholder="Поиск привода..."
            emptyMessage="Привод не найден"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Перед / зад */}
      <div>
        <Label htmlFor="front-rear-filter">Перед / зад</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.frontRear}
            onValueChange={(val) => onChange('frontRear', val)}
            options={frontRearOptions}
            placeholder="Все расположения"
            searchPlaceholder="Поиск..."
            emptyMessage="Не найдено"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Право / лево */}
      <div>
        <Label htmlFor="left-right-filter">Право / лево</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.leftRight}
            onValueChange={(val) => onChange('leftRight', val)}
            options={leftRightOptions}
            placeholder="Все стороны"
            searchPlaceholder="Поиск..."
            emptyMessage="Не найдено"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Верх / низ */}
      <div>
        <Label htmlFor="top-bottom-filter">Верх / низ</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.topBottom}
            onValueChange={(val) => onChange('topBottom', val)}
            options={topBottomOptions}
            placeholder="Все положения"
            searchPlaceholder="Поиск..."
            emptyMessage="Не найдено"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>
    </div>
  );
};
