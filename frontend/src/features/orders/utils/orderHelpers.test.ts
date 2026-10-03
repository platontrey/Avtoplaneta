/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { describe, it, expect } from 'vitest';
import { calcOrderTotal } from './orderHelpers';
import type { Order } from '@/lib/types';

describe('calcOrderTotal', () => {
  it('should return total_amount when explicitly provided and non-negative', () => {
    const order: Partial<Order> = {
      total_amount: 15000,
      items: [
        { id: 1, order_id: 100, part_id: 10, part_name: 'Бампер', quantity: 1, price: 99999 },
      ],
    };
    expect(calcOrderTotal(order as Order)).toBe(15000);
  });

  it('should calculate total from items subtotal minus discount when total_amount is not set', () => {
    const order: Partial<Order> = {
      items: [
        { id: 1, order_id: 101, part_id: 1, part_name: 'Фара левая', quantity: 2, price: 3000 },
        { id: 2, order_id: 101, part_id: 2, part_name: 'Фара правая', quantity: 1, price: 3500 },
      ],
      discount: 500,
    };
    // (2 * 3000) + (1 * 3500) - 500 = 6000 + 3500 - 500 = 9000
    expect(calcOrderTotal(order as Order)).toBe(9000);
  });

  it('should not return a negative total when discount exceeds subtotal', () => {
    const order: Partial<Order> = {
      items: [
        { id: 1, order_id: 102, part_id: 1, part_name: 'Крепеж', quantity: 1, price: 200 },
      ],
      discount: 1000,
    };
    expect(calcOrderTotal(order as Order)).toBe(0);
  });

  it('should handle orders with empty items array gracefully', () => {
    const order: Partial<Order> = {
      items: [],
      discount: 0,
    };
    expect(calcOrderTotal(order as Order)).toBe(0);
  });
});
