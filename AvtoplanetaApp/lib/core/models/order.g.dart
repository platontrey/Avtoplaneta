// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'order.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Order _$OrderFromJson(Map<String, dynamic> json) => _Order(
  id: (json['id'] as num?)?.toInt() ?? 0,
  customerId: (json['customer_id'] as num?)?.toInt(),
  partName: _readPartName(json, 'part_name') as String? ?? '',
  partId: (json['part_id'] as num?)?.toInt(),
  location: json['location'] as String? ?? '',
  buyerNumber: json['buyer_number'] as String? ?? '',
  orderNumber: json['order_number'] as String? ?? '',
  status: json['status'] as String? ?? 'Принят в обработку',
  statusText: json['status_text'] as String? ?? '',
  source: json['source'] as String? ?? 'drom',
  paymentStatus: json['payment_status'] as String? ?? 'unpaid',
  warehouseStatus: json['warehouse_status'] as String? ?? 'inspecting',
  deliveryMethod: json['delivery_method'] as String? ?? 'tk',
  transportCompany: json['transport_company'] as String? ?? '',
  trackingNumber: json['tracking_number'] as String? ?? '',
  notes: json['notes'] as String? ?? '',
  discount: (json['discount'] as num?)?.toDouble() ?? 0.0,
  totalAmount: (json['total_amount'] as num?)?.toDouble() ?? 0.0,
  sellerName: _readSellerName(json, 'seller_name') as String? ?? '',
  timeAgo: json['time_ago'] as String? ?? '',
  createdAtFormatted: json['created_at_formatted'] as String? ?? '',
  completedAtFormatted: json['completed_at_formatted'] as String? ?? '',
  createdAt: json['created_at'] == null
      ? null
      : DateTime.parse(json['created_at'] as String),
  autoDeleted: json['auto_deleted'] as bool? ?? false,
  items:
      (json['items'] as List<dynamic>?)
          ?.map((e) => OrderItem.fromJson(e as Map<String, dynamic>))
          .toList() ??
      const [],
);

Map<String, dynamic> _$OrderToJson(_Order instance) => <String, dynamic>{
  'id': instance.id,
  'customer_id': instance.customerId,
  'part_name': instance.partName,
  'part_id': instance.partId,
  'location': instance.location,
  'buyer_number': instance.buyerNumber,
  'order_number': instance.orderNumber,
  'status': instance.status,
  'status_text': instance.statusText,
  'source': instance.source,
  'payment_status': instance.paymentStatus,
  'warehouse_status': instance.warehouseStatus,
  'delivery_method': instance.deliveryMethod,
  'transport_company': instance.transportCompany,
  'tracking_number': instance.trackingNumber,
  'notes': instance.notes,
  'discount': instance.discount,
  'total_amount': instance.totalAmount,
  'seller_name': instance.sellerName,
  'time_ago': instance.timeAgo,
  'created_at_formatted': instance.createdAtFormatted,
  'completed_at_formatted': instance.completedAtFormatted,
  'created_at': instance.createdAt?.toIso8601String(),
  'auto_deleted': instance.autoDeleted,
  'items': instance.items.map((e) => e.toJson()).toList(),
};

_OrderItem _$OrderItemFromJson(Map<String, dynamic> json) => _OrderItem(
  id: (json['id'] as num?)?.toInt() ?? 0,
  partId: (json['part_id'] as num?)?.toInt() ?? 0,
  partName: _readItemPartName(json, 'part_name') as String? ?? '',
  partNameSnapshot: json['part_name_snapshot'] as String? ?? '',
  quantity: (json['quantity'] as num?)?.toInt() ?? 0,
  price: (json['price'] as num?)?.toDouble() ?? 0.0,
);

Map<String, dynamic> _$OrderItemToJson(_OrderItem instance) =>
    <String, dynamic>{
      'id': instance.id,
      'part_id': instance.partId,
      'part_name': instance.partName,
      'part_name_snapshot': instance.partNameSnapshot,
      'quantity': instance.quantity,
      'price': instance.price,
    };
