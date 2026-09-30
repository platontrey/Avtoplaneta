class Order {
  final int id;
  final String partName;
  final int? partId;
  final String location;
  final String buyerNumber;
  final String orderNumber;
  final String status;
  final String statusText;
  final String source;
  final String paymentStatus;
  final String warehouseStatus;
  final String deliveryMethod;
  final String transportCompany;
  final String trackingNumber;
  final String notes;
  final double discount;
  final double totalAmount;
  final String sellerName;
  final String timeAgo;
  final String createdAtFormatted;
  final String completedAtFormatted;
  final DateTime? createdAt;
  final bool autoDeleted;
  final List<OrderItem> items;

  const Order({
    required this.id,
    required this.partName,
    this.partId,
    this.location = '',
    this.buyerNumber = '',
    this.orderNumber = '',
    required this.status,
    required this.statusText,
    this.source = 'drom',
    this.paymentStatus = 'unpaid',
    this.warehouseStatus = 'inspecting',
    this.deliveryMethod = 'tk',
    this.transportCompany = '',
    this.trackingNumber = '',
    this.notes = '',
    this.discount = 0,
    this.totalAmount = 0,
    this.sellerName = '',
    this.timeAgo = '',
    this.createdAtFormatted = '',
    this.completedAtFormatted = '',
    this.createdAt,
    this.autoDeleted = false,
    this.items = const [],
  });

  factory Order.fromJson(Map<String, dynamic> json) => Order(
    id: (json['id'] as num?)?.toInt() ?? 0,
    partName: (json['part'] ?? json['part_name'] ?? '').toString(),
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
    discount: (json['discount'] as num?)?.toDouble() ?? 0,
    totalAmount: (json['total_amount'] as num?)?.toDouble() ?? 0,
    sellerName: (json['seller'] ?? json['seller_name']) as String? ?? '',
    timeAgo: json['time_ago'] as String? ?? '',
    createdAtFormatted: json['created_at_formatted'] as String? ?? '',
    completedAtFormatted: json['completed_at_formatted'] as String? ?? '',
    createdAt: json['created_at'] != null
        ? DateTime.tryParse(json['created_at'].toString())
        : null,
    autoDeleted: json['auto_deleted'] as bool? ?? false,
    items: (json['items'] as List<dynamic>? ?? const [])
        .whereType<Map>()
        .map((item) => OrderItem.fromJson(Map<String, dynamic>.from(item)))
        .toList(),
  );

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

  // Цвет статуса для UI
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

class OrderItem {
  final int id;
  final int partId;
  final String partName;
  final String partNameSnapshot;
  final int quantity;
  final double price;

  const OrderItem({
    required this.id,
    required this.partId,
    this.partName = '',
    this.partNameSnapshot = '',
    required this.quantity,
    required this.price,
  });

  factory OrderItem.fromJson(Map<String, dynamic> json) => OrderItem(
    id: (json['id'] as num?)?.toInt() ?? 0,
    partId: (json['part_id'] as num?)?.toInt() ?? 0,
    partName: (json['part_name'] ?? json['part_name_snapshot'] ?? '').toString(),
    partNameSnapshot: (json['part_name_snapshot'] ?? '').toString(),
    quantity: (json['quantity'] as num?)?.toInt() ?? 0,
    price: (json['price'] as num?)?.toDouble() ?? 0,
  );
}

class OrdersResponse {
  final List<Order> orders;
  final int total;

  const OrdersResponse({required this.orders, required this.total});

  factory OrdersResponse.fromJson(Map<String, dynamic> json) {
    final ordersRaw = json['orders'] as List<dynamic>? ?? [];
    return OrdersResponse(
      orders: ordersRaw
          .map((e) => Order.fromJson(e as Map<String, dynamic>))
          .toList(),
      total: (json['total'] as num?)?.toInt() ?? 0,
    );
  }
}
