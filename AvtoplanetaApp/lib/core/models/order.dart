class Order {
  final int id;
  final String partName;
  final int? partId;
  final String location;
  final String buyerNumber;
  final String orderNumber;
  final String status; // red, brown, yellow, green
  final String statusText;
  final String sellerName;
  final String timeAgo;
  final DateTime? createdAt;

  const Order({
    required this.id,
    required this.partName,
    this.partId,
    this.location = '',
    this.buyerNumber = '',
    this.orderNumber = '',
    required this.status,
    required this.statusText,
    this.sellerName = '',
    this.timeAgo = '',
    this.createdAt,
  });

  factory Order.fromJson(Map<String, dynamic> json) => Order(
        id: json['id'] as int,
        partName: json['part_name'] as String? ?? '',
        partId: json['part_id'] as int?,
        location: json['location'] as String? ?? '',
        buyerNumber: json['buyer_number'] as String? ?? '',
        orderNumber: json['order_number'] as String? ?? '',
        status: json['status'] as String? ?? 'red',
        statusText: json['status_text'] as String? ?? '',
        sellerName: json['seller_name'] as String? ?? '',
        timeAgo: json['time_ago'] as String? ?? '',
        createdAt: json['created_at'] != null
            ? DateTime.tryParse(json['created_at'] as String)
            : null,
      );

  // Цвет статуса для UI
  static const statusColors = {
    'red': 0xFFE53935,
    'brown': 0xFF795548,
    'yellow': 0xFFFDD835,
    'green': 0xFF43A047,
  };

  int get statusColor => statusColors[status] ?? 0xFF9E9E9E;
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
      total: json['total'] as int? ?? 0,
    );
  }
}
