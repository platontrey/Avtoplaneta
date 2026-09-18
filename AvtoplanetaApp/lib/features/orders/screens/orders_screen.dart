import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../app/theme.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/order.dart';
import '../../../shared/widgets/app_states.dart';
import '../../auth/providers/auth_provider.dart';
import '../../inventory/providers/inventory_provider.dart';
import '../providers/orders_provider.dart';

class OrdersScreen extends ConsumerWidget {
  const OrdersScreen({super.key});

  Future<void> _refresh(WidgetRef ref) async {
    ref.invalidate(ordersProvider);
    await ref.read(ordersProvider.future);
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final ordersAsync = ref.watch(ordersProvider);
    final user = ref.watch(authProvider).valueOrNull;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Заказы'),
        actions: [
          IconButton(
            tooltip: 'Обновить',
            icon: const Icon(LucideIcons.refresh_cw),
            onPressed: () => _refresh(ref),
          ),
        ],
      ),
      body: ordersAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => AppEmptyState(
          icon: LucideIcons.cloud_off,
          title: 'Не удалось загрузить заказы',
          message: 'Проверьте соединение и повторите попытку.',
          actionLabel: 'Повторить',
          onAction: () => ref.invalidate(ordersProvider),
        ),
        data: (data) => RefreshIndicator(
          onRefresh: () => _refresh(ref),
          child: ListView.separated(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.fromLTRB(16, 10, 16, 20),
            itemCount: data.orders.length + (data.orders.isEmpty ? 3 : 2),
            separatorBuilder: (_, i) => SizedBox(height: i == 0 ? 14 : 10),
            itemBuilder: (_, i) {
              if (i == 0) {
                return Card(
                  child: ListTile(
                    leading: const CircleAvatar(
                      child: Icon(LucideIcons.shopping_cart),
                    ),
                    title: const Text('Создать новый заказ'),
                    subtitle: const Text(
                      'Выберите запчасть в инвентаре и оформите заказ',
                    ),
                    trailing: const Icon(LucideIcons.chevron_right),
                    onTap: () => context.go('/inventory'),
                  ),
                );
              }
              if (i == 1) {
                return AppSectionHeader(
                  title: 'Активность',
                  caption: '${data.total} заказов',
                );
              }
              if (data.orders.isEmpty) {
                return const AppEmptyState(
                  icon: LucideIcons.receipt,
                  title: 'Заказов пока нет',
                  message: 'Выберите запчасть и создайте первый заказ.',
                );
              }
              return _OrderCard(
                order: data.orders[i - 2],
                canChangeStatus: user?.isOperator == true,
                onStatusChanged: () {
                  ref.invalidate(ordersProvider);
                  ref.invalidate(inventoryProvider);
                },
              );
            },
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
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 5,
                  ),
                  decoration: BoxDecoration(
                    color: statusColor.withValues(alpha: 0.12),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(
                      color: statusColor.withValues(alpha: 0.25),
                    ),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        width: 7,
                        height: 7,
                        decoration: BoxDecoration(
                          color: statusColor,
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 7),
                      Text(
                        order.displayStatusText,
                        style: TextStyle(
                          color: statusColor,
                          fontSize: 11,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ],
                  ),
                ),
                const Spacer(),
                Text(
                  '#${order.id}',
                  style: const TextStyle(
                    color: AppTheme.mutedColor,
                    fontWeight: FontWeight.w700,
                    fontSize: 12,
                  ),
                ),
                const SizedBox(width: 10),
                Text(
                  order.timeAgo.isNotEmpty
                      ? order.timeAgo
                      : order.createdAtFormatted,
                  style: const TextStyle(
                    color: AppTheme.mutedColor,
                    fontSize: 11,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 14),
            Text(
              '${order.totalQuantity} × ${order.partName}',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 16,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 10),
            Wrap(
              spacing: 12,
              runSpacing: 8,
              children: [
                if (order.location.isNotEmpty)
                  _info(LucideIcons.map_pin, order.location),
                if (order.sellerName.isNotEmpty)
                  _info(LucideIcons.user, order.sellerName),
              ],
            ),
            if (order.orderNumber.isNotEmpty ||
                order.buyerNumber.isNotEmpty) ...[
              const SizedBox(height: 8),
              Wrap(
                spacing: 12,
                runSpacing: 8,
                children: [
                  if (order.orderNumber.isNotEmpty)
                    _info(LucideIcons.hash, '# ${order.orderNumber}'),
                  if (order.buyerNumber.isNotEmpty)
                    _info(LucideIcons.phone, order.buyerNumber),
                ],
              ),
            ],
            if (canChangeStatus) ...[
              const SizedBox(height: 8),
              _StatusButtons(orderId: order.id, onChanged: onStatusChanged),
              const SizedBox(height: 8),
              _OrderActions(orderId: order.id, onChanged: onStatusChanged),
            ],
          ],
        ),
      ),
    );
  }

  Widget _info(IconData icon, String text) => Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      Icon(icon, size: 14, color: AppTheme.mutedColor),
      const SizedBox(width: 5),
      Text(
        text,
        style: const TextStyle(color: AppTheme.mutedColor, fontSize: 12),
      ),
    ],
  );
}

class _OrderActions extends StatelessWidget {
  final int orderId;
  final VoidCallback onChanged;
  const _OrderActions({required this.orderId, required this.onChanged});

  Future<void> _confirmAction(
    BuildContext context, {
    required String title,
    required String description,
    required String actionLabel,
    required String path,
    required String method,
  }) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(title),
        content: Text(description),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: const Text('Отмена'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            child: Text(actionLabel),
          ),
        ],
      ),
    );
    if (confirmed != true || !context.mounted) {
      return;
    }

    try {
      if (method == 'DELETE') {
        await apiClient.dio.delete(path);
      } else {
        await apiClient.dio.put(path);
      }
      if (!context.mounted) {
        return;
      }
      onChanged();
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text('$actionLabel: выполнено')));
    } catch (error) {
      if (!context.mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Не удалось выполнить действие: $error')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: FilledButton.tonal(
            onPressed: () => _confirmAction(
              context,
              title: 'Подтверждение завершения продажи',
              description:
                  'Завершить продажу по заказу $orderId? Заказ будет учтён в статистике продаж.',
              actionLabel: 'Завершить',
              path: '/admin/orders/$orderId/complete',
              method: 'PUT',
            ),
            child: const Text('Завершить'),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: OutlinedButton(
            onPressed: () => _confirmAction(
              context,
              title: 'Подтверждение удаления',
              description:
                  'Удалить заказ $orderId? Количество запчастей будет восстановлено.',
              actionLabel: 'Удалить',
              path: '/admin/orders/$orderId',
              method: 'DELETE',
            ),
            style: OutlinedButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Удалить'),
          ),
        ),
      ],
    );
  }
}

class _StatusButtons extends ConsumerWidget {
  final int orderId;
  final VoidCallback onChanged;
  const _StatusButtons({required this.orderId, required this.onChanged});

  static const statuses = [
    ('red', 'Нужен транспорт', Color(0xFFE53935)),
    ('brown', 'Ожидание ответа', Color(0xFF795548)),
    ('yellow', 'Нужна доставка', Color(0xFFFDD835)),
    ('green', 'Доставлено', Color(0xFF43A047)),
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
                  if (!context.mounted) {
                    return;
                  }
                  onChanged();
                } catch (error) {
                  if (!context.mounted) {
                    return;
                  }
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      content: Text('Не удалось изменить статус: $error'),
                    ),
                  );
                }
              },
              style: OutlinedButton.styleFrom(
                side: BorderSide(color: s.$3, width: 1),
                foregroundColor: s.$3,
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 4,
                ),
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
