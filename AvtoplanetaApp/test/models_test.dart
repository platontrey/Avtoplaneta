import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/core/models/part.dart';
import 'package:avtoplaneta_app/core/models/order.dart';
import 'package:avtoplaneta_app/core/models/statistics.dart';
import 'package:avtoplaneta_app/core/models/user.dart';
import 'package:avtoplaneta_app/features/inventory/data/part_catalog.dart';
import 'package:avtoplaneta_app/features/inventory/providers/inventory_provider.dart';

void main() {
  group('Part Model Tests', () {
    test('parsing from json', () {
      final partJson = {
        'id': 12,
        'name': 'Фара левая',
        'description': 'Б/У, оригинал',
        'category': 'Оптика',
        'price': 15000.50,
        'quantity': 3,
        'brand': 'Toyota',
        'model': 'Camry',
        'color': 'Черный',
        'condition': 'Хорошее',
        'oem_code': '81150-33630',
        'supplier_code': 'SUPP-998',
        'vin': 'WBAXX...1',
        'location': 'Стеллаж A-1',
        'salesman': 'Иван',
        'photos': ['/uploads/photo1.jpg', '/uploads/photo2.jpg'],
        'created_at': '2026-06-12T15:23:40Z',
        'marked_for_deletion': false,
      };

      final part = Part.fromJson(partJson);
      expect(part.id, 12);
      expect(part.name, 'Фара левая');
      expect(part.price, 15000.50);
      expect(part.quantity, 3);
      expect(part.photos.length, 2);
      expect(part.photos[0], '/uploads/photo1.jpg');
      expect(part.createdAt, isNotNull);
      expect(part.markedForDeletion, false);
    });

    test('serialization to json', () {
      final partToSerialize = Part(
        id: 99,
        name: 'Стартер',
        category: 'Двигатель',
        price: 5000.00,
        quantity: 1,
        brand: 'BMW',
        model: 'E90',
        location: 'Полка 2',
      );

      final serialized = partToSerialize.toJson();
      expect(serialized['name'], 'Стартер');
      expect(serialized['category'], 'Двигатель');
      expect(serialized['price'], 5000.0);
      expect(serialized['quantity'], 1);
      expect(serialized['brand'], 'BMW');
      expect(serialized['location'], 'Полка 2');
    });

    test('supports the current API photo and camelCase fields', () {
      final part = Part.fromJson({
        'id': 7,
        'name': 'Бампер',
        'quantity': 1,
        'photo': '/uploads/bumper.jpg',
        'oemCode': 'OEM-7',
        'supplierCode': 'SUP-7',
      });

      expect(part.photos, ['/uploads/bumper.jpg']);
      expect(part.oemCode, 'OEM-7');
      expect(part.supplierCode, 'SUP-7');
    });

    test('availability respects the API status with quantity fallback', () {
      final unavailable = Part.fromJson({
        'id': 8,
        'name': 'Деталь',
        'quantity': 5,
        'status': false,
      });
      final legacyAvailable = Part.fromJson({
        'id': 9,
        'name': 'Другая деталь',
        'quantity': 2,
      });

      expect(unavailable.isAvailable, false);
      expect(legacyAvailable.isAvailable, true);
    });
  });

  group('Inventory Filter Tests', () {
    test('counts every non-default advanced filter', () {
      const filter = InventoryFilter(
        category: 'Трансмиссия',
        brand: 'BMW',
        model: 'E90',
        location: 'A-12',
        salesman: 'Иван',
        status: 'true',
        hasPhoto: 'without',
      );

      expect(filter.activeFilterCount, 7);
    });

    test('copyWith can clear a filter and reset pagination', () {
      const filter = InventoryFilter(brand: 'BMW', page: 4);
      final cleared = filter.copyWith(brand: '', page: 1);

      expect(cleared.brand, isEmpty);
      expect(cleared.page, 1);
      expect(cleared.activeFilterCount, 0);
    });

    test('builds the exact query supported by the inventory API', () {
      const filter = InventoryFilter(
        search: 'рейка',
        category: 'Рулевое управление',
        brand: 'BMW',
        model: 'E90',
        location: 'A-12',
        salesman: 'Иван',
        status: 'true',
        hasPhoto: 'with',
        page: 3,
      );

      expect(filter.toQueryParameters(), {
        'page': 3,
        'limit': 20,
        'search': 'рейка',
        'category': 'Рулевое управление',
        'brand': 'BMW',
        'model': 'E90',
        'location': 'A-12',
        'salesman': 'Иван',
        'status': 'true',
        'hasPhoto': 'with',
      });
    });
  });

  group('API parity model tests', () {
    test('statistics accepts the camelCase response used by the website', () {
      final statistics = StatisticsData.fromJson({
        'totalParts': 4,
        'totalQuantity': 9,
        'totalValue': 125000.5,
        'totalEarnings': 25000,
        'categories': [
          {'name': 'Оптика', 'count': 3},
        ],
        'monthlySales': [
          {'month': '2026-08', 'sales': 25000},
        ],
      });

      expect(statistics.totalParts, 4);
      expect(statistics.totalQuantity, 9);
      expect(statistics.totalValue, 125000.5);
      expect(statistics.totalEarnings, 25000);
      expect(statistics.categories.single.name, 'Оптика');
      expect(statistics.monthlySales.single.month, '2026-08');
    });

    test('statistics keeps compatibility with snake_case responses', () {
      final statistics = StatisticsData.fromJson({
        'total_parts': '2',
        'total_quantity': 5,
        'total_value': '1000.25',
        'total_earnings': 300,
        'monthly_sales': const [],
      });

      expect(statistics.totalParts, 2);
      expect(statistics.totalValue, 1000.25);
    });

    test('orders use the same field names as the website', () {
      final order = Order.fromJson({
        'id': 11,
        'part': 'Фара левая',
        'seller': 'Иван',
        'part_id': 12,
        'status': 'red',
        'status_text': '',
      });

      expect(order.partName, 'Фара левая');
      expect(order.sellerName, 'Иван');
      expect(order.displayStatusText, 'Нужен транспорт');
    });

    test('orders parse item quantities and formatted dates', () {
      final order = Order.fromJson({
        'id': 21,
        'part': 'Комплект подвески',
        'status': 'yellow',
        'status_text': 'Needs delivery',
        'created_at_formatted': '2026-08-06 12:30:00',
        'items': [
          {'id': 1, 'part_id': 10, 'quantity': 2, 'price': 5000},
          {'id': 2, 'part_id': 11, 'quantity': 3, 'price': 2500.5},
        ],
      });

      expect(order.totalQuantity, 5);
      expect(order.createdAtFormatted, '2026-08-06 12:30:00');
      expect(order.items.last.price, 2500.5);
      expect(order.displayStatusText, 'Нужна доставка');
    });
  });

  group('User Model Tests', () {
    test('admin role checks', () {
      final admin = User.fromJson({
        'id': 1,
        'email': 'admin@avtoplaneta.ru',
        'name': 'Администратор',
        'role': 'admin',
      });
      expect(admin.isAdmin, true);
      expect(admin.isManager, true);
      expect(admin.isOperator, true);
    });

    test('manager role checks', () {
      final manager = User.fromJson({
        'id': 2,
        'email': 'manager@avtoplaneta.ru',
        'name': 'Менеджер',
        'role': 'manager',
      });
      expect(manager.isAdmin, false);
      expect(manager.isManager, true);
      expect(manager.isOperator, true);
    });

    test('operator role checks', () {
      final operatorUser = User.fromJson({
        'id': 3,
        'email': 'operator@avtoplaneta.ru',
        'name': 'Оператор',
        'role': 'operator',
      });
      expect(operatorUser.isAdmin, false);
      expect(operatorUser.isManager, false);
      expect(operatorUser.isOperator, true);
    });
  });

  group('Defect report catalog expansion', () {
    test('keeps template and vehicle specifications in selected parts', () {
      final catalog = PartCatalog.fromJson({
        'version': 'test',
        'attributes': const [],
        'part_form_categories': const [],
        'report_bindings': [
          {
            'source': 'year',
            'target': 'car_release_date',
          },
          {
            'source': 'drive',
            'target': 'drive',
            'categories': ['Трансмиссия'],
          },
        ],
        'parts': [
          {
            'id': 'part_1',
            'name': 'Привод',
            'category': 'Трансмиссия',
            'quantity': 0,
            'price': 1000,
            'defaults': {
              'front_rear': 'F',
              'left_right': 'L',
            },
          },
          {
            'id': 'part_2',
            'name': 'Стекло',
            'category': 'Стекла',
            'quantity': 0,
            'price': 2000,
          },
        ],
      });

      final parts = catalog.expandDefectReportParts(
        {'year': 2011, 'drive': 'Задний'},
        supplierCode: 'report-123',
      );

      expect(parts.first['front_rear'], 'F');
      expect(parts.first['left_right'], 'L');
      expect(parts.first['car_release_date'], '2011');
      expect(parts.first['drive'], 'Задний');
      expect(parts.first['supplier_code'], 'report-123');
      expect(parts.last['car_release_date'], '2011');
      expect(parts.last.containsKey('drive'), false);
    });
  });
}
