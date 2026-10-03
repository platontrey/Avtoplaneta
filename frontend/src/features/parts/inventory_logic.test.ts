/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { describe, it, expect } from 'vitest';
import type { Part } from './types';

describe('Inventory Quantity Business Logic', () => {
  // Функция проверки видимости запчасти в каталоге / инвентаре
  const isPartVisibleInCatalog = (part: { quantity: number }) => {
    return part.quantity >= 0;
  };

  // Определение типа наличия
  const getInventoryStatus = (part: { quantity: number }) => {
    if (part.quantity > 0) return 'in_stock';
    if (part.quantity === 0) return 'defective_or_special';
    if (part.quantity === -1) return 'sold_out_hidden';
    return 'unknown';
  };

  it('should treat parts with quantity > 0 as visible in stock', () => {
    const regularPart: Partial<Part> = { id: 1, name: 'Крыло', quantity: 5 };
    expect(isPartVisibleInCatalog(regularPart as Part)).toBe(true);
    expect(getInventoryStatus(regularPart as Part)).toBe('in_stock');
  });

  it('should keep defective/special parts with quantity == 0 visible in inventory and catalog', () => {
    const defectivePart: Partial<Part> = { id: 2, name: 'Дверь (дефект)', quantity: 0 };
    expect(isPartVisibleInCatalog(defectivePart as Part)).toBe(true);
    expect(getInventoryStatus(defectivePart as Part)).toBe('defective_or_special');
  });

  it('should hide fully sold parts with quantity == -1 from inventory and catalog', () => {
    const soldPart: Partial<Part> = { id: 3, name: 'Капот', quantity: -1 };
    expect(isPartVisibleInCatalog(soldPart as Part)).toBe(false);
    expect(getInventoryStatus(soldPart as Part)).toBe('sold_out_hidden');
  });

  it('should correctly simulate stock depletion upon order completion', () => {
    // Если остаток 1, и продали 1: новый остаток становится -1 (скрыта), а не 0 (дефект)
    const currentQuantity = 1;
    const soldAmount = 1;
    const newQuantity = (currentQuantity - soldAmount <= 0) ? -1 : currentQuantity - soldAmount;
    expect(newQuantity).toBe(-1);
    expect(isPartVisibleInCatalog({ quantity: newQuantity })).toBe(false);
  });
});
