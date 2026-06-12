import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/order.dart';
import '../../auth/providers/auth_provider.dart';

final ordersProvider = FutureProvider<OrdersResponse>((ref) async {
  final response = await apiClient.dio.get('/orders');
  final data = response.data;
  if (data is List) {
    final orders = data
        .map((e) => Order.fromJson(e as Map<String, dynamic>))
        .toList();
    return OrdersResponse(orders: orders, total: orders.length);
  } else if (data is Map) {
    return OrdersResponse.fromJson(Map<String, dynamic>.from(data));
  }
  return const OrdersResponse(orders: [], total: 0);
});

class OrdersScreen extends ConsumerWidget {
  const OrdersScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final ordersAsync = ref.watch(ordersProvider);
    final user = ref.watch(authProvider).valueOrNull;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Заказы'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.invalidate(ordersProvider),
          ),
        ],
      ),
      body: ordersAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (data) => data.orders.isEmpty
            ? const Center(
                child: Text('Заказов нет', style: TextStyle(color: Colors.white54)))
            : RefreshIndicator(
                onRefresh: () async => ref.invalidate(ordersProvider),
                child: ListView.builder(
                  padding: const EdgeInsets.all(8),
                  itemCount: data.orders.length,
                  itemBuilder: (_, i) => _OrderCard(
                    order: data.orders[i],
                    canChangeStatus: user?.isManager == true,
                    onStatusChanged: () => ref.invalidate(ordersProvider),
                  ),
                ),
              ),
      ),
    );
  }
}

class _OrderCard extends StatelessWidget {
  final Order order;
  final bool canChangeStatus;
  final VoidCallback onStatusChanged;

  const _OrderCard({
    required this.order,
    required this.canChangeStatus,
    required this.onStatusChanged,
  });

  @override
  Widget build(BuildContext context) {
    final statusColor = Color(order.statusColor);

    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 4),
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 12,
                  height: 12,
                  decoration: BoxDecoration(
                    color: statusColor,
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    order.statusText,
                    style: TextStyle(
                        color: statusColor,
                        fontSize: 12,
                        fontWeight: FontWeight.w600),
                  ),
                ),
                Text(
                  order.timeAgo,
                  style:
                      const TextStyle(color: Colors.white38, fontSize: 11),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              order.partName,
              style: const TextStyle(
                  color: Colors.white,
                  fontSize: 15,
                  fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 4),
            Row(
              children: [
                if (order.location.isNotEmpty)
                  _info(Icons.location_on_outlined, order.location),
                if (order.sellerName.isNotEmpty) ...[
                  const SizedBox(width: 12),
                  _info(Icons.person_outline, order.sellerName),
                ],
              ],
            ),
            if (order.orderNumber.isNotEmpty || order.buyerNumber.isNotEmpty) ...[
              const SizedBox(height: 4),
              Row(
                children: [
                  if (order.orderNumber.isNotEmpty)
                    _info(Icons.tag, '# ${order.orderNumber}'),
                  if (order.buyerNumber.isNotEmpty) ...[
                    const SizedBox(width: 12),
                    _info(Icons.phone_outlined, order.buyerNumber),
                  ],
                ],
              ),
            ],
            if (canChangeStatus) ...[
              const SizedBox(height: 8),
              _StatusButtons(
                  orderId: order.id, onChanged: onStatusChanged),
            ],
          ],
        ),
      ),
    );
  }

  Widget _info(IconData icon, String text) => Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 13, color: Colors.white38),
          const SizedBox(width: 3),
          Text(text, style: const TextStyle(color: Colors.white54, fontSize: 12)),
        ],
      );
}

class _StatusButtons extends ConsumerWidget {
  final int orderId;
  final VoidCallback onChanged;
  const _StatusButtons({required this.orderId, required this.onChanged});

  static const statuses = [
    ('red', 'Заказать транспорт', Color(0xFFE53935)),
    ('brown', 'Ожидание', Color(0xFF795548)),
    ('yellow', 'Доставить', Color(0xFFFDD835)),
    ('green', 'Доставлен', Color(0xFF43A047)),
  ];

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: statuses.map((s) {
          return Padding(
            padding: const EdgeInsets.only(right: 6),
            child: OutlinedButton(
              onPressed: () async {
                try {
                  await apiClient.dio.put(
                    '/admin/orders/$orderId/status',
                    data: {'status': s.$1, 'status_text': s.$2},
                  );
                  onChanged();
                } catch (_) {}
              },
              style: OutlinedButton.styleFrom(
                side: BorderSide(color: s.$3, width: 1),
                foregroundColor: s.$3,
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                minimumSize: Size.zero,
                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
              child: Text(s.$2, style: const TextStyle(fontSize: 11)),
            ),
          );
        }).toList(),
      ),
    );
  }
}
