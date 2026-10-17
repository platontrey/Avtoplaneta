/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { SelectOption } from '@/components/ui/searchable-select';
import type { PartFilters } from '@/hooks/useParts';
import { formatCarReleasePeriod } from '@/lib/utils';

export interface SearchFiltersState {
  searchQuery: string;
  category: string;
  brand: string;
  model: string;
  location: string;
  address: string;
  status: string;
  hasPhoto: string;
  number: string;
  oemCode: string;
  vin: string;
  carReleasePeriod: string;
  bodyBrand: string;
  engineBrand: string;
  carReleaseDate: string;
  transmission: string;
  drive: string;
  condition: string;
  manufacturer: string;
  defect: string;
  color: string;
  salesman: string;
  minPrice: string;
  maxPrice: string;
  minQuantity: string;
  maxQuantity: string;
  frontRear: string;
  leftRight: string;
  topBottom: string;
  manufacturerCode: string;
  supplierCode: string;
  transmissionModel: string;
  wearPercentage: string;
  season: string;
  diameter: string;
  width: string;
  profile: string;
  tireQuantity: string;
  drilling: string;
  offset: string;
  centerHoleDiameter: string;
  tireModel: string;
}

export const initialSearchFiltersState: SearchFiltersState = {
  searchQuery: '',
  category: '',
  brand: '',
  model: '',
  location: '',
  address: '',
  status: '',
  hasPhoto: 'with',
  number: '',
  oemCode: '',
  vin: '',
  carReleasePeriod: '',
  bodyBrand: '',
  engineBrand: '',
  carReleaseDate: '',
  transmission: '',
  drive: '',
  condition: '',
  manufacturer: '',
  defect: '',
  color: '',
  salesman: '',
  minPrice: '',
  maxPrice: '',
  minQuantity: '',
  maxQuantity: '',
  frontRear: '',
  leftRight: '',
  topBottom: '',
  manufacturerCode: '',
  supplierCode: '',
  transmissionModel: '',
  wearPercentage: '',
  season: '',
  diameter: '',
  width: '',
  profile: '',
  tireQuantity: '',
  drilling: '',
  offset: '',
  centerHoleDiameter: '',
  tireModel: '',
};

export const statusOptions: SelectOption[] = [
  { value: 'true', label: 'Доступно' },
  { value: 'false', label: 'Недоступно' },
];

export const photoOptions: SelectOption[] = [
  { value: 'all', label: 'Все' },
  { value: 'with', label: 'С фото' },
  { value: 'without', label: 'Без фото' },
];

export const transmissionOptions: SelectOption[] = [
  { value: 'АКПП', label: 'АКПП' },
  { value: 'МКПП', label: 'МКПП' },
  { value: 'Вариатор', label: 'Вариатор' },
  { value: 'Роботизированная', label: 'Роботизированная' },
];

export const driveOptions: SelectOption[] = [
  { value: 'Передний', label: 'Передний' },
  { value: 'Задний', label: 'Задний' },
  { value: 'Полный', label: 'Полный' },
];

export const frontRearOptions: SelectOption[] = [
  { value: 'F', label: 'F (Перед)' },
  { value: 'R', label: 'R (Зад)' },
];

export const leftRightOptions: SelectOption[] = [
  { value: 'L', label: 'L (Лево)' },
  { value: 'R', label: 'R (Право)' },
];

export const topBottomOptions: SelectOption[] = [
  { value: 'Верх', label: 'Верх' },
  { value: 'Низ', label: 'Низ' },
  { value: 'Середина', label: 'Середина' },
];

export const conditionOptions: SelectOption[] = [
  { value: 'Б/У', label: 'Б/У' },
  { value: 'Новое', label: 'Новое' },
];

export const seasonOptions: SelectOption[] = [
  { value: 'Зимние', label: 'Зимние' },
  { value: 'Летние', label: 'Летние' },
  { value: 'Всесезонные', label: 'Всесезонные' },
];

export interface FilterDescriptor {
  key: keyof SearchFiltersState;
  getLabel: (val: string) => string;
  clearVal?: string;
}

export const FILTER_DESCRIPTORS: FilterDescriptor[] = [
  { key: 'address', getLabel: (v) => `Адрес склада: ${v}` },
  { key: 'bodyBrand', getLabel: (v) => `Марка кузова: ${v}` },
  { key: 'brand', getLabel: (v) => `Марка авто: ${v}` },
  { key: 'carReleaseDate', getLabel: (v) => `Год выпуска: ${v}` },
  { key: 'carReleasePeriod', getLabel: (v) => `Период выпуска: ${formatCarReleasePeriod(v)}` },
  { key: 'category', getLabel: (v) => `Категория: ${v}` },
  { key: 'centerHoleDiameter', getLabel: (v) => `Диаметр ЦО: ${v} мм` },
  { key: 'color', getLabel: (v) => `Цвет: ${v}` },
  { key: 'condition', getLabel: (v) => `Состояние: ${v}` },
  { key: 'defect', getLabel: (v) => `Дефект: ${v}` },
  { key: 'diameter', getLabel: (v) => `Диаметр: R${v}` },
  { key: 'drilling', getLabel: (v) => `Сверловка: ${v}` },
  { key: 'drive', getLabel: (v) => `Привод: ${v}` },
  { key: 'engineBrand', getLabel: (v) => `Марка двигателя: ${v}` },
  { key: 'frontRear', getLabel: (v) => `Расположение: ${v === 'F' ? 'F (Перед)' : v === 'R' ? 'R (Зад)' : v}` },
  { key: 'hasPhoto', getLabel: (v) => `Фото: ${v === 'with' ? 'С фото' : 'Без фото'}`, clearVal: 'all' },
  { key: 'leftRight', getLabel: (v) => `Сторона: ${v === 'L' ? 'L (Лево)' : v === 'R' ? 'R (Право)' : v}` },
  { key: 'location', getLabel: (v) => `Местоположение: ${v}` },
  { key: 'manufacturer', getLabel: (v) => `Производитель: ${v}` },
  { key: 'manufacturerCode', getLabel: (v) => `Код производителя: ${v}` },
  { key: 'maxPrice', getLabel: (v) => `Цена до: ${v} ₽` },
  { key: 'maxQuantity', getLabel: (v) => `Количество до: ${v} шт` },
  { key: 'minPrice', getLabel: (v) => `Цена от: ${v} ₽` },
  { key: 'minQuantity', getLabel: (v) => `Количество от: ${v} шт` },
  { key: 'model', getLabel: (v) => `Модель: ${v}` },
  { key: 'number', getLabel: (v) => `Номер детали: ${v}` },
  { key: 'oemCode', getLabel: (v) => `OEM номер: ${v}` },
  { key: 'offset', getLabel: (v) => `Вылет (ET): ${v} мм` },
  { key: 'profile', getLabel: (v) => `Профиль шины: ${v}%` },
  { key: 'salesman', getLabel: (v) => `Продавец: ${v}` },
  { key: 'season', getLabel: (v) => `Сезонность: ${v}` },
  { key: 'status', getLabel: (v) => `Статус: ${v === 'true' ? 'Доступно' : 'Недоступно'}` },
  { key: 'supplierCode', getLabel: (v) => `Код поставщика: ${v}` },
  { key: 'tireQuantity', getLabel: (v) => `Количество шин: ${v} шт` },
  { key: 'topBottom', getLabel: (v) => `Положение: ${v}` },
  { key: 'transmission', getLabel: (v) => `КПП: ${v}` },
  { key: 'transmissionModel', getLabel: (v) => `Модель КПП: ${v}` },
  { key: 'vin', getLabel: (v) => `VIN: ${v}` },
  { key: 'wearPercentage', getLabel: (v) => `Износ: ${v}%` },
  { key: 'width', getLabel: (v) => `Ширина: ${v} мм` },
  { key: 'tireModel', getLabel: (v) => `Модель шины: ${v}` },
];

export function toPartFilters(f: SearchFiltersState): PartFilters {
  return {
    search: f.searchQuery,
    category: f.category,
    brand: f.brand,
    model: f.model,
    location: f.location,
    address: f.address,
    salesman: f.salesman,
    status: f.status,
    hasPhoto: f.hasPhoto === 'all' ? '' : f.hasPhoto,
    number: f.number,
    oem_code: f.oemCode,
    vin: f.vin,
    car_release_period: f.carReleasePeriod,
    body_brand: f.bodyBrand,
    engine_brand: f.engineBrand,
    car_release_date: f.carReleaseDate,
    transmission: f.transmission,
    drive: f.drive,
    condition: f.condition,
    manufacturer: f.manufacturer,
    defect: f.defect,
    color: f.color,
    min_price: f.minPrice,
    max_price: f.maxPrice,
    min_quantity: f.minQuantity,
    max_quantity: f.maxQuantity,
    front_rear: f.frontRear,
    left_right: f.leftRight,
    top_bottom: f.topBottom,
    manufacturer_code: f.manufacturerCode,
    supplier_code: f.supplierCode,
    transmission_model: f.transmissionModel,
    wear_percentage: f.wearPercentage,
    season: f.season,
    diameter: f.diameter,
    width: f.width,
    profile: f.profile,
    tire_quantity: f.tireQuantity,
    drilling: f.drilling,
    offset: f.offset,
    center_hole_diameter: f.centerHoleDiameter,
    tire_model: f.tireModel,
  };
}
