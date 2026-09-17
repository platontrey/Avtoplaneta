/// Справочник марок и моделей автомобилей.
///
/// Данные приходят с сервера (GET /api/vehicle-catalog) и общие с веб-клиентом:
/// список марок не хранится в коде приложения, его источник —
/// backend/parts-service/vehicles/vehicles.json.
class VehicleModel {
  final String name;
  final String slug;

  const VehicleModel({required this.name, this.slug = ''});

  factory VehicleModel.fromJson(Map<String, dynamic> json) => VehicleModel(
        name: json['name'] as String? ?? '',
        slug: json['slug'] as String? ?? '',
      );
}

class VehicleBrand {
  final String name;
  final String slug;
  final List<VehicleModel> models;

  const VehicleBrand({
    required this.name,
    this.slug = '',
    this.models = const [],
  });

  factory VehicleBrand.fromJson(Map<String, dynamic> json) => VehicleBrand(
        name: json['name'] as String? ?? '',
        slug: json['slug'] as String? ?? '',
        models: (json['models'] as List<dynamic>? ?? const [])
            .whereType<Map>()
            .map((value) => VehicleModel.fromJson(Map<String, dynamic>.from(value)))
            .where((model) => model.name.isNotEmpty)
            .toList(),
      );
}

class VehicleCatalog {
  final String version;
  final List<VehicleBrand> brands;

  const VehicleCatalog({required this.version, this.brands = const []});

  factory VehicleCatalog.fromJson(Map<String, dynamic> json) => VehicleCatalog(
        version: json['version'] as String? ?? '',
        brands: (json['brands'] as List<dynamic>? ?? const [])
            .whereType<Map>()
            .map((value) => VehicleBrand.fromJson(Map<String, dynamic>.from(value)))
            .where((brand) => brand.name.isNotEmpty)
            .toList(),
      );

  List<String> get brandNames =>
      brands.map((brand) => brand.name).toList(growable: false);

  /// Модели марки. Сравнение без учёта регистра: в уже заведённых запчастях
  /// марка лежит свободным текстом.
  List<String> modelsOf(String? brand) {
    final needle = (brand ?? '').trim().toLowerCase();
    if (needle.isEmpty) return const [];
    for (final item in brands) {
      if (item.name.trim().toLowerCase() == needle) {
        return item.models.map((model) => model.name).toList(growable: false);
      }
    }
    return const [];
  }
}
