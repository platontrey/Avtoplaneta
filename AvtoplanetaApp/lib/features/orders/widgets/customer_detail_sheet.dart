import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/customer.dart';
import '../../../core/utils/formatters.dart';
import '../providers/customers_provider.dart';
import 'customer_form_sheet.dart';

class CustomerDetailSheet extends ConsumerWidget {
  final int customerId;

  const CustomerDetailSheet({super.key, required this.customerId});

  static Future<void> show(BuildContext context, int customerId) {
    return showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (_) => CustomerDetailSheet(customerId: customerId),
    );
  }

  void _copyTkDetails(BuildContext context, Customer customer) {
    final text = [
      'Получатель: ${customer.name}',
      'Телефон: ${customer.phone}',
      if (customer.city.isNotEmpty) 'Город: ${customer.city}',
      if (customer.preferredTk.isNotEmpty) 'ТК: ${customer.preferredTk}',
      if (customer.tkDetails.isNotEmpty) 'Данные ТК: ${customer.tkDetails}',
    ].join('\n');

    Clipboard.setData(ClipboardData(text: text));
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Реквизиты для ТК скопированы в буфер обмена'),
        duration: Duration(seconds: 2),
      ),
    );
  }

  Future<void> _confirmDelete(BuildContext context, WidgetRef ref, Customer customer) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить клиента?'),
        content: Text(
          'Вы действительно хотите удалить клиента "${customer.name}"? '
          'История его заказов сохранится в системе.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Отмена'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Удалить'),
          ),
        ],
      ),
    );

    if (confirmed == true && context.mounted) {
      try {
        await CustomerService.deleteCustomer(customer.id);
        ref.invalidate(customersProvider);
        if (context.mounted) {
          Navigator.pop(context);
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Клиент удален')),
          );
        }
      } catch (e) {
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Ошибка удаления: $e')),
          );
        }
      }
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detailsAsync = ref.watch(customerDetailsProvider(customerId));

    return DraggableScrollableSheet(
      initialChildSize: 0.85,
      minChildSize: 0.5,
      maxChildSize: 0.95,
      expand: false,
      builder: (context, scrollController) {
        return detailsAsync.when(
          loading: () => const Center(
            child: Padding(
              padding: EdgeInsets.all(40.0),
              child: CircularProgressIndicator(),
            ),
          ),
          error: (err, _) => Center(
            child: Padding(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(LucideIcons.triangle_alert, color: Colors.red, size: 40),
                  const SizedBox(height: 12),
                  Text('Ошибка загрузки: $err', textAlign: TextAlign.center),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => ref.invalidate(customerDetailsProvider(customerId)),
                    child: const Text('Повторить'),
                  ),
                ],
              ),
            ),
          ),
          data: (details) {
            final customer = details.customer;
            final orders = details.orders;

            return ListView(
              controller: scrollController,
              padding: const EdgeInsets.fromLTRB(20, 12, 20, 32),
              children: [
                // Индикатор перетаскивания
                Center(
                  child: Container(
                    width: 40,
                    height: 4,
                    margin: const EdgeInsets.only(bottom: 16),
                    decoration: BoxDecoration(
                      color: Colors.grey.withAlpha(100),
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                ),

                // Заголовок и бейдж категории
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            customer.name,
                            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                                  fontWeight: FontWeight.bold,
                                ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            customer.phone,
                            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                  color: Theme.of(context).colorScheme.primary,
                                ),
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      icon: const Icon(LucideIcons.user_pen),
                      tooltip: 'Редактировать',
                      onPressed: () async {
                        final updated = await CustomerFormSheet.show(
                          context,
                          customer: customer,
                        );
                        if (updated == true) {
                          ref.invalidate(customerDetailsProvider(customerId));
                        }
                      },
                    ),
                    IconButton(
                      icon: const Icon(LucideIcons.trash, color: Colors.red),
                      tooltip: 'Удалить',
                      onPressed: () => _confirmDelete(context, ref, customer),
                    ),
                  ],
                ),
                const SizedBox(height: 10),

                // Категория и Скидка бейджи
                Wrap(
                  spacing: 8,
                  runSpacing: 6,
                  children: [
                    Chip(
                      avatar: Icon(customer.categoryIcon, size: 16, color: customer.categoryColor),
                      label: Text(
                        customer.categoryLabel,
                        style: TextStyle(
                          color: customer.categoryColor,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      backgroundColor: customer.categoryColor.withAlpha(25),
                      side: BorderSide(color: customer.categoryColor.withAlpha(60)),
                    ),
                    if (customer.discountPercent > 0)
                      Chip(
                        avatar: const Icon(LucideIcons.percent, size: 15, color: Colors.green),
                        label: Text(
                          'Скидка ${customer.discountPercent.toStringAsFixed(0)}%',
                          style: const TextStyle(color: Colors.green, fontWeight: FontWeight.w600),
                        ),
                        backgroundColor: Colors.green.withAlpha(25),
                        side: BorderSide(color: Colors.green.withAlpha(60)),
                      ),
                  ],
                ),

                if (customer.isBlacklist) ...[
                  const SizedBox(height: 12),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.red.withAlpha(25),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(color: Colors.red.withAlpha(80)),
                    ),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Icon(LucideIcons.triangle_alert, color: Colors.red, size: 22),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                'Клиент находится в Черном Списке!',
                                style: TextStyle(
                                  color: Colors.red,
                                  fontWeight: FontWeight.bold,
                                  fontSize: 14,
                                ),
                              ),
                              if (customer.notes.isNotEmpty) ...[
                                const SizedBox(height: 4),
                                Text(
                                  customer.notes,
                                  style: TextStyle(
                                    color: Colors.red.shade900,
                                    fontSize: 13,
                                  ),
                                ),
                              ],
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ],

                const SizedBox(height: 16),

                // Карточка статистики
                Row(
                  children: [
                    Expanded(
                      child: Container(
                        padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 16),
                        decoration: BoxDecoration(
                          color: Theme.of(context).colorScheme.surfaceContainerHighest.withAlpha(80),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Заказов всего',
                              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: Colors.grey,
                                  ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              '${customer.totalOrders}',
                              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                                    fontWeight: FontWeight.bold,
                                  ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Container(
                        padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 16),
                        decoration: BoxDecoration(
                          color: Theme.of(context).colorScheme.surfaceContainerHighest.withAlpha(80),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Сумма покупок',
                              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: Colors.grey,
                                  ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              formatPrice(customer.totalSpent),
                              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                                    fontWeight: FontWeight.bold,
                                    color: Colors.green.shade700,
                                  ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 16),

                // Карточка реквизитов доставки ТК
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Theme.of(context).colorScheme.surfaceContainerHighest.withAlpha(80),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Row(
                            children: [
                              Icon(LucideIcons.truck, size: 20, color: Theme.of(context).colorScheme.primary),
                              const SizedBox(width: 8),
                              const Text(
                                'Доставка и ТК',
                                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                              ),
                            ],
                          ),
                          OutlinedButton.icon(
                            style: OutlinedButton.styleFrom(
                              visualDensity: VisualDensity.compact,
                              padding: const EdgeInsets.symmetric(horizontal: 10),
                            ),
                            onPressed: () => _copyTkDetails(context, customer),
                            icon: const Icon(LucideIcons.copy, size: 14),
                            label: const Text('Скопировать'),
                          ),
                        ],
                      ),
                      const SizedBox(height: 10),
                      if (customer.city.isNotEmpty) ...[
                        Row(
                          children: [
                            const Icon(LucideIcons.map_pin, size: 16, color: Colors.grey),
                            const SizedBox(width: 6),
                            Text(customer.city, style: const TextStyle(fontWeight: FontWeight.w500)),
                          ],
                        ),
                        const SizedBox(height: 6),
                      ],
                      if (customer.preferredTk.isNotEmpty) ...[
                        Row(
                          children: [
                            const Icon(LucideIcons.package, size: 16, color: Colors.grey),
                            const SizedBox(width: 6),
                            Text('ТК: ${customer.preferredTk}', style: const TextStyle(fontWeight: FontWeight.w500)),
                          ],
                        ),
                        const SizedBox(height: 6),
                      ],
                      if (customer.tkDetails.isNotEmpty) ...[
                        const SizedBox(height: 4),
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.all(10),
                          decoration: BoxDecoration(
                            color: Theme.of(context).colorScheme.surface,
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(color: Colors.grey.withAlpha(50)),
                          ),
                          child: SelectableText(
                            customer.tkDetails,
                            style: const TextStyle(fontSize: 13, fontFamily: 'monospace'),
                          ),
                        ),
                      ] else if (customer.city.isEmpty && customer.preferredTk.isEmpty) ...[
                        const Text(
                          'Данные доставки не указаны',
                          style: TextStyle(color: Colors.grey, fontStyle: FontStyle.italic),
                        ),
                      ],
                    ],
                  ),
                ),

                if (customer.notes.isNotEmpty && !customer.isBlacklist) ...[
                  const SizedBox(height: 16),
                  Container(
                    padding: const EdgeInsets.all(14),
                    decoration: BoxDecoration(
                      color: Theme.of(context).colorScheme.surfaceContainerHighest.withAlpha(50),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Row(
                          children: [
                            Icon(LucideIcons.notebook_pen, size: 18, color: Colors.grey),
                            SizedBox(width: 8),
                            Text('Заметки', style: TextStyle(fontWeight: FontWeight.w600)),
                          ],
                        ),
                        const SizedBox(height: 6),
                        Text(customer.notes, style: const TextStyle(fontSize: 14)),
                      ],
                    ),
                  ),
                ],

                const SizedBox(height: 24),

                // История заказов
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'История заказов (${orders.length})',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                  ],
                ),
                const SizedBox(height: 10),

                if (orders.isEmpty)
                  Container(
                    padding: const EdgeInsets.all(24),
                    alignment: Alignment.center,
                    child: const Text(
                      'У клиента пока нет оформленных заказов',
                      style: TextStyle(color: Colors.grey),
                    ),
                  )
                else
                  ...orders.map((order) {
                    return Card(
                      margin: const EdgeInsets.only(bottom: 8),
                      elevation: 0,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                        side: BorderSide(
                          color: Theme.of(context).colorScheme.outlineVariant.withAlpha(60),
                        ),
                      ),
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  order.orderNumber.isNotEmpty
                                      ? '№${order.orderNumber}'
                                      : 'Заказ #${order.id}',
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold,
                                    fontSize: 14,
                                  ),
                                ),
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                                  decoration: BoxDecoration(
                                    color: Color(order.statusColor).withAlpha(30),
                                    borderRadius: BorderRadius.circular(6),
                                    border: Border.all(color: Color(order.statusColor).withAlpha(100)),
                                  ),
                                  child: Text(
                                    order.displayStatusText,
                                    style: TextStyle(
                                      color: Color(order.statusColor),
                                      fontSize: 12,
                                      fontWeight: FontWeight.w500,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 6),
                            Text(
                              order.partName,
                              style: const TextStyle(fontWeight: FontWeight.w500, fontSize: 13),
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                            ),
                            const SizedBox(height: 8),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  order.createdAtFormatted.isNotEmpty
                                      ? order.createdAtFormatted
                                      : order.timeAgo,
                                  style: const TextStyle(color: Colors.grey, fontSize: 12),
                                ),
                                Text(
                                  formatPrice(order.totalAmount),
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold,
                                    color: Colors.green,
                                    fontSize: 13,
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    );
                  }),
              ],
            );
          },
        );
      },
    );
  }
}
