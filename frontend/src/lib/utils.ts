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