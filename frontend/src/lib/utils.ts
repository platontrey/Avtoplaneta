/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Утилита для объединения CSS классов с помощью clsx и tailwind-merge
 * @param inputs - Массив значений классов
 * @returns Объединенная строка классов
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Форматирует период выпуска автомобиля (например, 2001-2007).
 * После ввода 4-й цифры при вводе 5-й автоматически подставляет дефис.
 */
export function formatCarReleasePeriod(value: string): string {
  if (!value) return '';
  const digits = value.replace(/\D/g, '').slice(0, 8);
  if (digits.length > 4) {
    return `${digits.slice(0, 4)}-${digits.slice(4)}`;
  }
  if (digits.length === 4 && (value.endsWith('-') || value.endsWith('/'))) {
    return `${digits}-`;
  }
  return digits;
}