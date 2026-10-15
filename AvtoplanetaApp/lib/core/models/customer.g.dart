// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'customer.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Customer _$CustomerFromJson(Map<String, dynamic> json) => _Customer(
  id: (json['id'] as num?)?.toInt() ?? 0,
  name: json['name'] as String? ?? '',
  phone: json['phone'] as String? ?? '',
  city: json['city'] as String? ?? '',
  preferredTk: _readPreferredTk(json, 'preferredTk') as String? ?? '',
  tkDetails: _readTkDetails(json, 'tkDetails') as String? ?? '',
  notes: json['notes'] as String? ?? '',
  category: json['category'] as String? ?? 'regular',
  discountPercent:
      (_readDiscountPercent(json, 'discountPercent') as num?)?.toDouble() ??
      0.0,
  totalOrders: (_readTotalOrders(json, 'totalOrders') as num?)?.toInt() ?? 0,
  totalSpent: (_readTotalSpent(json, 'totalSpent') as num?)?.toDouble() ?? 0.0,
  createdAt: _readCreatedAt(json, 'createdAt') == null
      ? null
      : DateTime.parse(_readCreatedAt(json, 'createdAt') as String),
  updatedAt: _readUpdatedAt(json, 'updatedAt') == null
      ? null
      : DateTime.parse(_readUpdatedAt(json, 'updatedAt') as String),
);

Map<String, dynamic> _$CustomerToJson(_Customer instance) => <String, dynamic>{
  'id': instance.id,
  'name': instance.name,
  'phone': instance.phone,
  'city': instance.city,
  'preferredTk': instance.preferredTk,
  'tkDetails': instance.tkDetails,
  'notes': instance.notes,
  'category': instance.category,
  'discountPercent': instance.discountPercent,
  'totalOrders': instance.totalOrders,
  'totalSpent': instance.totalSpent,
  'createdAt': instance.createdAt?.toIso8601String(),
  'updatedAt': instance.updatedAt?.toIso8601String(),
};
