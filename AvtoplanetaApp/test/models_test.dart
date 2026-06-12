import '../lib/core/models/part.dart';
import '../lib/core/models/user.dart';

void main() {
  print('Running Part and User Models self-contained unit tests...');

  // === Part Model Tests ===
  // Test 1: parsing
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
  if (part.id != 12) throw Exception('Part parser failed: id');
  if (part.name != 'Фара левая') throw Exception('Part parser failed: name');
  if (part.price != 15000.50) throw Exception('Part parser failed: price');
  if (part.quantity != 3) throw Exception('Part parser failed: quantity');
  if (part.photos.length != 2) throw Exception('Part parser failed: photos length');
  if (part.photos[0] != '/uploads/photo1.jpg') throw Exception('Part parser failed: photos value');
  if (part.createdAt == null) throw Exception('Part parser failed: createdAt is null');
  if (part.markedForDeletion != false) throw Exception('Part parser failed: markedForDeletion');
  print('✓ Part parser test passed');

  // Test 2: serialization
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
  if (serialized['name'] != 'Стартер') throw Exception('Part serialization failed: name');
  if (serialized['category'] != 'Двигатель') throw Exception('Part serialization failed: category');
  if (serialized['price'] != 5000.0) throw Exception('Part serialization failed: price');
  if (serialized['quantity'] != 1) throw Exception('Part serialization failed: quantity');
  if (serialized['brand'] != 'BMW') throw Exception('Part serialization failed: brand');
  if (serialized['location'] != 'Полка 2') throw Exception('Part serialization failed: location');
  print('✓ Part serialization test passed');

  // === User Model Tests ===
  // Test 3: Admin role
  final admin = User.fromJson({
    'id': 1,
    'email': 'admin@avtoplaneta.ru',
    'name': 'Администратор',
    'role': 'admin',
  });
  if (!admin.isAdmin) throw Exception('User admin check failed: isAdmin');
  if (!admin.isManager) throw Exception('User admin check failed: isManager');
  if (!admin.isOperator) throw Exception('User admin check failed: isOperator');

  // Test 4: Manager role
  final manager = User.fromJson({
    'id': 2,
    'email': 'manager@avtoplaneta.ru',
    'name': 'Менеджер',
    'role': 'manager',
  });
  if (manager.isAdmin) throw Exception('User manager check failed: isAdmin should be false');
  if (!manager.isManager) throw Exception('User manager check failed: isManager');
  if (!manager.isOperator) throw Exception('User manager check failed: isOperator');

  // Test 5: Operator role
  final operatorUser = User.fromJson({
    'id': 3,
    'email': 'operator@avtoplaneta.ru',
    'name': 'Оператор',
    'role': 'operator',
  });
  if (operatorUser.isAdmin) throw Exception('User operator check failed: isAdmin should be false');
  if (operatorUser.isManager) throw Exception('User operator check failed: isManager should be false');
  if (!operatorUser.isOperator) throw Exception('User operator check failed: isOperator');

  print('✓ User roles and parsing test passed');
  print('\nAll Part and User model tests passed successfully!');
}
