/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import { z } from 'zod';

export const partSchema = z.object({
    name: z.string().min(1, "Введите название запчасти").max(255, "Название слишком длинное (макс 255 символов)"),
    brand: z.string().min(1, "Выберите бренд"),
    model: z.string().min(1, "Введите модель"),
    quantity: z.number().min(0, "Количество должно быть неотрицательным числом"),
    price: z.number().optional(),
    description: z.string().max(1000, "Описание слишком длинное (макс 1000 символов)").optional(),
    category: z.string().optional(),
    location: z.string().optional(),
    address: z.string().optional(),
    seller_id: z.string().optional(),
    vin: z.string().optional(),

    // Характеристики кузова и двигателя
    body_brand: z.string().optional(),
    engine_brand: z.string().optional(),
    car_release_date: z.string().optional(),
    car_release_period: z.string().optional(),
    front_rear: z.string().optional(),
    left_right: z.string().optional(),
    top_bottom: z.string().optional(),
    color: z.string().optional(),

    // Трансмиссия и состояние
    transmission: z.string().optional(),
    transmission_model: z.string().optional(),
    drive: z.string().optional(),
    condition: z.string().optional(),
    defect: z.string().optional(),

    // Колеса и шины
    diameter: z.string().optional(),
    width: z.string().optional(),
    profile: z.string().optional(),
    tire_quantity: z.string().optional(),
    drilling: z.string().optional(),
    offset: z.string().optional(),
    center_hole_diameter: z.string().optional(),
    tire_model: z.string().optional(),
    wear_percentage: z.string().optional(),
    season: z.string().optional(),

    // Номера и поставщики
    number: z.string().optional(),
    manufacturer: z.string().optional(),
    manufacturer_code: z.string().optional(),
    oem_code: z.string().optional(),
    supplier_code: z.string().optional(),
});

export type PartFormData = z.infer<typeof partSchema>;
