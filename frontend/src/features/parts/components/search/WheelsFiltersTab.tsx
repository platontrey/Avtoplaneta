/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { SearchableSelect } from '@/components/ui/searchable-select';
import type { SelectOption } from '@/components/ui/searchable-select';

interface WheelsFiltersTabProps {
  season: string;
  setSeason: (val: string) => void;
  tireModel: string;
  setTireModel: (val: string) => void;
  diameter: string;
  setDiameter: (val: string) => void;
  width: string;
  setWidth: (val: string) => void;
  profile: string;
  setProfile: (val: string) => void;
  tireQuantity: string;
  setTireQuantity: (val: string) => void;
  drilling: string;
  setDrilling: (val: string) => void;
  offset: string;
  setOffset: (val: string) => void;
  centerHoleDiameter: string;
  setCenterHoleDiameter: (val: string) => void;
  seasonOptions: SelectOption[];
}

export const WheelsFiltersTab: React.FC<WheelsFiltersTabProps> = ({
  season,
  setSeason,
  tireModel,
  setTireModel,
  diameter,
  setDiameter,
  width,
  setWidth,
  profile,
  setProfile,
  tireQuantity,
  setTireQuantity,
  drilling,
  setDrilling,
  offset,
  setOffset,
  centerHoleDiameter,
  setCenterHoleDiameter,
  seasonOptions,
}) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {/* Сезонность */}
      <div>
        <Label htmlFor="season-filter">Сезонность шин</Label>
        <div className="mt-1.5">
          <SearchableSelect
            value={season}
            onValueChange={setSeason}
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
          value={tireModel}
          onChange={(e) => setTireModel(e.target.value)}
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
          value={diameter}
          onChange={(e) => setDiameter(e.target.value)}
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
          value={width}
          onChange={(e) => setWidth(e.target.value)}
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
          value={profile}
          onChange={(e) => setProfile(e.target.value)}
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
          value={tireQuantity}
          onChange={(e) => setTireQuantity(e.target.value)}
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
          value={drilling}
          onChange={(e) => setDrilling(e.target.value)}
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
          value={offset}
          onChange={(e) => setOffset(e.target.value)}
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
          value={centerHoleDiameter}
          onChange={(e) => setCenterHoleDiameter(e.target.value)}
          placeholder="60.1, 67.1..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>
    </div>
  );
};
