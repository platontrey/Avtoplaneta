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
  });
}
