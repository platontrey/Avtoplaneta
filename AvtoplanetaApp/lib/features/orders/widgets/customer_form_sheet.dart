import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/customer.dart';
import '../providers/customers_provider.dart';

class CustomerFormSheet extends ConsumerStatefulWidget {
  final Customer? customer;

  const CustomerFormSheet({super.key, this.customer});

  static Future<bool?> show(BuildContext context, {Customer? customer}) {
    return showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (_) => CustomerFormSheet(customer: customer),
    );
  }

  @override
  ConsumerState<CustomerFormSheet> createState() => _CustomerFormSheetState();
}

class _CustomerFormSheetState extends ConsumerState<CustomerFormSheet> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _nameController;
  late final TextEditingController _phoneController;
  late final TextEditingController _cityController;
  late final TextEditingController _tkController;
  late final TextEditingController _tkDetailsController;
  late final TextEditingController _notesController;
  late final TextEditingController _discountController;
  late String _category;
  bool _submitting = false;

  final List<String> _commonTKs = [
    'СДЭК',
    'Энергия',
    'ПЭК',
    'Деловые Линии',
    'КИТ (GTD)',
    'Почта России',
    'Байкал Сервис',
  ];

  @override
  void initState() {
    super.initState();
    final c = widget.customer;
    _nameController = TextEditingController(text: c?.name ?? '');
    _phoneController = TextEditingController(text: c?.phone ?? '');
    _cityController = TextEditingController(text: c?.city ?? '');
    _tkController = TextEditingController(text: c?.preferredTk ?? '');
    _tkDetailsController = TextEditingController(text: c?.tkDetails ?? '');
    _notesController = TextEditingController(text: c?.notes ?? '');
    _discountController = TextEditingController(
      text: c != null && c.discountPercent > 0
          ? c.discountPercent.toStringAsFixed(0)
          : '0',
    );
    _category = c?.category ?? 'regular';
  }

  @override
  void dispose() {
    _nameController.dispose();
    _phoneController.dispose();
    _cityController.dispose();
    _tkController.dispose();
    _tkDetailsController.dispose();
    _notesController.dispose();
    _discountController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate() || _submitting) return;

    setState(() => _submitting = true);
    try {
      final discount = double.tryParse(_discountController.text) ?? 0.0;
      if (widget.customer != null) {
        await CustomerService.updateCustomer(
          id: widget.customer!.id,
          name: _nameController.text.trim(),
          phone: _phoneController.text.trim(),
          city: _cityController.text.trim(),
          preferredTk: _tkController.text.trim(),
          tkDetails: _tkDetailsController.text.trim(),
          notes: _notesController.text.trim(),
          category: _category,
          discountPercent: discount,
        );
        ref.invalidate(customerDetailsProvider(widget.customer!.id));
      } else {
        await CustomerService.createCustomer(
          name: _nameController.text.trim(),
          phone: _phoneController.text.trim(),
          city: _cityController.text.trim(),
          preferredTk: _tkController.text.trim(),
          tkDetails: _tkDetailsController.text.trim(),
          notes: _notesController.text.trim(),
          category: _category,
          discountPercent: discount,
        );
      }

      ref.invalidate(customersProvider);
      if (mounted) {
        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка сохранения: $e')),
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
    final bottomInset = MediaQuery.viewInsetsOf(context).bottom;
    final isEdit = widget.customer != null;

    return Padding(
      padding: EdgeInsets.fromLTRB(20, 16, 20, bottomInset + 20),
      child: Form(
        key: _formKey,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(
                    isEdit ? LucideIcons.user_pen : LucideIcons.user_plus,
                    color: Theme.of(context).colorScheme.primary,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      isEdit ? 'Редактировать клиента' : 'Новый клиент',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(LucideIcons.x),
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              ),
              const Divider(height: 24),
              // Имя
              TextFormField(
                controller: _nameController,
                decoration: const InputDecoration(
                  labelText: 'Имя / Название СТО *',
                  hintText: 'Иван Иванов или Автосервис Форсаж',
                  prefixIcon: Icon(LucideIcons.user),
                ),
                validator: (val) {
                  if (val == null || val.trim().isEmpty) {
                    return 'Введите имя клиента';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 12),
              // Телефон
              TextFormField(
                controller: _phoneController,
                keyboardType: TextInputType.phone,
                decoration: const InputDecoration(
                  labelText: 'Телефон *',
                  hintText: '+7 999 123-45-67',
                  prefixIcon: Icon(LucideIcons.phone),
                ),
                validator: (val) {
                  if (val == null || val.trim().isEmpty) {
                    return 'Введите номер телефона';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 12),
              // Категория
              DropdownButtonFormField<String>(
                initialValue: _category,
                decoration: const InputDecoration(
                  labelText: 'Категория клиента',
                  prefixIcon: Icon(LucideIcons.tag),
                ),
                items: const [
                  DropdownMenuItem(
                    value: 'regular',
                    child: Text('Обычный'),
                  ),
                  DropdownMenuItem(
                    value: 'vip',
                    child: Row(
                      children: [
                        Text('⭐ СТО / VIP'),
                      ],
                    ),
                  ),
                  DropdownMenuItem(
                    value: 'wholesale',
                    child: Row(
                      children: [
                        Text('🏢 Оптовик'),
                      ],
                    ),
                  ),
                  DropdownMenuItem(
                    value: 'blacklist',
                    child: Row(
                      children: [
                        Text('⚠️ Черный список'),
                      ],
                    ),
                  ),
                ],
                onChanged: (val) {
                  if (val != null) setState(() => _category = val);
                },
              ),
              if (_category == 'blacklist') ...[
                const SizedBox(height: 8),
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: Colors.red.withAlpha(25),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.red.withAlpha(80)),
                  ),
                  child: const Row(
                    children: [
                      Icon(LucideIcons.triangle_alert, color: Colors.red, size: 20),
                      SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          'Клиент будет помечен предупреждением при оформлении заказов.',
                          style: TextStyle(color: Colors.red, fontSize: 13),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
              const SizedBox(height: 12),
              // Город и Скидка в одну строку
              Row(
                children: [
                  Expanded(
                    flex: 2,
                    child: TextFormField(
                      controller: _cityController,
                      decoration: const InputDecoration(
                        labelText: 'Город доставки',
                        hintText: 'г. Новосибирск',
                        prefixIcon: Icon(LucideIcons.map_pin),
                      ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    flex: 1,
                    child: TextFormField(
                      controller: _discountController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        labelText: 'Скидка %',
                        hintText: '5',
                        suffixText: '%',
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              // Предпочитаемая ТК
              Autocomplete<String>(
                initialValue: TextEditingValue(text: _tkController.text),
                optionsBuilder: (textEditingValue) {
                  if (textEditingValue.text.isEmpty) {
                    return _commonTKs;
                  }
                  return _commonTKs.where((tk) =>
                      tk.toLowerCase().contains(textEditingValue.text.toLowerCase()));
                },
                onSelected: (selection) {
                  _tkController.text = selection;
                },
                fieldViewBuilder: (context, controller, focusNode, onFieldSubmitted) {
                  controller.addListener(() {
                    _tkController.text = controller.text;
                  });
                  return TextFormField(
                    controller: controller,
                    focusNode: focusNode,
                    decoration: const InputDecoration(
                      labelText: 'Предпочитаемая ТК',
                      hintText: 'СДЭК, Энергия, ПЭК...',
                      prefixIcon: Icon(LucideIcons.truck),
                    ),
                  );
                },
              ),
              const SizedBox(height: 12),
              // Реквизиты для отправки ТК
              TextFormField(
                controller: _tkDetailsController,
                maxLines: 2,
                decoration: const InputDecoration(
                  labelText: 'Реквизиты ТК для накладной',
                  hintText: 'Серия/номер паспорта получателя, адрес склада ТК, ИНН...',
                  prefixIcon: Icon(LucideIcons.file_text),
                ),
              ),
              const SizedBox(height: 12),
              // Заметки
              TextFormField(
                controller: _notesController,
                maxLines: 2,
                decoration: const InputDecoration(
                  labelText: 'Заметки / Причина ЧС',
                  hintText: 'Особенности работы, контактные лица...',
                  prefixIcon: Icon(LucideIcons.notebook_pen),
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
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Icon(isEdit ? LucideIcons.check : LucideIcons.user_plus),
                  label: Text(
                    isEdit ? 'Сохранить изменения' : 'Создать клиента',
                    style: const TextStyle(fontWeight: FontWeight.w600),
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
