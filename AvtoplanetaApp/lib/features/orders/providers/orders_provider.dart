import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/order.dart';

final ordersProvider = FutureProvider<OrdersResponse>((ref) async {
  final response = await apiClient.dio.get('/orders');
  final data = response.data;
  if (data is List) {
    final orders = data
        .whereType<Map>()
        .map((order) => Order.fromJson(Map<String, dynamic>.from(order)))
        .toList();
    return OrdersResponse(orders: orders, total: orders.length);
  }
  if (data is Map) {
    return OrdersResponse.fromJson(Map<String, dynamic>.from(data));
  }
  return const OrdersResponse(orders: [], total: 0);
});
