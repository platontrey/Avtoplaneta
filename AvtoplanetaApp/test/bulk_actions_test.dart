import 'package:avtoplaneta_app/features/inventory/widgets/bulk_part_actions.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('bulk update payload follows the API contract', () {
    expect(
      buildBulkUpdatePayload(
        [12, 34],
        {'category': 'Трансмиссия', 'status': 'true'},
      ),
      {
        'parts': [
          {
            'id': 12,
            'fields': {'category': 'Трансмиссия', 'status': 'true'},
          },
          {
            'id': 34,
            'fields': {'category': 'Трансмиссия', 'status': 'true'},
          },
        ],
      },
    );
  });
}
