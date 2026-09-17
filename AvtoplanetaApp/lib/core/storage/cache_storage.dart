import 'package:hive_flutter/hive_flutter.dart';

class CacheStorage {
  static const _partsBoxName = 'parts_cache';
  static const _settingsBoxName = 'settings_cache';

  static Future<void> init() async {
    await Hive.initFlutter();
    await Future.wait([
      Hive.openBox(_partsBoxName),
      Hive.openBox(_settingsBoxName),
    ]);
  }

  static Box get _partsBox => Hive.box(_partsBoxName);
  static Box get _settingsBox => Hive.box(_settingsBoxName);

  /// Сохранение кэша запчастей (список Map)
  static Future<void> savePartsCache(List<dynamic> parts) async {
    await _partsBox.put('parts_list', parts);
    await _partsBox.put('cached_at', DateTime.now().millisecondsSinceEpoch);
  }

  /// Получение кэша запчастей
  static List<dynamic>? getPartsCache() {
    return _partsBox.get('parts_list') as List<dynamic>?;
  }

  /// Время последнего кэширования
  static DateTime? getCachedAt() {
    final ms = _partsBox.get('cached_at') as int?;
    if (ms == null) return null;
    return DateTime.fromMillisecondsSinceEpoch(ms);
  }

  static Future<void> savePartCatalog(Map<String, dynamic> catalog) async {
    await _settingsBox.put('part_catalog', catalog);
  }

  static Future<void> saveVehicleCatalog(Map<String, dynamic> catalog) async {
    await _settingsBox.put('vehicle_catalog', catalog);
  }

  static Map<String, dynamic>? getVehicleCatalog() {
    final value = _settingsBox.get('vehicle_catalog');
    if (value == null) return null;
    return Map<String, dynamic>.from(value as Map);
  }

  static Map<String, dynamic>? getPartCatalog() {
    final value = _settingsBox.get('part_catalog');
    if (value is! Map) {
      return null;
    }
    return Map<String, dynamic>.from(value);
  }

  /// Очистка всего кэша
  static Future<void> clearCache() async {
    await Future.wait([_partsBox.clear(), _settingsBox.clear()]);
  }
}
