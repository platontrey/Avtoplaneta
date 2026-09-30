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
    ref.invalidate(completedOrdersProvider);
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
  final _buyerNumberController = TextEditingController();
  final _orderNumberController = TextEditingController();
  final _transportCompanyController = TextEditingController();
  final _notesController = TextEditingController();
  final Map<int, TextEditingController> _quantityControllers = {};
  final Map<int, TextEditingController> _priceControllers = {};
  String _mode = 'order'; // 'order', 'quick', 'existing'
  String _source = 'drom';
  String _paymentStatus = 'unpaid';
  String _deliveryMethod = 'tk';
  bool _submitting = false;
  int? _selectedOrderId;

  @override
  void initState() {
    super.initState();
    for (final part in widget.parts) {
      _quantityControllers[part.id] = TextEditingController(text: '1');
      _priceControllers[part.id] = TextEditingController(
        text: part.price > 0 ? part.price.toStringAsFixed(0) : '',
      );
    }
  }

  @override
  void dispose() {
    _buyerNumberController.dispose();
    _orderNumberController.dispose();
    _transportCompanyController.dispose();
    _notesController.dispose();
    for (final controller in _quantityControllers.values) {
      controller.dispose();
    }
    for (final controller in _priceControllers.values) {
      controller.dispose();
    }
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate() || _submitting) {
      return;
    }
    if (_mode == 'existing' && _selectedOrderId == null) {
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
            'price': double.tryParse(_priceControllers[part.id]!.text) ?? part.price,
          },
        )
        .toList();
    setState(() => _submitting = true);
    try {
      if (_mode == 'existing') {
        for (final item in items) {
          await apiClient.dio.post(
            '/orders/$_selectedOrderId/items',
            data: item,
          );
        }
      } else {
        final isQuick = _mode == 'quick';
        final buyer = _buyerNumberController.text.trim();
        await apiClient.dio.post(
          '/orders',
          data: {
            'customer_id': 0,
            'order_number': _orderNumberController.text.trim(),
            'part': widget.parts.map((part) => part.name).join(', '),
            'part_id': widget.parts.first.id,
            'buyer_number': buyer.isNotEmpty
                ? buyer
                : (isQuick ? 'Продажа на месте' : 'Без контакта'),
            'source': isQuick ? 'pickup' : _source,
            'payment_status': isQuick ? 'paid' : _paymentStatus,
            'warehouse_status': isQuick ? 'ready' : 'inspecting',
            'delivery_method': isQuick ? 'pickup' : _deliveryMethod,
            'transport_company': isQuick
                ? ''
                : _transportCompanyController.text.trim(),
            'notes': _notesController.text.trim(),
            'quick_sale': isQuick,
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
              const SizedBox(height: 16),
              SizedBox(
                width: double.infinity,
                child: SegmentedButton<String>(
                  segments: const [
                    ButtonSegment(value: 'order', label: Text('В заказ')),
                    ButtonSegment(value: 'quick', label: Text('На месте')),
                    ButtonSegment(value: 'existing', label: Text('К заказу')),
                  ],
                  selected: {_mode},
                  onSelectionChanged: (selection) {
                    setState(() => _mode = selection.first);
                  },
                ),
              ),
              const SizedBox(height: 16),
              if (_mode == 'existing')
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
                  controller: _buyerNumberController,
                  decoration: InputDecoration(
                    labelText: _mode == 'quick'
                        ? 'Контакт покупателя (необязательно)'
                        : 'Контакт покупателя *',
                  ),
                  validator: (value) {
                    if (_mode == 'quick') return null;
                    return value == null || value.trim().isEmpty
                        ? 'Укажите контакт покупателя'
                        : null;
                  },
                ),
                const SizedBox(height: 12),
                if (_mode == 'order') ...[
                  Row(
                    children: [
                      Expanded(
                        child: DropdownButtonFormField<String>(
                          initialValue: _source,
                          decoration: const InputDecoration(
                            labelText: 'Площадка',
                          ),
                          items: const [
                            DropdownMenuItem(value: 'drom', child: Text('Дром')),
                            DropdownMenuItem(value: 'avito', child: Text('Авито')),
                            DropdownMenuItem(value: 'messenger', child: Text('Мессенджер')),
                            DropdownMenuItem(value: 'pickup', child: Text('Самовывоз')),
                          ],
                          onChanged: (v) {
                            if (v != null) setState(() => _source = v);
                          },
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: TextFormField(
                          controller: _orderNumberController,
                          decoration: const InputDecoration(
                            labelText: '№ сделки',
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: DropdownButtonFormField<String>(
                          initialValue: _paymentStatus,
                          decoration: const InputDecoration(
                            labelText: 'Оплата',
                          ),
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
                          initialValue: _deliveryMethod,
                          decoration: const InputDecoration(
                            labelText: 'Доставка',
                          ),
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
                    ],
                  ),
                  if (_deliveryMethod == 'tk') ...[
                    const SizedBox(height: 12),
                    TextFormField(
                      controller: _transportCompanyController,
                      decoration: const InputDecoration(
                        labelText: 'Транспортная компания (СДЭК, Энергия...)',
                      ),
                    ),
                  ],
                ],
                const SizedBox(height: 12),
                TextFormField(
                  controller: _notesController,
                  decoration: const InputDecoration(
                    labelText: 'Примечание (торг, упаковка...)',
                  ),
                ),
              ],
              const SizedBox(height: 12),
              ...widget.parts.map(
                (part) => Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextFormField(
                          controller: _quantityControllers[part.id],
                          keyboardType: TextInputType.number,
                          inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                          decoration: InputDecoration(
                            labelText: widget.parts.length == 1
                                ? 'Кол-во *'
                                : '${part.name} (${part.quantity} шт.)',
                          ),
                          validator: (value) {
                            final quantity = int.tryParse(value ?? '');
                            if (quantity == null || quantity <= 0) {
                              return 'Укажите кол-во';
                            }
                            if (part.quantity > 0 && quantity > part.quantity) {
                              return 'Макс. ${part.quantity} шт.';
                            }
                            return null;
                          },
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: TextFormField(
                          controller: _priceControllers[part.id],
                          keyboardType: TextInputType.number,
                          decoration: const InputDecoration(
                            labelText: 'Цена за шт. (₽)',
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
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
                      : Icon(
                          _mode == 'quick'
                              ? LucideIcons.zap
                              : LucideIcons.shopping_cart,
                        ),
                  label: Text(
                    _mode == 'existing'
                        ? 'Добавить в заказ'
                        : _mode == 'quick'
                        ? 'Продать и списать'
                        : 'Создать заказ',
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

