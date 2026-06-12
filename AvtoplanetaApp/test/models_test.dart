import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/core/models/part.dart';
import 'package:avtoplaneta_app/core/models/user.dart';

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
}
