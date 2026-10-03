/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { SearchableSelect } from '@/components/ui/searchable-select';
import type { SelectOption } from '@/components/ui/searchable-select';

interface GeneralFiltersTabProps {
  brand: string;
  setBrand: (val: string) => void;
  model: string;
  setModel: (val: string) => void;
  category: string;
  setCategory: (val: string) => void;
  carReleaseDate: string;
  setCarReleaseDate: (val: string) => void;
  condition: string;
  setCondition: (val: string) => void;
  minPrice: string;
  setMinPrice: (val: string) => void;
  maxPrice: string;
  setMaxPrice: (val: string) => void;
  minQuantity: string;
  setMinQuantity: (val: string) => void;
  maxQuantity: string;
  setMaxQuantity: (val: string) => void;
  status: string;
  setStatus: (val: string) => void;
  hasPhoto: string;
  setHasPhoto: (val: string) => void;
  brandOptions: SelectOption[];
  modelOptions: SelectOption[];
  categoryOptions: SelectOption[];
  statusOptions: SelectOption[];
  photoOptions: SelectOption[];
}

export const GeneralFiltersTab: React.FC<GeneralFiltersTabProps> = ({
  brand,
  setBrand,
  model,
  setModel,
  category,
  setCategory,
  carReleaseDate,
  setCarReleaseDate,
  condition,
  setCondition,
  minPrice,
  setMinPrice,
  maxPrice,
  setMaxPrice,
  minQuantity,
  setMinQuantity,
  maxQuantity,
  setMaxQuantity,
  status,
  setStatus,
  hasPhoto,
  setHasPhoto,
  brandOptions,
  modelOptions,
  categoryOptions,
  statusOptions,
  photoOptions,
}) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {/* Бренд */}
      <div>
        <Label htmlFor="brand-filter">Бренд</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={brand}
            onValueChange={(value) => {
              setBrand(value);
              setModel('');
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
            value={model}
            onValueChange={setModel}
            options={modelOptions}
            placeholder={brand ? 'Все модели' : 'Сначала выберите бренд'}
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
            value={category}
            onValueChange={setCategory}
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
          value={carReleaseDate}
          onChange={(e) => setCarReleaseDate(e.target.value)}
          placeholder="2010..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Состояние */}
      <div>
        <Label htmlFor="condition-filter">Состояние</Label>
        <Input
          id="condition-filter"
          type="text"
          value={condition}
          onChange={(e) => setCondition(e.target.value)}
          placeholder="Контрактная, б/у, новая..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Сдвоенный диапазон: Цена */}
      <div>
        <Label>Цена (₽)</Label>
        <div className="grid grid-cols-2 gap-2 mt-1.5">
          <Input
            id="min-price-filter"
            type="number"
            value={minPrice}
            onChange={(e) => setMinPrice(e.target.value)}
            placeholder="От 0"
            min="0"
            className="bg-transparent border border-gray-300"
          />
          <Input
            id="max-price-filter"
            type="number"
            value={maxPrice}
            onChange={(e) => setMaxPrice(e.target.value)}
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
            value={minQuantity}
            onChange={(e) => setMinQuantity(e.target.value)}
            placeholder="От"
            min="0"
            className="bg-transparent border border-gray-300"
          />
          <Input
            id="max-quantity-filter"
            type="number"
            value={maxQuantity}
            onChange={(e) => setMaxQuantity(e.target.value)}
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
            value={status}
            onValueChange={setStatus}
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
            value={hasPhoto}
            onValueChange={setHasPhoto}
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
