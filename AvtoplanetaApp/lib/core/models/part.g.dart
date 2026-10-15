// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'part.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Part _$PartFromJson(Map<String, dynamic> json) => _Part(
  id: (json['id'] as num?)?.toInt() ?? 0,
  name: json['name'] as String? ?? '',
  description: json['description'] as String?,
  category: json['category'] as String? ?? '',
  price: (json['price'] as num?)?.toDouble() ?? 0.0,
  quantity: (json['quantity'] as num?)?.toInt() ?? 0,
  status: json['status'] as bool?,
  brand: json['brand'] as String?,
  model: json['model'] as String?,
  bodyBrand: _readBodyBrand(json, 'body_brand') as String?,
  engineBrand: _readEngineBrand(json, 'engine_brand') as String?,
  carReleaseDate: _readCarReleaseDate(json, 'car_release_date') as String?,
  frontRear: _readFrontRear(json, 'front_rear') as String?,
  leftRight: _readLeftRight(json, 'left_right') as String?,
  topBottom: _readTopBottom(json, 'top_bottom') as String?,
  number: json['number'] as String?,
  manufacturer: json['manufacturer'] as String?,
  manufacturerCode: _readManufacturerCode(json, 'manufacturer_code') as String?,
  color: json['color'] as String?,
  condition: json['condition'] as String?,
  oemCode: _readOemCode(json, 'oem_code') as String?,
  supplierCode: _readSupplierCode(json, 'supplier_code') as String?,
  transmissionModel:
      _readTransmissionModel(json, 'transmission_model') as String?,
  defect: json['defect'] as String?,
  transmission: json['transmission'] as String?,
  drive: json['drive'] as String?,
  wearPercentage: _readWearPercentage(json, 'wear_percentage') as String?,
  season: json['season'] as String?,
  diameter: json['diameter'] as String?,
  width: json['width'] as String?,
  profile: json['profile'] as String?,
  tireQuantity: _readTireQuantity(json, 'tire_quantity') as String?,
  drilling: json['drilling'] as String?,
  offset: json['offset'] as String?,
  centerHoleDiameter:
      _readCenterHoleDiameter(json, 'center_hole_diameter') as String?,
  tireModel: _readTireModel(json, 'tire_model') as String?,
  vin: json['vin'] as String?,
  carReleasePeriod:
      _readCarReleasePeriod(json, 'car_release_period') as String?,
  location: json['location'] as String? ?? '',
  address: json['address'] as String? ?? '',
  salesman: json['salesman'] as String?,
  photos:
      (json['photos'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      const [],
  createdAt: _readCreatedAt(json, 'created_at') == null
      ? null
      : DateTime.parse(_readCreatedAt(json, 'created_at') as String),
  markedForDeletion:
      _readMarkedForDeletion(json, 'marked_for_deletion') as bool? ?? false,
);

Map<String, dynamic> _$PartToJson(_Part instance) => <String, dynamic>{
  'id': instance.id,
  'name': instance.name,
  'description': instance.description,
  'category': instance.category,
  'price': instance.price,
  'quantity': instance.quantity,
  'status': instance.status,
  'brand': instance.brand,
  'model': instance.model,
  'body_brand': instance.bodyBrand,
  'engine_brand': instance.engineBrand,
  'car_release_date': instance.carReleaseDate,
  'front_rear': instance.frontRear,
  'left_right': instance.leftRight,
  'top_bottom': instance.topBottom,
  'number': instance.number,
  'manufacturer': instance.manufacturer,
  'manufacturer_code': instance.manufacturerCode,
  'color': instance.color,
  'condition': instance.condition,
  'oem_code': instance.oemCode,
  'supplier_code': instance.supplierCode,
  'transmission_model': instance.transmissionModel,
  'defect': instance.defect,
  'transmission': instance.transmission,
  'drive': instance.drive,
  'wear_percentage': instance.wearPercentage,
  'season': instance.season,
  'diameter': instance.diameter,
  'width': instance.width,
  'profile': instance.profile,
  'tire_quantity': instance.tireQuantity,
  'drilling': instance.drilling,
  'offset': instance.offset,
  'center_hole_diameter': instance.centerHoleDiameter,
  'tire_model': instance.tireModel,
  'vin': instance.vin,
  'car_release_period': instance.carReleasePeriod,
  'location': instance.location,
  'address': instance.address,
  'salesman': instance.salesman,
  'photos': instance.photos,
  'created_at': instance.createdAt?.toIso8601String(),
  'marked_for_deletion': instance.markedForDeletion,
};
