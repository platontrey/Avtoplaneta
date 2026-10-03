/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { SearchableSelect } from '@/components/ui/searchable-select';
import type { SelectOption } from '@/components/ui/searchable-select';
import { formatCarReleasePeriod } from '@/lib/utils';

interface BodyEngineFiltersTabProps {
  brand: string;
  bodyBrand: string;
  setBodyBrand: (val: string) => void;
  engineBrand: string;
  setEngineBrand: (val: string) => void;
  vin: string;
  setVin: (val: string) => void;
  carReleasePeriod: string;
  setCarReleasePeriod: (val: string) => void;
  number: string;
  setNumber: (val: string) => void;
  oemCode: string;
  setOemCode: (val: string) => void;
  defect: string;
  setDefect: (val: string) => void;
  color: string;
  setColor: (val: string) => void;
  bodyOptions: SelectOption[];
  engineOptions: SelectOption[];
}

export const BodyEngineFiltersTab: React.FC<BodyEngineFiltersTabProps> = ({
  brand,
  bodyBrand,
  setBodyBrand,
  engineBrand,
  setEngineBrand,
  vin,
  setVin,
  carReleasePeriod,
  setCarReleasePeriod,
  number,
  setNumber,
  oemCode,
  setOemCode,
  defect,
  setDefect,
  color,
  setColor,
  bodyOptions,
  engineOptions,
}) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {/* Марка кузова */}
      <div>
        <Label htmlFor="body-brand-filter">Марка кузова</Label>
        <SearchableSelect
          value={bodyBrand}
          onValueChange={setBodyBrand}
          options={bodyOptions}
          placeholder={brand ? 'Выберите кузов' : 'ACV40, E90, W212...'}
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
          value={engineBrand}
          onValueChange={setEngineBrand}
          options={engineOptions}
          placeholder={brand ? 'Выберите двигатель' : '2AZ-FE, N46, 1JZ...'}
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
          value={vin}
          onChange={(e) => setVin(e.target.value)}
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
          value={carReleasePeriod}
          onChange={(e) => setCarReleasePeriod(formatCarReleasePeriod(e.target.value))}
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
          value={number}
          onChange={(e) => setNumber(e.target.value)}
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
          value={oemCode}
          onChange={(e) => setOemCode(e.target.value)}
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
          value={defect}
          onChange={(e) => setDefect(e.target.value)}
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
          value={color}
          onChange={(e) => setColor(e.target.value)}
          placeholder="Черный, белый, серебристый..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>
    </div>
  );
};
