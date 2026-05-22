/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

export interface Part {
  id: number;
  name: string;
  quantity: number;
  description?: string;
  category?: string;
  price?: number;
  salesman?: string;
  location?: string;
  status?: boolean;
  brand?: string;
  model?: string;
  photos?: string[];
  photo?: string; // Для обратной совместимости
  inn?: string;                    // ИНН поставщика
  vin?: string;
  // Характеристики запчасти
  body_brand?: string;
  engine_brand?: string;
  car_release_date?: string;
  front_rear?: string;
  left_right?: string;
  top_bottom?: string;
  number?: string;
  manufacturer?: string;
  manufacturer_code?: string;
  oem_code?: string;
  color?: string;
  condition?: string;
  supplier_code?: string;
  defect?: string;
  transmission?: string;
  drive?: string;
  wear_percentage?: string;
  season?: string;
  diameter?: string;
  width?: string;
  profile?: string;
  tire_quantity?: string;
  drilling?: string;
  offset?: string;
  center_hole_diameter?: string;
  tire_model?: string;
}

export interface PartFormData {
   name: string;
   quantity: string;
   description: string;
   category: string;
   price: string;
   salesman: string;
   location: string;
   status: boolean;
   photos?: string[];
   photo?: string; // Для обратной совместимости
   // Характеристики запчасти
   body_brand?: string;
   engine_brand?: string;
   car_release_date?: string;
   front_rear?: string;
   left_right?: string;
   top_bottom?: string;
   number?: string;
   manufacturer?: string;
   manufacturer_code?: string;
   oem_code?: string;
   color?: string;
   condition?: string;
   supplier_code?: string;
   defect?: string;
   transmission?: string;
   drive?: string;
   wear_percentage?: string;
   season?: string;
   diameter?: string;
   width?: string;
   profile?: string;
   tire_quantity?: string;
   drilling?: string;
   offset?: string;
   center_hole_diameter?: string;
   tire_model?: string;
}

export interface PartFilters {
  category: string;
  brand: string;
  search: string;
}