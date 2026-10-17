import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/core/models/part.dart';
import 'package:avtoplaneta_app/features/inventory/data/inventory_filter.dart';
import 'package:avtoplaneta_app/features/inventory/data/part_catalog.dart';

void main() {
  group('Condition and Season Options Tests', () {
    test('PartCatalog optionsForAttribute returns condition and season options', () {
      final catalog = PartCatalog(
        version: '1',
        attributes: [
          const PartCatalogAttribute(
            code: 'condition',
            label: 'Состояние',
            inputType: 'select',
            options: ['Б/У', 'Новое'],
          ),
          const PartCatalogAttribute(
            code: 'season',
            label: 'Сезон',
            inputType: 'select',
            options: ['Зимние', 'Летние', 'Всесезонные'],
          ),
        ],
        partFormCategories: [],
      );

      expect(catalog.optionsForAttribute('condition'), ['Б/У', 'Новое']);
      expect(catalog.optionsForAttribute('season'), ['Зимние', 'Летние', 'Всесезонные']);
    });

    test('InventoryFilter copyWith correctly handles condition and season', () {
      const filter = InventoryFilter();
      final updated = filter.copyWith(
        condition: 'Б/У',
        season: 'Зимние',
      );

      expect(updated.condition, 'Б/У');
      expect(updated.season, 'Зимние');
    });

    test('Part model parses condition and season', () {
      final part = Part.fromJson({
        'id': 100,
        'name': 'Шина Michelin',
        'condition': 'Б/У',
        'season': 'Зимние',
        'price': 5000,
        'quantity': 4,
      });

      expect(part.condition, 'Б/У');
      expect(part.season, 'Зимние');
    });
  });
}
