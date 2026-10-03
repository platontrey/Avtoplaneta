/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';

interface WarehouseFiltersTabProps {
  location: string;
  setLocation: (val: string) => void;
  address: string;
  setAddress: (val: string) => void;
  salesman: string;
  setSalesman: (val: string) => void;
  manufacturer: string;
  setManufacturer: (val: string) => void;
  manufacturerCode: string;
  setManufacturerCode: (val: string) => void;
  supplierCode: string;
  setSupplierCode: (val: string) => void;
  wearPercentage: string;
  setWearPercentage: (val: string) => void;
}

export const WarehouseFiltersTab: React.FC<WarehouseFiltersTabProps> = ({
  location,
  setLocation,
  address,
  setAddress,
  salesman,
  setSalesman,
  manufacturer,
  setManufacturer,
  manufacturerCode,
  setManufacturerCode,
  supplierCode,
  setSupplierCode,
  wearPercentage,
  setWearPercentage,
}) => {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {/* Местоположение */}
      <div>
        <Label htmlFor="location-filter">Местоположение (полка/стеллаж)</Label>
        <Input
          id="location-filter"
          type="text"
          value={location}
          onChange={(e) => setLocation(e.target.value)}
          placeholder="Shelf A-12..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Адрес склада */}
      <div>
        <Label htmlFor="address-filter">Адрес склада</Label>
        <Input
          id="address-filter"
          type="text"
          value={address}
          onChange={(e) => setAddress(e.target.value)}
          placeholder="Профсоюзная 2/11..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>

      {/* Продавец */}
      <div>
        <Label htmlFor="salesman-filter">Продавец</Label>
        <Input
          id="salesman-filter"
          type="text"
          value={salesman}
          onChange={(e) => setSalesman(e.target.value)}
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
          value={manufacturer}
          onChange={(e) => setManufacturer(e.target.value)}
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
          value={manufacturerCode}
          onChange={(e) => setManufacturerCode(e.target.value)}
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
          value={supplierCode}
          onChange={(e) => setSupplierCode(e.target.value)}
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
          value={wearPercentage}
          onChange={(e) => setWearPercentage(e.target.value)}
          placeholder="5%, 10%..."
          className="mt-1.5 bg-transparent border border-gray-300"
        />
      </div>
    </div>
  );
};
