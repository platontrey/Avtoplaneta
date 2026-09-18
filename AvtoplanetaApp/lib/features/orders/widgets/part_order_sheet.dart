import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/part.dart';
import '../../inventory/providers/inventory_provider.dart';
import '../providers/orders_provider.dart';

Future<bool> showPartOrderSheet(
  BuildContext context,
  WidgetRef ref,
  Part part,
) async {
  return showPartsOrderSheet(context, ref, [part]);
}

Future<bool> showPartsOrderSheet(
  BuildContext context,
  WidgetRef ref,
  List<Part> parts,
) async {
  if (parts.isEmpty) return false;
  final created = await showModalBottomSheet<bool>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    builder: (_) => _PartOrderSheet(parts: parts),
  );
  if (created == true) {
    ref.invalidate(ordersProvider);
    ref.invalidate(inventoryProvider);
    for (final part in parts) {
      ref.invalidate(partProvider(part.id));
    }
  }
  return created ?? false;
}

class _PartOrderSheet extends ConsumerStatefulWidget {
  final List<Part> parts;

  const _PartOrderSheet({required this.parts});

  @override
  ConsumerState<_PartOrderSheet> createState() => _PartOrderSheetState();
}

class _PartOrderSheetState extends ConsumerState<_PartOrderSheet> {
  final _formKey = GlobalKey<FormState>();
  final _customerIdController = TextEditingController();
  final _buyerNumberController = TextEditingController();
  final Map<int, TextEditingController> _quantityControllers = {};
  bool _addToExisting = false;
  bool _submitting = false;
  int? _selectedOrderId;

  @override
  void initState() {
    super.initState();
    for (final part in widget.parts) {
      _quantityControllers[part.id] = TextEditingController(text: '1');
    }
  }

  @override
  void dispose() {
    _customerIdController.dispose();
    _buyerNumberController.dispose();
    for (final controller in _quantityControllers.values) {
      controller.dispose();
    }
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate() || _submitting) {
      return;
    }
    if (_addToExisting && _selectedOrderId == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Выберите заказ')));
      return;
    }

    final items = widget.parts
        .map(
          (part) => {
            'part_id': part.id,
            'quantity': int.parse(_quantityControllers[part.id]!.text),
          },
        )
        .toList();
    setState(() => _submitting = true);
    try {
      if (_addToExisting) {
        for (final item in items) {
          await apiClient.dio.post(
            '/orders/$_selectedOrderId/items',
            data: item,
          );
        }
      } else {
        await apiClient.dio.post(
          '/orders',
          data: {
            'customer_id': int.parse(_customerIdController.text),
            'order_number': '',
            'part': widget.parts.map((part) => part.name).join(', '),
            'part_id': widget.parts.first.id,
            'buyer_number': _buyerNumberController.text.trim(),
            'items': items,
          },
        );
      }
      if (mounted) {
        Navigator.pop(context, true);
      }
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось оформить заказ: $error')),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _submitting = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final ordersAsync = ref.watch(ordersProvider);
    final bottomInset = MediaQuery.viewInsetsOf(context).bottom;

    return Padding(
      padding: EdgeInsets.fromLTRB(20, 12, 20, bottomInset + 20),
      child: SingleChildScrollView(
        child: Form(
          key: _formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      widget.parts.length == 1
                          ? 'Оформить заказ'
                          : 'Заказ из ${widget.parts.length} запчастей',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                  ),
                  IconButton(
                    tooltip: 'Закрыть',
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(LucideIcons.x),
                  ),
                ],
              ),
              if (widget.parts.length == 1)
                Text(
                  '${widget.parts.first.name} · ${widget.parts.first.quantity} шт. в наличии',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              const SizedBox(height: 18),
              SegmentedButton<bool>(
                segments: const [
                  ButtonSegment(value: false, label: Text('Новый заказ')),
                  ButtonSegment(value: true, label: Text('В существующий')),
                ],
                selected: {_addToExisting},
                onSelectionChanged: (selection) {
                  setState(() => _addToExisting = selection.first);
                },
              ),
              const SizedBox(height: 16),
              if (_addToExisting)
                ordersAsync.when(
                  loading: () => const Center(
                    child: Padding(
                      padding: EdgeInsets.all(16),
                      child: CircularProgressIndicator(),
                    ),
                  ),
                  error: (error, _) => Row(
                    children: [
                      const Expanded(
                        child: Text('Не удалось загрузить заказы'),
                      ),
                      TextButton(
                        onPressed: () => ref.invalidate(ordersProvider),
                        child: const Text('Повторить'),
                      ),
                    ],
                  ),
                  data: (data) => DropdownButtonFormField<int>(
                    initialValue: _selectedOrderId,
                    isExpanded: true,
                    decoration: const InputDecoration(labelText: 'Заказ *'),
                    items: data.orders
                        .map(
                          (order) => DropdownMenuItem(
                            value: order.id,
                            child: Text(
                              '#${order.id} · ${order.buyerNumber} · ${order.partName}',
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        )
                        .toList(),
                    onChanged: (value) {
                      setState(() => _selectedOrderId = value);
                    },
                  ),
                )
              else ...[
                TextFormField(
                  controller: _customerIdController,
                  keyboardType: TextInputType.number,
                  inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                  decoration: const InputDecoration(labelText: 'ID клиента *'),
                  validator: (value) {
                    final id = int.tryParse(value ?? '');
                    return id == null || id <= 0 ? 'Укажите ID клиента' : null;
                  },
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: _buyerNumberController,
                  decoration: const InputDecoration(
                    labelText: 'Номер покупателя *',
                  ),
                  validator: (value) => value == null || value.trim().isEmpty
                      ? 'Укажите номер покупателя'
                      : null,
                ),
              ],
              const SizedBox(height: 12),
              ...widget.parts.map(
                (part) => Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: TextFormField(
                    controller: _quantityControllers[part.id],
                    keyboardType: TextInputType.number,
                    inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                    decoration: InputDecoration(
                      labelText: widget.parts.length == 1
                          ? 'Количество *'
                          : '${part.name} (${part.quantity} шт.)',
                    ),
                    validator: (value) {
                      final quantity = int.tryParse(value ?? '');
                      if (quantity == null || quantity <= 0) {
                        return 'Укажите количество';
                      }
                      if (quantity > part.quantity) {
                        return 'Доступно только ${part.quantity} шт.';
                      }
                      return null;
                    },
                  ),
                ),
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: FilledButton.icon(
                  onPressed: _submitting ? null : _submit,
                  icon: _submitting
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: Colors.white,
                          ),
                        )
                      : const Icon(LucideIcons.shopping_cart),
                  label: Text(
                    _addToExisting ? 'Добавить в заказ' : 'Создать заказ',
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
