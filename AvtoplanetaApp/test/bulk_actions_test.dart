import 'package:avtoplaneta_app/features/inventory/widgets/bulk_part_actions.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('bulk update payload follows the Protobuf API contract', () {
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
}
