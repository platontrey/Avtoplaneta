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

/**
 * Безопасно извлекает инициалы пользователя (до 2 символов в верхнем регистре).
 * Защищает от исключений при отсутствии имени или пустых полях.
 */
export function getUserInitials(user?: { name?: string; username?: string; email?: string; initials?: string } | null): string {
  if (!user) return 'U';
  if (user.initials && user.initials.trim().length > 0) return user.initials.trim();
  const source = (user.name || user.username || user.email || 'U').trim();
  if (source.length === 0) return 'U';
  const parts = source.split(/\s+/).filter(Boolean);
  if (parts.length >= 2 && parts[0][0] && parts[1][0]) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return source.slice(0, 2).toUpperCase();
}