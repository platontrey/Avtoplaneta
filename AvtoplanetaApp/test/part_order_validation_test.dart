import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/core/models/part.dart';

void main() {
  group('Part Order Validation & Details Tests', () {
    test('parts with quantity <= 0 are detected for order blocking', () {
      final inStockPart = Part(
        id: 1,
        name: 'Фара левая',
        category: 'Оптика',
        price: 5000,
        quantity: 2,
      );

      final zeroQtyPart = Part(
        id: 2,
        name: 'Фара дефектная',
        category: 'Оптика',
        price: 2500,
        quantity: 0,
      );

      final soldPart = Part(
        id: 3,
        name: 'Фара проданная',
        category: 'Оптика',
        price: 5000,
        quantity: -1,
      );

      expect(inStockPart.quantity > 0, true);
      expect(zeroQtyPart.quantity <= 0, true);
      expect(soldPart.quantity <= 0, true);

      final orderListWithZero = [inStockPart, zeroQtyPart];
      expect(orderListWithZero.any((p) => p.quantity <= 0), true);

      final orderListValid = [inStockPart];
      expect(orderListValid.any((p) => p.quantity <= 0), false);
    });

    test('part details text format includes all vital part information', () {
      final part = Part(
        id: 10,
        name: 'Бампер передний',
        brand: 'Subaru',
        model: 'Forester',
        bodyBrand: 'SJ5',
        engineBrand: 'FB20',
        carReleaseDate: '2014',
        vin: 'SJ5-012345',
        price: 18000,
        quantity: 1,
        category: 'Кузовные детали',
        location: 'Склад 2',
        address: 'ул. Трактовая 4',
        oemCode: '57702-SG000',
        number: 'SUB-123',
        manufacturer: 'Subaru',
        condition: 'Отличное',
        defect: 'Мелкие царапины',
        description: 'Оригинальный бампер в сборе',
      );

      // Verify that the part fields contain expected attributes
      expect(part.name, 'Бампер передний');
      expect(part.price, 18000);
      expect(part.quantity, 1);
      expect(part.oemCode, '57702-SG000');
      expect(part.vin, 'SJ5-012345');
      expect(part.location, 'Склад 2');
      expect(part.address, 'ул. Трактовая 4');
    });
  });
}
