import 'package:freezed_annotation/freezed_annotation.dart';

part 'part.freezed.dart';
part 'part.g.dart';

Object? _readPartField(Map json, String snakeKey, String camelKey) {
  return json[snakeKey] ?? json[camelKey];
}

@freezed
abstract class Part with _$Part {
  const Part._();

  @JsonSerializable(fieldRename: FieldRename.snake)
  const factory Part({
    @Default(0) int id,
    @Default('') String name,
    String? description,
    @Default('') String category,
    @Default(0.0) double price,
    @Default(0) int quantity,
    bool? status,
    String? brand,
    String? model,
    @JsonKey(readValue: _readBodyBrand) String? bodyBrand,
    @JsonKey(readValue: _readEngineBrand) String? engineBrand,
    @JsonKey(readValue: _readCarReleaseDate) String? carReleaseDate,
    @JsonKey(readValue: _readFrontRear) String? frontRear,
    @JsonKey(readValue: _readLeftRight) String? leftRight,
    @JsonKey(readValue: _readTopBottom) String? topBottom,
    String? number,
    String? manufacturer,
    @JsonKey(readValue: _readManufacturerCode) String? manufacturerCode,
    String? color,
    String? condition,
    @JsonKey(readValue: _readOemCode) String? oemCode,
    @JsonKey(readValue: _readSupplierCode) String? supplierCode,
    @JsonKey(readValue: _readTransmissionModel) String? transmissionModel,
    String? defect,
    String? transmission,
    String? drive,
    @JsonKey(readValue: _readWearPercentage) String? wearPercentage,
    String? season,
    String? diameter,
    String? width,
    String? profile,
    @JsonKey(readValue: _readTireQuantity) String? tireQuantity,
    String? drilling,
    String? offset,
    @JsonKey(readValue: _readCenterHoleDiameter) String? centerHoleDiameter,
    @JsonKey(readValue: _readTireModel) String? tireModel,
    String? vin,
    @JsonKey(readValue: _readCarReleasePeriod) String? carReleasePeriod,
    @Default('') String location,
    @Default('') String address,
    String? salesman,
    @Default([]) List<String> photos,
    @JsonKey(readValue: _readCreatedAt) DateTime? createdAt,
    @JsonKey(readValue: _readMarkedForDeletion) @Default(false) bool markedForDeletion,
  }) = _Part;

  factory Part.fromJson(Map<String, dynamic> json) =>
      _$PartFromJson(_normalizePartJson(json));

  /// Доменные инварианты Avtoplaneta (из AGENTS.md):
  /// - quantity > 0: в наличии на складе
  /// - quantity == 0: дефектные ведомости / спецпозиции (видимы!)
  /// - quantity == -1: полностью списанные/проданные детали
  bool get isAvailable => status ?? quantity > 0;
  bool get isSold => quantity == -1;
  bool get isDefectReport => quantity == 0;
  bool get isInStock => quantity >= 0;
}

Map<String, dynamic> _normalizePartJson(Map<String, dynamic> json) {
  final normalized = Map<String, dynamic>.from(json);

  // Поддержка legacy одиночного поля photo
  final photosRaw = normalized['photos'];
  List<String> photosList = [];
  if (photosRaw is List) {
    photosList = photosRaw
        .map((e) => e.toString())
        .where((value) => value.isNotEmpty)
        .toList();
  }
  final legacyPhoto = normalized['photo']?.toString();
  if (photosList.isEmpty && legacyPhoto != null && legacyPhoto.isNotEmpty) {
    photosList = [legacyPhoto];
  }
  normalized['photos'] = photosList;

  // Безопасное приведение числовых типов из JSON
  if (normalized['price'] != null) {
    normalized['price'] = (normalized['price'] as num).toDouble();
  }
  if (normalized['quantity'] != null) {
    normalized['quantity'] = (normalized['quantity'] as num).toInt();
  }
  if (normalized['id'] != null) {
    normalized['id'] = (normalized['id'] as num).toInt();
  }

  return normalized;
}

Object? _readBodyBrand(Map json, String key) => _readPartField(json, 'body_brand', 'bodyBrand');
Object? _readEngineBrand(Map json, String key) => _readPartField(json, 'engine_brand', 'engineBrand');
Object? _readCarReleaseDate(Map json, String key) => _readPartField(json, 'car_release_date', 'carReleaseDate');
Object? _readFrontRear(Map json, String key) => _readPartField(json, 'front_rear', 'frontRear');
Object? _readLeftRight(Map json, String key) => _readPartField(json, 'left_right', 'leftRight');
Object? _readTopBottom(Map json, String key) => _readPartField(json, 'top_bottom', 'topBottom');
Object? _readManufacturerCode(Map json, String key) => _readPartField(json, 'manufacturer_code', 'manufacturerCode');
Object? _readOemCode(Map json, String key) => _readPartField(json, 'oem_code', 'oemCode');
Object? _readSupplierCode(Map json, String key) => _readPartField(json, 'supplier_code', 'supplierCode');
Object? _readTransmissionModel(Map json, String key) => _readPartField(json, 'transmission_model', 'transmissionModel');
Object? _readWearPercentage(Map json, String key) => _readPartField(json, 'wear_percentage', 'wearPercentage');
Object? _readTireQuantity(Map json, String key) => _readPartField(json, 'tire_quantity', 'tireQuantity');
Object? _readCenterHoleDiameter(Map json, String key) => _readPartField(json, 'center_hole_diameter', 'centerHoleDiameter');
Object? _readTireModel(Map json, String key) => _readPartField(json, 'tire_model', 'tireModel');
Object? _readCarReleasePeriod(Map json, String key) => _readPartField(json, 'car_release_period', 'carReleasePeriod');
Object? _readCreatedAt(Map json, String key) => _readPartField(json, 'created_at', 'createdAt');
Object? _readMarkedForDeletion(Map json, String key) => _readPartField(json, 'marked_for_deletion', 'markedForDeletion');

@freezed
abstract class InventoryResponse with _$InventoryResponse {
  const factory InventoryResponse({
    @Default([]) List<Part> parts,
    @Default(0) int total,
    @Default(1) int page,
    @Default(20) int limit,
    @Default(false) bool isOffline,
  }) = _InventoryResponse;

  factory InventoryResponse.fromJson(Map<String, dynamic> json) {
    final partsRaw = json['parts'] as List<dynamic>? ?? [];
    return InventoryResponse(
      parts: partsRaw
          .map((e) => Part.fromJson(e as Map<String, dynamic>))
          .toList(),
      total: (json['total'] as num?)?.toInt() ?? 0,
      page: (json['page'] as num?)?.toInt() ?? 1,
      limit: (json['limit'] as num?)?.toInt() ?? 20,
      isOffline: (json['is_offline'] ?? json['isOffline']) as bool? ?? false,
    );
  }
}
