import 'package:flutter/material.dart';
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
  final created = await showModalBottomSheet<bool>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    builder: (_) => _PartOrderSheet(part: part),
  );
  if (created == true) {
    ref.invalidate(ordersProvider);
    ref.invalidate(inventoryProvider);
    ref.invalidate(partProvider(part.id));
  }
  return created ?? false;
}

class _PartOrderSheet extends ConsumerStatefulWidget {
  final Part part;

  const _PartOrderSheet({required this.part});

  @override
  ConsumerState<_PartOrderSheet> createState() => _PartOrderSheetState();
}

class _PartOrderSheetState extends ConsumerState<_PartOrderSheet> {
  final _formKey = GlobalKey<FormState>();
  final _customerIdController = TextEditingController();
  final _buyerNumberController = TextEditingController();
  final _quantityController = TextEditingController(text: '1');
  bool _addToExisting = false;
  bool _submitting = false;
  int? _selectedOrderId;

  @override
  void dispose() {
    _customerIdController.dispose();
    _buyerNumberController.dispose();
    _quantityController.dispose();
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

    final quantity = int.parse(_quantityController.text);
    setState(() => _submitting = true);
    try {
      if (_addToExisting) {
        await apiClient.dio.post(
          '/orders/$_selectedOrderId/items',
          data: {'part_id': widget.part.id, 'quantity': quantity},
        );
      } else {
        await apiClient.dio.post(
          '/orders',
          data: {
            'customer_id': int.parse(_customerIdController.text),
            'part': widget.part.name,
            'part_id': widget.part.id,
            'buyer_number': _buyerNumberController.text.trim(),
            'items': [
              {'part_id': widget.part.id, 'quantity': quantity},
            ],
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
                      'Оформить заказ',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                  ),
                  IconButton(
                    tooltip: 'Закрыть',
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
              Text(
                '${widget.part.name} · ${widget.part.quantity} шт. в наличии',
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
              TextFormField(
                controller: _quantityController,
                keyboardType: TextInputType.number,
                inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                decoration: const InputDecoration(labelText: 'Количество *'),
                validator: (value) {
                  final quantity = int.tryParse(value ?? '');
                  if (quantity == null || quantity <= 0) {
                    return 'Укажите количество';
                  }
                  if (quantity > widget.part.quantity) {
                    return 'Доступно только ${widget.part.quantity} шт.';
                  }
                  return null;
                },
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
                      : const Icon(Icons.shopping_cart_checkout_rounded),
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
