/*
 * Copyright (c) 2026 Avtoplaneta. All rights reserved.
 */

import type { Part } from '@/features/parts/types';

const BAZA_DROM_BASE_URL = 'https://baza.drom.ru/sell_spare_parts/';

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function containsToken(haystack: string, needle: string): boolean {
  if (!haystack || !needle) return false;
  const pattern = new RegExp(`(^|[\\s,./;()\\-])${escapeRegExp(needle)}($|[\\s,./;()\\-])`, 'i');
  return pattern.test(haystack);
}

/**
 * Формирует прямую ссылку поиска запчастей на baza.drom.ru
 */
export function buildBazaDromSearchUrl(query: string): string {
  const normalized = query.trim().replace(/\s+/g, ' ');
  if (!normalized) {
    return BAZA_DROM_BASE_URL;
  }

  const params = new URLSearchParams({ query: normalized });
  return `${BAZA_DROM_BASE_URL}?${params.toString()}`;
}

/**
 * Формирует поисковый запрос для запчасти целиком (название, марка, модель, кузов/двигатель)
 * без дублирования уже присутствующих в названии токенов.
 */
export function buildBazaDromPartQuery(
  part: Partial<Pick<Part, 'name' | 'brand' | 'model' | 'body_brand' | 'engine_brand' | 'tire_model'>>
): string {
  const candidates: Array<string | undefined> = [
    part.name,
    part.brand,
    part.model,
    part.body_brand,
  ];

  if (!part.body_brand?.trim() && part.engine_brand?.trim()) {
    candidates.push(part.engine_brand);
  }

  if (!part.brand?.trim() && !part.model?.trim() && part.tire_model?.trim()) {
    candidates.push(part.tire_model);
  }

  let result = '';
  for (const raw of candidates) {
    const token = raw?.trim().replace(/\s+/g, ' ');
    if (!token) continue;
    if (!result) {
      result = token;
    } else if (!containsToken(result, token)) {
      result = `${result} ${token}`;
    }
  }

  return result;
}

/**
 * Нормализует значение свойства для более точного поиска на Базе Дром
 */
export function normalizeDromPropertyQuery(code: string | undefined, value: string): string {
  const trimmed = value.trim();
  if (!trimmed) return '';

  if (code === 'front_rear') {
    if (/^(f|f\s*\(перед\)|перед|передний)$/i.test(trimmed)) return 'перед';
    if (/^(r|r\s*\(зад\)|зад|задний)$/i.test(trimmed)) return 'зад';
  }

  if (code === 'left_right') {
    if (/^(l|l\s*\(лево\)|лево|левый)$/i.test(trimmed)) return 'лево';
    if (/^(r|r\s*\(право\)|право|правый)$/i.test(trimmed)) return 'право';
  }

  return trimmed;
}

/**
 * Открывает поиск на baza.drom.ru в новой вкладке браузера
 */
export function openBazaDromSearch(query: string): void {
  const url = buildBazaDromSearchUrl(query);
  window.open(url, '_blank', 'noopener,noreferrer');
}
