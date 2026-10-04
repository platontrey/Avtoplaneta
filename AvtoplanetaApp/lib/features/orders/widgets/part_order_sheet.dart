import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/customer.dart';
import '../../../core/models/part.dart';
import '../../inventory/providers/inventory_provider.dart';
import '../providers/customers_provider.dart';
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
    ref.invalidate(customersProvider);
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
  final _discountController = TextEditingController(text: '0');
  final Map<int, TextEditingController> _quantityControllers = {};
  final Map<int, TextEditingController> _priceControllers = {};
  String _mode = 'order'; // 'order', 'quick', 'existing'
  String _source = 'drom';
  String _paymentStatus = 'unpaid';
  String _deliveryMethod = 'tk';
  bool _submitting = false;
  int? _selectedOrderId;
  Customer? _selectedCustomer;

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
    _discountController.dispose();
    for (final controller in _quantityControllers.values) {
      controller.dispose();
    }
    for (final controller in _priceControllers.values) {
      controller.dispose();
    }
    super.dispose();
  }

  void _onCustomerSelected(Customer customer) {
    setState(() {
      _selectedCustomer = customer;
      _buyerNumberController.text = customer.phone;
      if (customer.preferredTk.isNotEmpty) {
        _transportCompanyController.text = customer.preferredTk;
      }
      if (customer.discountPercent > 0) {
        _discountController.text = customer.discountPercent.toStringAsFixed(0);
        for (final part in widget.parts) {
          final discounted = part.price * (1 - customer.discountPercent / 100);
          _priceControllers[part.id]?.text = discounted.toStringAsFixed(0);
        }
      }
    });
  }

  void _clearSelectedCustomer() {
    setState(() {
      _selectedCustomer = null;
    });
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
            '/api/v1/orders/$_selectedOrderId/items',
            data: item,
          );
        }
      } else {
        final isQuick = _mode == 'quick';
        final rawBuyer = _buyerNumberController.text.trim();
        final buyerDigits = rawBuyer.replaceAll(RegExp(r'\D'), '');

        // Нормализация контакта покупателя (очистка от тире, скобок, пробелов в канонический +7XXXXXXXXXX)
        String buyer = rawBuyer;
        if (buyerDigits.length == 10) {
          buyer = '+7$buyerDigits';
        } else if (buyerDigits.length == 11 &&
            (buyerDigits.startsWith('7') || buyerDigits.startsWith('8'))) {
          buyer = '+7${buyerDigits.substring(1)}';
        } else if (buyerDigits.length > 11) {
          buyer = '+$buyerDigits';
        }

        // Если клиент не был явно выбран из выпадающего списка, ищем по цифрам номера
        Customer? matchedCustomer = _selectedCustomer;
        if (matchedCustomer == null && buyerDigits.length >= 10) {
          final allCustomers =
              ref.read(customersProvider).valueOrNull ?? const <Customer>[];
          final canonicalDigits = buyerDigits.length == 11 &&
                  (buyerDigits.startsWith('7') || buyerDigits.startsWith('8'))
              ? buyerDigits.substring(1)
              : (buyerDigits.length == 10 ? buyerDigits : '');
          for (final c in allCustomers) {
            final cDigits = c.phone.replaceAll(RegExp(r'\D'), '');
            if (cDigits.isNotEmpty &&
                (cDigits == buyerDigits ||
                    (canonicalDigits.isNotEmpty &&
                        cDigits.endsWith(canonicalDigits)))) {
              matchedCustomer = c;
              break;
            }
          }
        }

        final discount = double.tryParse(_discountController.text) ?? 0.0;
        await apiClient.dio.post(
          '/api/v1/orders',
          data: {
            'customer_id': matchedCustomer?.id ?? 0,
            'discount': discount,
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
        HapticFeedback.mediumImpact();
        Navigator.pop(context, true);
      }
    } catch (error) {
      if (mounted) {
        String message = error.toString();
        if (error is DioException && error.response?.data != null) {
          final data = error.response!.data;
          if (data is Map) {
            final details = data['details'] ?? data['error'] ?? data['message'];
            if (details != null && details.toString().isNotEmpty) {
              message = details.toString();
            }
          } else if (data is String && data.isNotEmpty) {
            message = data;
          }
        }
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Не удалось оформить заказ: $message'),
            backgroundColor: Colors.red.shade800,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _submitting = false);
      }
    }
  }

  Widget _buildSelectedCustomerCard(Customer customer) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: customer.isBlacklist
            ? Colors.red.withAlpha(20)
            : customer.categoryColor.withAlpha(20),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(
          color: customer.isBlacklist
              ? Colors.red.withAlpha(120)
              : customer.categoryColor.withAlpha(80),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(customer.categoryIcon, size: 16, color: customer.categoryColor),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  '${customer.name} (${customer.categoryLabel})',
                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (customer.discountPercent > 0)
                Container(
                  margin: const EdgeInsets.only(right: 6),
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
                      fontSize: 11,
                    ),
                  ),
                ),
              InkWell(
                onTap: _clearSelectedCustomer,
                borderRadius: BorderRadius.circular(12),
                child: const Padding(
                  padding: EdgeInsets.all(4.0),
                  child: Icon(LucideIcons.x, size: 16),
                ),
              ),
            ],
          ),
          if (customer.isBlacklist) ...[
            const SizedBox(height: 6),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(LucideIcons.triangle_alert, size: 15, color: Colors.red),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    '⚠️ ВНИМАНИЕ: Клиент в черном списке! ${customer.notes}',
                    style: const TextStyle(
                      color: Colors.red,
                      fontWeight: FontWeight.bold,
                      fontSize: 12,
                    ),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildBuyerContactField(List<Customer> allCustomers) {
    if (_selectedCustomer != null) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildSelectedCustomerCard(_selectedCustomer!),
          TextFormField(
            controller: _buyerNumberController,
            decoration: const InputDecoration(
              labelText: 'Контакт покупателя *',
              prefixIcon: Icon(LucideIcons.phone),
            ),
            validator: (value) {
              if (_mode == 'quick') return null;
              return value == null || value.trim().isEmpty
                  ? 'Укажите контакт покупателя'
                  : null;
            },
          ),
        ],
      );
    }

    return LayoutBuilder(
      builder: (context, constraints) {
        return Autocomplete<Customer>(
          optionsBuilder: (textEditingValue) {
            final query = textEditingValue.text.trim().toLowerCase();
            if (query.length < 2) return const [];
            final queryDigits = query.replaceAll(RegExp(r'\D'), '');
            return allCustomers.where((c) {
              final phoneDigits = c.phone.replaceAll(RegExp(r'\D'), '');
              final matchesPhone = queryDigits.length >= 2 &&
                  (phoneDigits.contains(queryDigits) ||
                      (queryDigits.length >= 10 &&
                          phoneDigits.endsWith(
                              queryDigits.substring(queryDigits.length - 10))));
              return matchesPhone ||
                  c.name.toLowerCase().contains(query) ||
                  c.phone.toLowerCase().contains(query) ||
                  c.city.toLowerCase().contains(query);
            });
          },
          displayStringForOption: (c) => c.phone,
          onSelected: _onCustomerSelected,
          fieldViewBuilder: (context, controller, focusNode, onFieldSubmitted) {
            controller.addListener(() {
              if (_buyerNumberController.text != controller.text) {
                _buyerNumberController.text = controller.text;
              }
            });
            return TextFormField(
              controller: controller,
              focusNode: focusNode,
              keyboardType: TextInputType.phone,
              decoration: InputDecoration(
                labelText: _mode == 'quick'
                    ? 'Контакт покупателя (необязательно)'
                    : 'Контакт / Поиск клиента *',
                hintText: '+7 999 123-45-67 или имя клиента',
                prefixIcon: const Icon(LucideIcons.search),
              ),
              validator: (value) {
                if (_mode == 'quick') return null;
                return value == null || value.trim().isEmpty
                    ? 'Укажите контакт покупателя'
                    : null;
              },
            );
          },
          optionsViewBuilder: (context, onSelected, options) {
            return Align(
              alignment: Alignment.topLeft,
              child: Material(
                elevation: 4,
                borderRadius: BorderRadius.circular(8),
                child: SizedBox(
                  width: constraints.maxWidth,
                  child: ListView.separated(
                    padding: EdgeInsets.zero,
                    shrinkWrap: true,
                    itemCount: options.length,
                    separatorBuilder: (_, _) => const Divider(height: 1),
                    itemBuilder: (context, index) {
                      final option = options.elementAt(index);
                      return ListTile(
                        dense: true,
                        leading: Icon(
                          option.categoryIcon,
                          color: option.categoryColor,
                          size: 18,
                        ),
                        title: Text(option.name, style: const TextStyle(fontWeight: FontWeight.w600)),
                        subtitle: Text(
                          '${option.phone}${option.city.isNotEmpty ? ' · ${option.city}' : ''}',
                          style: const TextStyle(fontSize: 12),
                        ),
                        trailing: option.discountPercent > 0
                            ? Text(
                                '-${option.discountPercent.toStringAsFixed(0)}%',
                                style: const TextStyle(color: Colors.green, fontWeight: FontWeight.bold),
                              )
                            : null,
                        onTap: () => onSelected(option),
                      );
                    },
                  ),
                ),
              ),
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final ordersAsync = ref.watch(ordersProvider);
    final customers = ref.watch(customersProvider).valueOrNull ?? const <Customer>[];
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
                _buildBuyerContactField(customers),
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
                const SizedBox(height: 12),
                TextFormField(
                  controller: _discountController,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    labelText: 'Скидка (%)',
                    suffixText: '%',
                    prefixIcon: Icon(LucideIcons.percent),
                  ),
                  onChanged: (val) {
                    final pct = double.tryParse(val) ?? 0.0;
                    for (final part in widget.parts) {
                      final discounted = part.price * (1 - pct / 100);
                      _priceControllers[part.id]?.text =
                          discounted > 0 ? discounted.toStringAsFixed(0) : '';
                    }
                  },
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

