/*
 * Copyright (c) 2026 Avtoplaneta. All rights reserved.
 */

import { describe, it, expect } from 'vitest';
import {
  buildBazaDromPartQuery,
  buildBazaDromSearchUrl,
  normalizeDromPropertyQuery,
} from './drom';

describe('drom utils', () => {
  describe('buildBazaDromSearchUrl', () => {
    it('should return base URL when query is empty', () => {
      expect(buildBazaDromSearchUrl('   ')).toBe('https://baza.drom.ru/sell_spare_parts/');
    });

    it('should encode query and collapse multiple spaces', () => {
      const url = buildBazaDromSearchUrl('  Фара   Toyota   Camry  ');
      expect(url).toBe(
        'https://baza.drom.ru/sell_spare_parts/?query=%D0%A4%D0%B0%D1%80%D0%B0+Toyota+Camry'
      );
    });
  });

  describe('buildBazaDromPartQuery', () => {
    it('should combine name, brand, model, and body_brand', () => {
      const query = buildBazaDromPartQuery({
        name: 'Фара правая LED',
        brand: 'Toyota',
        model: 'Camry',
        body_brand: 'XV70',
        engine_brand: '2AR-FE',
      });
      expect(query).toBe('Фара правая LED Toyota Camry XV70');
    });

    it('should avoid duplicating tokens already in part name', () => {
      const query = buildBazaDromPartQuery({
        name: 'Двигатель 1NZ-FE Toyota Corolla',
        brand: 'Toyota',
        model: 'Corolla',
        body_brand: 'NZE121',
      });
      expect(query).toBe('Двигатель 1NZ-FE Toyota Corolla NZE121');
    });

    it('should fallback to engine_brand when body_brand is absent', () => {
      const query = buildBazaDromPartQuery({
        name: 'Двигатель',
        brand: 'Toyota',
        model: 'Corolla',
        engine_brand: '1NZ-FE',
      });
      expect(query).toBe('Двигатель Toyota Corolla 1NZ-FE');
    });

    it('should include tire_model when brand and model are absent', () => {
      const query = buildBazaDromPartQuery({
        name: 'Шина летняя 205/55 R16',
        tire_model: 'Bridgestone Turanza',
      });
      expect(query).toBe('Шина летняя 205/55 R16 Bridgestone Turanza');
    });
  });

  describe('normalizeDromPropertyQuery', () => {
    it('should normalize front_rear and left_right codes', () => {
      expect(normalizeDromPropertyQuery('front_rear', 'F (Перед)')).toBe('перед');
      expect(normalizeDromPropertyQuery('front_rear', 'R')).toBe('зад');
      expect(normalizeDromPropertyQuery('left_right', 'L (Лево)')).toBe('лево');
      expect(normalizeDromPropertyQuery('left_right', 'R (Право)')).toBe('право');
    });

    it('should trim standard property values', () => {
      expect(normalizeDromPropertyQuery('oem_code', '  81110-33D30  ')).toBe('81110-33D30');
    });
  });
});
