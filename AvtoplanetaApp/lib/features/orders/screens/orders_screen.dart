import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../app/theme.dart';
import '../../../core/api/api_client.dart';
import '../../../core/api/api_endpoints.dart';
import '../../../core/models/customer.dart';
import '../../../core/models/order.dart';
import '../../../core/utils/formatters.dart';
import '../../../shared/widgets/app_states.dart';
import '../../../shared/widgets/shimmer_skeletons.dart';
import '../../auth/providers/auth_provider.dart';
import '../../inventory/providers/inventory_provider.dart';
import '../providers/customers_provider.dart';
import '../providers/orders_provider.dart';
import '../widgets/customer_detail_sheet.dart';
import '../widgets/customer_form_sheet.dart';

class OrdersScreen extends ConsumerStatefulWidget {
  const OrdersScreen({super.key});

  @override
  ConsumerState<OrdersScreen> createState() => _OrdersScreenState();
}

class _OrdersScreenState extends ConsumerState<OrdersScreen> {
  String _tab = 'active'; // 'active', 'completed', or 'customers'
  String _customerCategory = ''; // '', 'regular', 'vip', 'wholesale', 'blacklist'
  String _searchQuery = '';
  final _searchController = TextEditingController();

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _refresh() async {
    ref.invalidate(ordersProvider);
    ref.invalidate(completedOrdersProvider);
    ref.invalidate(customersProvider);
    if (_tab == 'active') {
      await ref.read(ordersProvider.future);
    } else if (_tab == 'completed') {
      await ref.read(completedOrdersProvider.future);
    } else {
      await ref.read(customersProvider.future);
    }
  }

  List<Order> _filterOrders(List<Order> orders) {
    final q = _searchQuery.trim().toLowerCase();
    if (q.isEmpty) return orders;
    return orders.where((o) {
      return o.id.toString().contains(q) ||
          o.partName.toLowerCase().contains(q) ||
          o.buyerNumber.toLowerCase().contains(q) ||
          o.orderNumber.toLowerCase().contains(q) ||
          o.trackingNumber.toLowerCase().contains(q) ||
          o.sellerName.toLowerCase().contains(q);
    }).toList();
  }

  List<Customer> _filterCustomers(List<Customer> customers) {
    var result = customers;
    if (_customerCategory.isNotEmpty) {
      result = result.where((c) => c.category == _customerCategory).toList();
    }
    final q = _searchQuery.trim().toLowerCase();
    if (q.isNotEmpty) {
      result = result.where((c) {
        return c.name.toLowerCase().contains(q) ||
            c.phone.toLowerCase().contains(q) ||
            c.city.toLowerCase().contains(q) ||
            c.preferredTk.toLowerCase().contains(q) ||
            c.notes.toLowerCase().contains(q);
      }).toList();
    }
    return result;
  }

  Widget _categoryFilterChip(String key, String label) {
    final isSelected = _customerCategory == key;
    return ChoiceChip(
      label: Text(label),
      selected: isSelected,
      onSelected: (_) {
        setState(() => _customerCategory = key);
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final ordersAsync = _tab == 'active'
        ? ref.watch(ordersProvider)
        : ref.watch(completedOrdersProvider);
    final user = ref.watch(authProvider).valueOrNull;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Заказы'),
        actions: [
          IconButton(
            tooltip: 'Обновить',
            icon: const Icon(LucideIcons.refresh_cw),
            onPressed: _refresh,
          ),
        ],
      ),
      floatingActionButton: _tab == 'customers'
          ? FloatingActionButton.extended(
              onPressed: () => CustomerFormSheet.show(context),
              icon: const Icon(LucideIcons.user_plus),
              label: const Text('Новый клиент'),
            )
          : null,
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 10, 16, 6),
            child: Column(
              children: [
                SizedBox(
                  width: double.infinity,
                  child: SegmentedButton<String>(
                    segments: const [
                      ButtonSegment(
                        value: 'active',
                        icon: Icon(LucideIcons.shopping_bag, size: 16),
                        label: Text('В работе'),
                      ),
                      ButtonSegment(
                        value: 'completed',
                        icon: Icon(LucideIcons.archive, size: 16),
                        label: Text('История'),
                      ),
                      ButtonSegment(
                        value: 'customers',
                        icon: Icon(LucideIcons.users, size: 16),
                        label: Text('Клиенты'),
                      ),
                    ],
                    selected: {_tab},
                    onSelectionChanged: (selection) {
                      setState(() => _tab = selection.first);
                    },
                  ),
                ),
                const SizedBox(height: 10),
                TextField(
                  controller: _searchController,
                  onChanged: (v) => setState(() => _searchQuery = v),
                  decoration: InputDecoration(
                    hintText: _tab == 'customers'
                        ? 'Поиск клиентов: имя, телефон, город, ТК...'
                        : 'Поиск по запчасти, клиенту, треку, №...',
                    prefixIcon: const Icon(LucideIcons.search, size: 18),
                    suffixIcon: _searchQuery.isNotEmpty
                        ? IconButton(
                            icon: const Icon(LucideIcons.x, size: 16),
                            onPressed: () {
                              _searchController.clear();
                              setState(() => _searchQuery = '');
                            },
                          )
                        : null,
                    isDense: true,
                  ),
                ),
                if (_tab == 'customers') ...[
                  const SizedBox(height: 8),
                  SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: Row(
                      children: [
                        _categoryFilterChip('', 'Все'),
                        const SizedBox(width: 6),
                        _categoryFilterChip('regular', 'Обычные'),
                        const SizedBox(width: 6),
                        _categoryFilterChip('vip', '⭐ СТО / VIP'),
                        const SizedBox(width: 6),
                        _categoryFilterChip('wholesale', '🏢 Оптовики'),
                        const SizedBox(width: 6),
                        _categoryFilterChip('blacklist', '⚠️ ЧС'),
                      ],
                    ),
                  ),
                ],
              ],
            ),
          ),
          Expanded(
            child: _tab == 'customers'
                ? _buildCustomersView(context, ref)
                : ordersAsync.when(
                    loading: () => const OrderListSkeleton(),
                    error: (e, _) => AppEmptyState(
                      icon: LucideIcons.cloud_off,
                      title: 'Не удалось загрузить заказы',
                      message: 'Проверьте соединение и повторите попытку.',
                      actionLabel: 'Повторить',
                      onAction: _refresh,
                    ),
                    data: (data) {
                      final filtered = _filterOrders(data.orders);
                      return RefreshIndicator(
                        onRefresh: _refresh,
                        child: ListView.separated(
                          physics: const AlwaysScrollableScrollPhysics(),
                          padding: const EdgeInsets.fromLTRB(16, 8, 16, 20),
                          itemCount: filtered.length + (filtered.isEmpty ? 3 : 2),
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
                                title: _tab == 'active' ? 'В работе' : 'Завершённые продажи',
                                caption: '${filtered.length} заказов',
                              );
                            }
                            if (filtered.isEmpty) {
                              return AppEmptyState(
                                icon: LucideIcons.receipt,
                                title: _tab == 'active'
                                    ? 'Активных заказов нет'
                                    : 'История заказов пуста',
                                message: _tab == 'active'
                                    ? 'Выберите запчасть и создайте заказ.'
                                    : 'Завершённые продажи появятся здесь.',
                              );
                            }
                            return _OrderCard(
                              order: filtered[i - 2],
                              isCompleted: _tab == 'completed',
                              canChangeStatus: user?.isOperator == true,
                              onStatusChanged: () {
                                ref.invalidate(ordersProvider);
                                ref.invalidate(completedOrdersProvider);
                                ref.invalidate(inventoryProvider);
                              },
                            );
                          },
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildCustomersView(BuildContext context, WidgetRef ref) {
    final customersAsync = ref.watch(customersProvider);

    return customersAsync.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (e, _) => AppEmptyState(
        icon: LucideIcons.cloud_off,
        title: 'Не удалось загрузить клиентов',
        message: 'Проверьте соединение и повторите попытку.',
        actionLabel: 'Повторить',
        onAction: _refresh,
      ),
      data: (customers) {
        final filtered = _filterCustomers(customers);
        return RefreshIndicator(
          onRefresh: _refresh,
          child: ListView.separated(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 80),
            itemCount: filtered.isEmpty ? 1 : filtered.length,
            separatorBuilder: (_, _) => const SizedBox(height: 10),
            itemBuilder: (context, index) {
              if (filtered.isEmpty) {
                return AppEmptyState(
                  icon: LucideIcons.users,
                  title: 'Клиенты не найдены',
                  message: _searchQuery.isNotEmpty || _customerCategory.isNotEmpty
                      ? 'Попробуйте изменить параметры поиска или фильтра.'
                      : 'В базе пока нет клиентов. Создайте первого!',
                  actionLabel: 'Добавить клиента',
                  onAction: () => CustomerFormSheet.show(context),
                );
              }
              final customer = filtered[index];
              return _CustomerCard(
                customer: customer,
                onTap: () => CustomerDetailSheet.show(context, customer.id),
              );
            },
          ),
        );
      },
    );
  }
}

class _CustomerCard extends StatelessWidget {
  final Customer customer;
  final VoidCallback onTap;

  const _CustomerCard({required this.customer, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: customer.isBlacklist
              ? Colors.red.withAlpha(120)
              : Theme.of(context).colorScheme.outlineVariant.withAlpha(80),
          width: customer.isBlacklist ? 1.5 : 1.0,
        ),
      ),
      color: customer.isBlacklist ? Colors.red.withAlpha(15) : null,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(14.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  CircleAvatar(
                    backgroundColor: customer.categoryColor.withAlpha(30),
                    child: Icon(customer.categoryIcon, color: customer.categoryColor, size: 20),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          customer.name,
                          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          customer.phone,
                          style: TextStyle(
                            color: Theme.of(context).colorScheme.primary,
                            fontWeight: FontWeight.w500,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: customer.categoryColor.withAlpha(25),
                      borderRadius: BorderRadius.circular(6),
                      border: Border.all(color: customer.categoryColor.withAlpha(80)),
                    ),
                    child: Text(
                      customer.categoryLabel,
                      style: TextStyle(
                        color: customer.categoryColor,
                        fontWeight: FontWeight.bold,
                        fontSize: 11,
                      ),
                    ),
                  ),
                ],
              ),
              if (customer.isBlacklist && customer.notes.isNotEmpty) ...[
                const SizedBox(height: 8),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: Colors.red.withAlpha(30),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Row(
                    children: [
                      const Icon(LucideIcons.triangle_alert, size: 14, color: Colors.red),
                      const SizedBox(width: 6),
                      Expanded(
                        child: Text(
                          customer.notes,
                          style: const TextStyle(color: Colors.red, fontSize: 12),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
              const Divider(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Row(
                      children: [
                        if (customer.city.isNotEmpty) ...[
                          const Icon(LucideIcons.map_pin, size: 14, color: Colors.grey),
                          const SizedBox(width: 4),
                          Flexible(
                            child: Text(
                              customer.city,
                              style: const TextStyle(fontSize: 13, color: Colors.grey),
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                          const SizedBox(width: 8),
                        ],
                        if (customer.preferredTk.isNotEmpty) ...[
                          const Icon(LucideIcons.truck, size: 14, color: Colors.grey),
                          const SizedBox(width: 4),
                          Flexible(
                            child: Text(
                              customer.preferredTk,
                              style: const TextStyle(fontSize: 13, color: Colors.grey),
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                  Row(
                    children: [
                      if (customer.discountPercent > 0) ...[
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(
                            color: Colors.green.withAlpha(30),
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            '-${customer.discountPercent.toStringAsFixed(0)}%',
                            style: const TextStyle(
                              color: Colors.green,
                              fontWeight: FontWeight.bold,
                              fontSize: 12,
                            ),
                          ),
                        ),
                        const SizedBox(width: 8),
                      ],
                      Text(
                        '${customer.totalOrders} зак. · ${formatPrice(customer.totalSpent)}',
                        style: const TextStyle(
                          fontWeight: FontWeight.w600,
                          fontSize: 13,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _OrderCard extends StatelessWidget {
  final Order order;
  final bool isCompleted;
  final bool canChangeStatus;
  final VoidCallback onStatusChanged;

  const _OrderCard({
    required this.order,
    required this.isCompleted,
    required this.canChangeStatus,
    required this.onStatusChanged,
  });

  static const _sourceLabels = {
    'drom': 'Дром',
    'avito': 'Авито',
    'messenger': 'Мессенджер',
    'pickup': 'На месте',
  };

  static const _paymentLabels = {
    'unpaid': ('Не оплачен', Color(0xFFE53935)),
    'prepaid': ('Предоплата', Color(0xFFFB8C00)),
    'paid': ('Оплачен', Color(0xFF43A047)),
  };

  static const _warehouseLabels = {
    'inspecting': ('На проверке', Color(0xFF00ACC1)),
    'transfer': ('Перемещение', Color(0xFF5E35B1)),
    'ready': ('Собран', Color(0xFF43A047)),
  };

  static const _deliveryLabels = {
    'tk': ('ТК', Color(0xFF1E88E5)),
    'pickup': ('Самовывоз', Color(0xFF8E24AA)),
    'city': ('По городу', Color(0xFF039BE5)),
  };

  @override
  Widget build(BuildContext context) {
    final statusColor = Color(order.statusColor);
    final payInfo = _paymentLabels[order.paymentStatus] ?? ('Не оплачен', const Color(0xFFE53935));
    final whInfo = _warehouseLabels[order.warehouseStatus] ?? ('На проверке', const Color(0xFF00ACC1));
    final delInfo = _deliveryLabels[order.deliveryMethod] ?? ('ТК', const Color(0xFF1E88E5));

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
                        isCompleted ? 'Выдан / Завершён' : order.displayStatusText,
                        style: TextStyle(
                          color: statusColor,
                          fontSize: 11,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 6),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
                  decoration: BoxDecoration(
                    color: AppTheme.surfaceColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    _sourceLabels[order.source] ?? order.source,
                    style: const TextStyle(
                      fontSize: 10,
                      color: AppTheme.mutedColor,
                      fontWeight: FontWeight.w600,
                    ),
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
                const SizedBox(width: 8),
                Text(
                  isCompleted && order.completedAtFormatted.isNotEmpty
                      ? order.completedAtFormatted
                      : (order.timeAgo.isNotEmpty
                            ? order.timeAgo
                            : order.createdAtFormatted),
                  style: const TextStyle(
                    color: AppTheme.mutedColor,
                    fontSize: 11,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            if (order.items.length > 1)
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  for (final item in order.items)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 4),
                      child: _PartLinkRow(
                        label:
                            '${item.quantity} × ${item.partName.isNotEmpty ? item.partName : (order.partName.isNotEmpty ? order.partName : 'Запчасть #${item.partId}')}${item.price > 0 ? ' (${item.price.toStringAsFixed(0)} ₽)' : ''}',
                        partId: !isCompleted && item.partId > 0 ? item.partId : null,
                      ),
                    ),
                ],
              )
            else
              _PartLinkRow(
                label: '${order.totalQuantity} × ${order.partName}',
                partId: !isCompleted ? order.effectivePartId : null,
              ),
            if (order.totalAmount > 0) ...[
              const SizedBox(height: 6),
              Text(
                'Итого: ${order.totalAmount.toStringAsFixed(0)} ₽${order.discount > 0 ? ' (скидка ${order.discount.toStringAsFixed(0)} ₽)' : ''}',
                style: const TextStyle(
                  color: Color(0xFF43A047),
                  fontWeight: FontWeight.w800,
                  fontSize: 14,
                ),
              ),
            ],
            const SizedBox(height: 10),
            Wrap(
              spacing: 6,
              runSpacing: 6,
              children: [
                _chip(payInfo.$1, payInfo.$2),
                _chip(whInfo.$1, whInfo.$2),
                _chip(
                  order.transportCompany.isNotEmpty
                      ? '${delInfo.$1}: ${order.transportCompany}'
                      : delInfo.$1,
                  delInfo.$2,
                ),
              ],
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
                if (order.orderNumber.isNotEmpty)
                  _info(LucideIcons.hash, 'Сделка ${order.orderNumber}'),
                if (order.buyerNumber.isNotEmpty)
                  _info(LucideIcons.phone, order.buyerNumber),
                if (order.trackingNumber.isNotEmpty)
                  _info(LucideIcons.truck, 'Трек: ${order.trackingNumber}'),
              ],
            ),
            if (order.notes.isNotEmpty) ...[
              const SizedBox(height: 8),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                decoration: BoxDecoration(
                  color: AppTheme.surfaceColor,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  '💬 ${order.notes}',
                  style: const TextStyle(
                    fontSize: 12,
                    color: AppTheme.mutedColor,
                  ),
                ),
              ),
            ],
            if (!isCompleted && canChangeStatus) ...[
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed: () => _EditOrderSheet.show(
                    context,
                    order: order,
                    onChanged: onStatusChanged,
                  ),
                  icon: const Icon(LucideIcons.pencil, size: 15),
                  label: const Text('Статусы, трек и цена'),
                ),
              ),
              const SizedBox(height: 8),
              _OrderActions(orderId: order.id, onChanged: onStatusChanged),
            ],
          ],
        ),
      ),
    );
  }

  Widget _chip(String label, Color color) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
    decoration: BoxDecoration(
      color: color.withValues(alpha: 0.14),
      borderRadius: BorderRadius.circular(8),
      border: Border.all(color: color.withValues(alpha: 0.35)),
    ),
    child: Text(
      label,
      style: TextStyle(
        fontSize: 11,
        fontWeight: FontWeight.w600,
        color: color,
      ),
    ),
  );

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

class _PartLinkRow extends StatelessWidget {
  final String label;
  final int? partId;

  const _PartLinkRow({required this.label, this.partId});

  @override
  Widget build(BuildContext context) {
    final hasLink = partId != null && partId! > 0;
    if (!hasLink) {
      return Text(
        label,
        style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 15),
      );
    }

    return InkWell(
      borderRadius: BorderRadius.circular(6),
      onTap: () => context.push('/inventory/part/$partId'),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 2),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Flexible(
              child: Text(
                label,
                style: const TextStyle(
                  fontWeight: FontWeight.w700,
                  fontSize: 15,
                  color: AppTheme.primaryColor,
                ),
              ),
            ),
            const SizedBox(width: 6),
            const Icon(
              LucideIcons.external_link,
              size: 14,
              color: AppTheme.primaryColor,
            ),
          ],
        ),
      ),
    );
  }
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
                  'Завершить продажу по заказу #$orderId? Запчасти будут списаны со склада и учтены в статистике продаж.',
              actionLabel: 'Выдать / Завершить',
              path: ApiEndpoints.completeOrder(orderId),
              method: 'PUT',
            ),
            child: const Text('Выдать / Завершить'),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: OutlinedButton(
            onPressed: () => _confirmAction(
              context,
              title: 'Отмена заказа',
              description:
                  'Отменить заказ #$orderId? Запчасти останутся в наличии на складе.',
              actionLabel: 'Отменить заказ',
              path: ApiEndpoints.orderById(orderId),
              method: 'DELETE',
            ),
            style: OutlinedButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Отменить'),
          ),
        ),
      ],
    );
  }
}

class _EditOrderSheet extends StatefulWidget {
  final Order order;
  final VoidCallback onChanged;

  const _EditOrderSheet({required this.order, required this.onChanged});

  static Future<void> show(
    BuildContext context, {
    required Order order,
    required VoidCallback onChanged,
  }) async {
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (_) => _EditOrderSheet(order: order, onChanged: onChanged),
    );
  }

  @override
  State<_EditOrderSheet> createState() => _EditOrderSheetState();
}

class _EditOrderSheetState extends State<_EditOrderSheet> {
  late String _paymentStatus;
  late String _warehouseStatus;
  late String _deliveryMethod;
  late String _source;
  late final TextEditingController _buyerController;
  late final TextEditingController _orderNumberController;
  late final TextEditingController _tkController;
  late final TextEditingController _trackingController;
  late final TextEditingController _notesController;
  late final TextEditingController _discountController;
  final Map<int, TextEditingController> _qtyControllers = {};
  final Map<int, TextEditingController> _priceControllers = {};
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final o = widget.order;
    _paymentStatus = o.paymentStatus;
    _warehouseStatus = o.warehouseStatus;
    _deliveryMethod = o.deliveryMethod;
    _source = o.source;
    _buyerController = TextEditingController(text: o.buyerNumber);
    _orderNumberController = TextEditingController(text: o.orderNumber);
    _tkController = TextEditingController(text: o.transportCompany);
    _trackingController = TextEditingController(text: o.trackingNumber);
    _notesController = TextEditingController(text: o.notes);
    _discountController = TextEditingController(
      text: o.discount > 0 ? o.discount.toStringAsFixed(0) : '0',
    );
    for (final item in o.items) {
      _qtyControllers[item.id] = TextEditingController(
        text: item.quantity.toString(),
      );
      _priceControllers[item.id] = TextEditingController(
        text: item.price.toStringAsFixed(0),
      );
    }
  }

  @override
  void dispose() {
    _buyerController.dispose();
    _orderNumberController.dispose();
    _tkController.dispose();
    _trackingController.dispose();
    _notesController.dispose();
    _discountController.dispose();
    for (final c in _qtyControllers.values) {
      c.dispose();
    }
    for (final c in _priceControllers.values) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _save() async {
    if (_saving) return;
    setState(() => _saving = true);
    try {
      await apiClient.dio.patch(
        ApiEndpoints.orderById(widget.order.id),
        data: {
          'payment_status': _paymentStatus,
          'warehouse_status': _warehouseStatus,
          'delivery_method': _deliveryMethod,
          'source': _source,
          'buyer_number': _buyerController.text.trim(),
          'order_number': _orderNumberController.text.trim(),
          'transport_company': _tkController.text.trim(),
          'tracking_number': _trackingController.text.trim(),
          'notes': _notesController.text.trim(),
          'discount': double.tryParse(_discountController.text) ?? 0,
        },
      );

      for (final item in widget.order.items) {
        final newQty = int.tryParse(_qtyControllers[item.id]?.text ?? '') ?? item.quantity;
        final newPrice = double.tryParse(_priceControllers[item.id]?.text ?? '') ?? item.price;
        if (newQty != item.quantity || newPrice != item.price) {
          await apiClient.dio.patch(
            ApiEndpoints.orderItem(widget.order.id, item.id),
            data: {'quantity': newQty, 'price': newPrice},
          );
        }
      }

      if (!mounted) return;
      widget.onChanged();
      Navigator.pop(context);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Заказ #${widget.order.id} обновлён')),
      );
    } catch (error) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Ошибка обновления заказа: $error')),
      );
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final bottomInset = MediaQuery.viewInsetsOf(context).bottom;
    return Padding(
      padding: EdgeInsets.fromLTRB(20, 12, 20, bottomInset + 20),
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    'Заказ #${widget.order.id}',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                ),
                IconButton(
                  onPressed: () => Navigator.pop(context),
                  icon: const Icon(LucideIcons.x),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: DropdownButtonFormField<String>(
                    initialValue: _paymentStatus,
                    decoration: const InputDecoration(labelText: 'Оплата'),
                    items: const [
                      DropdownMenuItem(value: 'unpaid', child: Text('Не оплачен')),
                      DropdownMenuItem(value: 'prepaid', child: Text('Предоплата')),
                      DropdownMenuItem(value: 'paid', child: Text('Оплачен')),
                    ],
                    onChanged: (v) {
                      if (v != null) setState(() => _paymentStatus = v);
                    },
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: DropdownButtonFormField<String>(
                    initialValue: _warehouseStatus,
                    decoration: const InputDecoration(labelText: 'Склад'),
                    items: const [
                      DropdownMenuItem(value: 'inspecting', child: Text('Проверка')),
                      DropdownMenuItem(value: 'transfer', child: Text('Перемещение')),
                      DropdownMenuItem(value: 'ready', child: Text('Собран')),
                    ],
                    onChanged: (v) {
                      if (v != null) setState(() => _warehouseStatus = v);
                    },
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: DropdownButtonFormField<String>(
                    initialValue: _deliveryMethod,
                    decoration: const InputDecoration(labelText: 'Доставка'),
                    items: const [
                      DropdownMenuItem(value: 'tk', child: Text('ТК')),
                      DropdownMenuItem(value: 'pickup', child: Text('Самовывоз')),
                      DropdownMenuItem(value: 'city', child: Text('По городу')),
                    ],
                    onChanged: (v) {
                      if (v != null) setState(() => _deliveryMethod = v);
                    },
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: DropdownButtonFormField<String>(
                    initialValue: _source,
                    decoration: const InputDecoration(labelText: 'Площадка'),
                    items: const [
                      DropdownMenuItem(value: 'drom', child: Text('Дром')),
                      DropdownMenuItem(value: 'avito', child: Text('Авито')),
                      DropdownMenuItem(value: 'messenger', child: Text('Мессенджер')),
                      DropdownMenuItem(value: 'pickup', child: Text('На месте')),
                    ],
                    onChanged: (v) {
                      if (v != null) setState(() => _source = v);
                    },
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: TextFormField(
                    controller: _buyerController,
                    decoration: const InputDecoration(labelText: 'Контакт клиента'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextFormField(
                    controller: _orderNumberController,
                    decoration: const InputDecoration(labelText: '№ сделки'),
                  ),
                ),
              ],
            ),
            if (_deliveryMethod == 'tk') ...[
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: TextFormField(
                      controller: _tkController,
                      decoration: const InputDecoration(labelText: 'ТК (СДЭК, Энергия...)'),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: TextFormField(
                      controller: _trackingController,
                      decoration: const InputDecoration(labelText: 'Трек-номер'),
                    ),
                  ),
                ],
              ),
            ],
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  flex: 2,
                  child: TextFormField(
                    controller: _notesController,
                    decoration: const InputDecoration(labelText: 'Примечание'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextFormField(
                    controller: _discountController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Скидка (₽)'),
                  ),
                ),
              ],
            ),
            if (widget.order.items.isNotEmpty) ...[
              const SizedBox(height: 16),
              const Text(
                'Позиции заказа',
                style: TextStyle(fontWeight: FontWeight.w700, fontSize: 13),
              ),
              const SizedBox(height: 8),
              for (final item in widget.order.items)
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Row(
                    children: [
                      Expanded(
                        flex: 2,
                        child: Text(
                          item.partName.isNotEmpty
                              ? item.partName
                              : 'Запчасть #${item.partId}',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(fontSize: 13),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: TextFormField(
                          controller: _qtyControllers[item.id],
                          keyboardType: TextInputType.number,
                          inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                          decoration: const InputDecoration(
                            labelText: 'Кол-во',
                            isDense: true,
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: TextFormField(
                          controller: _priceControllers[item.id],
                          keyboardType: TextInputType.number,
                          decoration: const InputDecoration(
                            labelText: 'Цена ₽',
                            isDense: true,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
            ],
            const SizedBox(height: 16),
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: _saving ? null : _save,
                icon: _saving
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(LucideIcons.check),
                label: const Text('Сохранить изменения'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
