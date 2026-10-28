import 'package:avtoplaneta_app/features/inventory/widgets/bulk_part_actions.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('buildBulkUpdatePayload', () {
    test('follows the Protobuf API contract for IDS scope', () {
      expect(
        buildBulkUpdatePayload(
          ids: [12, 34],
          category: 'Трансмиссия',
          status: true,
          priceMode: PriceMode.percent,
          priceValue: 10.0,
        ),
        {
          'scope': {
            'type': 'BULK_SCOPE_TYPE_IDS',
            'ids': [12, 34],
          },
          'operations': {
            'price': {
              'mode': 'PRICE_OPERATION_MODE_PERCENT',
              'value': 10.0,
            },
            'category': 'Трансмиссия',
            'status': true,
          },
        },
      );
    });

    test('follows the Protobuf API contract for FILTER scope with cleaned pagination', () {
      expect(
        buildBulkUpdatePayload(
          ids: [],
          scopeType: 'BULK_SCOPE_TYPE_FILTER',
          filterQuery: {
            'page': 1,
            'limit': 20,
            'search': 'бампер',
            'category': 'Кузов',
          },
          location: 'Склад 1',
        ),
        {
          'scope': {
            'type': 'BULK_SCOPE_TYPE_FILTER',
            'filter': {
              'search': 'бампер',
              'category': 'Кузов',
            },
          },
          'operations': {
            'location': 'Склад 1',
          },
        },
      );
    });

    test('includes confirm_entire_catalog safeguard flag when requested', () {
      expect(
        buildBulkUpdatePayload(
          ids: [],
          scopeType: 'BULK_SCOPE_TYPE_FILTER',
          filterQuery: {},
          confirmEntireCatalog: true,
          salesman: 'Иванов',
        ),
        {
          'scope': {
            'type': 'BULK_SCOPE_TYPE_FILTER',
            'filter': {},
          },
          'operations': {
            'salesman': 'Иванов',
          },
          'confirm_entire_catalog': true,
        },
      );
    });

    test('includes body_brand and engine_brand in operations payload', () {
      expect(
        buildBulkUpdatePayload(
          ids: [100],
          bodyBrand: 'AE100',
          engineBrand: '4A-GE',
        ),
        {
          'scope': {
            'type': 'BULK_SCOPE_TYPE_IDS',
            'ids': [100],
          },
          'operations': {
            'body_brand': 'AE100',
            'engine_brand': '4A-GE',
          },
        },
      );
    });

    test('includes quantity and tire/wheel parameters in operations payload', () {
      expect(
        buildBulkUpdatePayload(
          ids: [200],
          quantity: 5,
          customOperations: {
            'diameter': '17',
            'width': '225',
            'height': '55',
            'season': 'Зимние (шипованные)',
            'pcd': '5x114.3',
            'et': '45',
            'front_rear': 'F',
            'left_right': 'L',
            'drive': 'Полный (4WD)',
          },
        ),
        {
          'scope': {
            'type': 'BULK_SCOPE_TYPE_IDS',
            'ids': [200],
          },
          'operations': {
            'quantity': 5,
            'diameter': '17',
            'width': '225',
            'height': '55',
            'season': 'Зимние (шипованные)',
            'pcd': '5x114.3',
            'et': '45',
            'front_rear': 'F',
            'left_right': 'L',
            'drive': 'Полный (4WD)',
          },
        },
      );
    });

    test('ALL_BULK_FIELDS contains 40 fields across all 3 sections', () {
      expect(ALL_BULK_FIELDS.length, 40);
      final basicFields = ALL_BULK_FIELDS.where((f) => f.section == FieldSection.basic).toList();
      final specsFields = ALL_BULK_FIELDS.where((f) => f.section == FieldSection.specs).toList();
      final tiresFields = ALL_BULK_FIELDS.where((f) => f.section == FieldSection.tires).toList();

      expect(basicFields.length, 14);
      expect(specsFields.length, 17);
      expect(tiresFields.length, 9);

      // Verify essential key fields
      expect(ALL_BULK_FIELDS.any((f) => f.key == 'price'), isTrue);
      expect(ALL_BULK_FIELDS.any((f) => f.key == 'quantity'), isTrue);
      expect(ALL_BULK_FIELDS.any((f) => f.key == 'location'), isTrue);
      expect(ALL_BULK_FIELDS.any((f) => f.key == 'front_rear'), isTrue);
      expect(ALL_BULK_FIELDS.any((f) => f.key == 'left_right'), isTrue);
      expect(ALL_BULK_FIELDS.any((f) => f.key == 'diameter'), isTrue);
      expect(ALL_BULK_FIELDS.any((f) => f.key == 'width'), isTrue);
      expect(ALL_BULK_FIELDS.any((f) => f.key == 'season'), isTrue);
    });
  });
}
