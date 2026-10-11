/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { Part } from "@/features/parts/types";

/**
 * Надежное копирование текста в буфер обмена для ПК и мобильных устройств (iOS Safari, Android, desktop)
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  if (!text) return false;

  // 1. Попытка через современный Clipboard API
  if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      // Игнорируем и переходим к запасному варианту
    }
  }

  // 2. Универсальный fallback через скрытый textarea (работает на смартфонах и в HTTP/локальной сети)
  try {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.top = '0';
    textArea.style.left = '0';
    textArea.style.width = '2em';
    textArea.style.height = '2em';
    textArea.style.padding = '0';
    textArea.style.border = 'none';
    textArea.style.outline = 'none';
    textArea.style.boxShadow = 'none';
    textArea.style.background = 'transparent';
    textArea.setAttribute('readonly', '');
    document.body.appendChild(textArea);

    textArea.focus();
    textArea.select();
    textArea.setSelectionRange(0, textArea.value.length);

    const successful = document.execCommand('copy');
    document.body.removeChild(textArea);
    return successful;
  } catch (err) {
    console.error('Ошибка копирования в буфер обмена:', err);
    return false;
  }
}

/**
 * Форматирует данные запчасти в понятный текстовый вид для копирования
 */
export function formatPartCopyText(part: Part): string {
  const lines: string[] = [];

  if (part.name) lines.push(part.name);

  const car = [part.brand, part.model].filter(Boolean).join(' ');
  if (car) lines.push(`Автомобиль: ${car}`);

  if (part.price && Number(part.price) > 0) {
    lines.push(`Цена: ₽${Number(part.price).toLocaleString('ru-RU')}`);
  }
  lines.push(`В наличии: ${part.quantity ?? 0} шт.`);

  if (part.location) {
    lines.push(`Склад: ${part.location}${part.address ? ` (${part.address})` : ''}`);
  }
  if (part.oem_code) lines.push(`OEM: ${part.oem_code}`);
  if (part.number) lines.push(`Номер: ${part.number}`);
  if (part.manufacturer_code) lines.push(`Код производителя: ${part.manufacturer_code}`);
  if (part.manufacturer) lines.push(`Производитель: ${part.manufacturer}`);
  if (part.vin) lines.push(`VIN / Номер кузова: ${part.vin}`);
  if (part.body_brand) lines.push(`Марка кузова: ${part.body_brand}`);
  if (part.engine_brand) lines.push(`Марка двигателя: ${part.engine_brand}`);
  if (part.car_release_date) lines.push(`Год выпуска: ${part.car_release_date}`);
  if (part.car_release_period) lines.push(`Период выпуска: ${part.car_release_period}`);
  if (part.condition) lines.push(`Состояние: ${part.condition}`);
  if (part.defect) lines.push(`Дефект: ${part.defect}`);
  if (part.transmission) lines.push(`Трансмиссия: ${part.transmission}`);
  if (part.transmission_model) lines.push(`Модель КПП: ${part.transmission_model}`);
  if (part.drive) lines.push(`Привод: ${part.drive}`);
  if (part.color) lines.push(`Цвет: ${part.color}`);
  if (part.supplier_code) lines.push(`Код поставки: ${part.supplier_code}`);

  if (part.description && part.description.trim()) {
    lines.push(`\nОписание:\n${part.description.trim()}`);
  }

  return lines.join('\n');
}
