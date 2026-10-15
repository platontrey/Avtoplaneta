import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/order.dart';
import '../repositories/orders_repository.dart';

export '../repositories/orders_repository.dart';

/// Провайдер активных заказов
final ordersProvider = FutureProvider<OrdersResponse>((ref) async {
  final repository = ref.watch(ordersRepositoryProvider);
  return repository.fetchOrders();
});

/// Провайдер завершённых заказов
final completedOrdersProvider = FutureProvider<OrdersResponse>((ref) async {
  final repository = ref.watch(ordersRepositoryProvider);
  return repository.fetchOrders(state: 'completed');
});
