/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { SearchableSelect } from '@/components/ui/searchable-select';
import type { SelectOption } from '@/components/ui/searchable-select';
import { conditionOptions as defaultConditionOptions } from './searchTypes';
import type { SearchFiltersState } from './searchTypes';

export interface GeneralFiltersTabProps {
  filters: SearchFiltersState;
  onChange: <K extends keyof SearchFiltersState>(key: K, val: SearchFiltersState[K]) => void;
  brandOptions: SelectOption[];
  modelOptions: SelectOption[];
  categoryOptions: SelectOption[];
  statusOptions: SelectOption[];
  photoOptions: SelectOption[];
  conditionOptions?: SelectOption[];
}

export const GeneralFiltersTab: React.FC<GeneralFiltersTabProps> = ({
  filters,
  onChange,
  brandOptions,
  modelOptions,
  categoryOptions,
  statusOptions,
  photoOptions,
  conditionOptions = defaultConditionOptions,
}) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {/* Бренд */}
      <div>
        <Label htmlFor="brand-filter">Бренд</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.brand}
            onValueChange={(value) => {
              onChange('brand', value);
              onChange('model', '');
            }}
            options={brandOptions}
            placeholder="Все бренды"
            searchPlaceholder="Поиск бренда..."
            emptyMessage="Бренд не найден"
            allowCustom={true}
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Модель */}
      <div>
        <Label htmlFor="model-filter">Модель</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.model}
            onValueChange={(val) => onChange('model', val)}
            options={modelOptions}
            placeholder={filters.brand ? 'Все модели' : 'Сначала выберите бренд'}
            searchPlaceholder="Поиск модели..."
            emptyMessage="Модель не найдена — можно ввести свою"
            allowCustom={true}
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Категория */}
      <div>
        <Label htmlFor="category-filter">Категория</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.category}
            onValueChange={(val) => onChange('category', val)}
            options={categoryOptions}
            placeholder="Все категории"
            searchPlaceholder="Поиск категории..."
            emptyMessage="Категория не найдена"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Год выпуска */}
      <div>
        <Label htmlFor="release-date-filter">Год выпуска</Label>
        <Input
          id="release-date-filter"
          type="text"
          value={filters.carReleaseDate}
          onChange={(e) => onChange('carReleaseDate', e.target.value)}
          placeholder="2010..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Состояние */}
      <div>
        <Label htmlFor="condition-filter">Состояние</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.condition}
            onValueChange={(val) => onChange('condition', val)}
            options={conditionOptions}
            placeholder="Все состояния"
            searchPlaceholder="Поиск состояния..."
            emptyMessage="Состояние не найдено"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Сдвоенный диапазон: Цена */}
      <div>
        <Label>Цена (₽)</Label>
        <div className="grid grid-cols-2 gap-2 mt-1.5">
          <Input
            id="min-price-filter"
            type="number"
            value={filters.minPrice}
            onChange={(e) => onChange('minPrice', e.target.value)}
            placeholder="От 0"
            min="0"
            className="bg-transparent border border-gray-300"
          />
          <Input
            id="max-price-filter"
            type="number"
            value={filters.maxPrice}
            onChange={(e) => onChange('maxPrice', e.target.value)}
            placeholder="До"
            min="0"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Сдвоенный диапазон: Количество */}
      <div>
        <Label>Количество</Label>
        <div className="grid grid-cols-2 gap-2 mt-1.5">
          <Input
            id="min-quantity-filter"
            type="number"
            value={filters.minQuantity}
            onChange={(e) => onChange('minQuantity', e.target.value)}
            placeholder="От"
            min="0"
            className="bg-transparent border border-gray-300"
          />
          <Input
            id="max-quantity-filter"
            type="number"
            value={filters.maxQuantity}
            onChange={(e) => onChange('maxQuantity', e.target.value)}
            placeholder="До"
            min="0"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Статус */}
      <div>
        <Label htmlFor="status-filter">Статус</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.status}
            onValueChange={(val) => onChange('status', val)}
            options={statusOptions}
            placeholder="Все статусы"
            searchPlaceholder="Поиск статуса..."
            emptyMessage="Статус не найден"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>

      {/* Фото */}
      <div>
        <Label htmlFor="photo-filter">Фото</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={filters.hasPhoto}
            onValueChange={(val) => onChange('hasPhoto', val)}
            options={photoOptions}
            placeholder="Все"
            searchPlaceholder="Поиск по фото..."
            emptyMessage="Опция не найдена"
            className="bg-transparent border border-gray-300"
          />
        </div>
      </div>
    </div>
  );
};
