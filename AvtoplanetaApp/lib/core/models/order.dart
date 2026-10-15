import 'package:freezed_annotation/freezed_annotation.dart';

part 'order.freezed.dart';
part 'order.g.dart';

@freezed
abstract class Order with _$Order {
  const Order._();

  @JsonSerializable(fieldRename: FieldRename.snake, explicitToJson: true)
  const factory Order({
    @Default(0) int id,
    int? customerId,
    @JsonKey(readValue: _readPartName) @Default('') String partName,
    int? partId,
    @Default('') String location,
    @Default('') String buyerNumber,
    @Default('') String orderNumber,
    @Default('Принят в обработку') String status,
    @Default('') String statusText,
    @Default('drom') String source,
    @Default('unpaid') String paymentStatus,
    @Default('inspecting') String warehouseStatus,
    @Default('tk') String deliveryMethod,
    @Default('') String transportCompany,
    @Default('') String trackingNumber,
    @Default('') String notes,
    @Default(0.0) double discount,
    @Default(0.0) double totalAmount,
    @JsonKey(readValue: _readSellerName) @Default('') String sellerName,
    @Default('') String timeAgo,
    @Default('') String createdAtFormatted,
    @Default('') String completedAtFormatted,
    DateTime? createdAt,
    @Default(false) bool autoDeleted,
    @Default([]) List<OrderItem> items,
  }) = _Order;

  factory Order.fromJson(Map<String, dynamic> json) =>
      _$OrderFromJson(_normalizeOrderJson(json));

  int get totalQuantity =>
      items.isEmpty ? 1 : items.fold(0, (total, item) => total + item.quantity);

  int? get effectivePartId {
    if (partId != null && partId! > 0) {
      return partId;
    }
    for (final item in items) {
      if (item.partId > 0) {
        return item.partId;
      }
    }
    return null;
  }

  // Цвета статусов для UI
  static const statusColors = {
    'Ожидает забора ТК': 0xFFFB8C00,
    'Требуется заказ ТК': 0xFFE53935,
    'К отправке в ТК': 0xFF1E88E5,
    'Ожидает трек-номер': 0xFF8E24AA,
    'Требует уточнения': 0xFFD81B60,
    'Принят в обработку': 0xFF039BE5,
    'На фотофиксации': 0xFF00ACC1,
    'Перемещение между складами': 0xFF5E35B1,
    'Ожидает предоплаты': 0xFFFDD835,
    'Проверен': 0xFF43A047,
    'Выдан / Завершён': 0xFF43A047,
    'red': 0xFFE53935,
    'brown': 0xFF795548,
    'yellow': 0xFFFDD835,
    'green': 0xFF43A047,
  };

  int get statusColor => statusColors[status] ?? 0xFF9E9E9E;

  String get displayStatusText {
    final localizedStatus = const {
      'red': 'Нужен транспорт',
      'brown': 'Ожидание ответа',
      'yellow': 'Нужна доставка',
      'green': 'Доставлено',
    }[status];
    if (localizedStatus != null) {
      return localizedStatus;
    }
    return status.isNotEmpty ? status : statusText;
  }
}

Map<String, dynamic> _normalizeOrderJson(Map<String, dynamic> json) {
  final normalized = Map<String, dynamic>.from(json);

  // Числовые типы
  if (normalized['id'] != null) {
    normalized['id'] = (normalized['id'] as num).toInt();
  }
  if (normalized['customer_id'] != null) {
    normalized['customer_id'] = (normalized['customer_id'] as num).toInt();
  }
  if (normalized['part_id'] != null) {
    normalized['part_id'] = (normalized['part_id'] as num).toInt();
  }
  if (normalized['discount'] != null) {
    normalized['discount'] = (normalized['discount'] as num).toDouble();
  }
  if (normalized['total_amount'] != null) {
    normalized['total_amount'] = (normalized['total_amount'] as num).toDouble();
  }

  // Items
  if (normalized['items'] is List) {
    normalized['items'] = (normalized['items'] as List)
        .whereType<Map>()
        .map((item) => OrderItem.fromJson(Map<String, dynamic>.from(item)))
        .map((item) => item.toJson())
        .toList();
  }

  return normalized;
}

Object? _readPartName(Map json, String key) =>
    json['part_name'] ?? json['part'] ?? json['partName'] ?? '';

Object? _readSellerName(Map json, String key) =>
    json['seller_name'] ?? json['seller'] ?? json['sellerName'] ?? '';

Map<String, dynamic> _normalizeOrderItemJson(Map<String, dynamic> json) {
  final normalized = Map<String, dynamic>.from(json);
  if (normalized['id'] != null) {
    normalized['id'] = (normalized['id'] as num).toInt();
  }
  if (normalized['part_id'] != null) {
    normalized['part_id'] = (normalized['part_id'] as num).toInt();
  }
  if (normalized['quantity'] != null) {
    normalized['quantity'] = (normalized['quantity'] as num).toInt();
  }
  if (normalized['price'] != null) {
    normalized['price'] = (normalized['price'] as num).toDouble();
  }
  return normalized;
}

@freezed
abstract class OrderItem with _$OrderItem {
  const OrderItem._();

  @JsonSerializable(fieldRename: FieldRename.snake)
  const factory OrderItem({
    @Default(0) int id,
    @Default(0) int partId,
    @JsonKey(readValue: _readItemPartName) @Default('') String partName,
    @Default('') String partNameSnapshot,
    @Default(0) int quantity,
    @Default(0.0) double price,
  }) = _OrderItem;

  factory OrderItem.fromJson(Map<String, dynamic> json) =>
      _$OrderItemFromJson(_normalizeOrderItemJson(json));
}

Object? _readItemPartName(Map json, String key) =>
    json['part_name'] ?? json['part_name_snapshot'] ?? json['partName'] ?? '';

@freezed
abstract class OrdersResponse with _$OrdersResponse {
  const factory OrdersResponse({
    @Default([]) List<Order> orders,
    @Default(0) int total,
  }) = _OrdersResponse;

  factory OrdersResponse.fromJson(Map<String, dynamic> json) {
    final ordersRaw = json['orders'] as List<dynamic>? ?? [];
    return OrdersResponse(
      orders: ordersRaw
          .whereType<Map>()
          .map((e) => Order.fromJson(Map<String, dynamic>.from(e)))
          .toList(),
      total: (json['total'] as num?)?.toInt() ?? 0,
    );
  }
}
