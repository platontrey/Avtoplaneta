import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/part.dart';
import '../providers/inventory_provider.dart';

enum PriceMode {
  set('PRICE_OPERATION_MODE_SET', 'Фиксированная'),
  percent('PRICE_OPERATION_MODE_PERCENT', 'Процент %'),
  offset('PRICE_OPERATION_MODE_OFFSET', 'Смещение ₽');

  final String protoName;
  final String label;
  const PriceMode(this.protoName, this.label);
}

Map<String, dynamic> buildBulkUpdatePayload({
  required Iterable<int> ids,
  PriceMode? priceMode,
  double? priceValue,
  String? category,
  String? location,
  String? salesman,
  bool? status,
}) {
  final operations = <String, dynamic>{};
  if (priceMode != null && priceValue != null) {
    operations['price'] = {
      'mode': priceMode.protoName,
      'value': priceValue,
    };
  }
  if (category != null && category.isNotEmpty) {
    operations['category'] = category;
  }
  if (location != null && location.isNotEmpty) {
    operations['location'] = location;
  }
  if (salesman != null && salesman.isNotEmpty) {
    operations['salesman'] = salesman;
  }
  if (status != null) {
    operations['status'] = status;
  }

  return {
    'scope': {
      'type': 'BULK_SCOPE_TYPE_IDS',
      'ids': ids.toList(),
    },
    'operations': operations,
  };
}

class CriteriaSelectionResult {
  final double? minPrice;
  final double? maxPrice;
  final String? category;
  final bool? status;

  const CriteriaSelectionResult({
    this.minPrice,
    this.maxPrice,
    this.category,
    this.status,
  });

  bool matches(Part part) {
    if (minPrice != null && part.price < minPrice!) return false;
    if (maxPrice != null && part.price > maxPrice!) return false;
    if (category != null && category!.isNotEmpty && part.category != category) {
      return false;
    }
    if (status != null && part.status != status) return false;
    return true;
  }
}

Future<CriteriaSelectionResult?> showSelectByCriteriaSheet(
  BuildContext context, {
  required List<String> categories,
}) async {
  return showModalBottomSheet<CriteriaSelectionResult>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    builder: (_) => _SelectByCriteriaSheet(categories: categories),
  );
}

Future<bool> showBulkEditSheet(
  BuildContext context,
  WidgetRef ref, {
  required List<Part> parts,
  required List<String> categories,
}) async {
  final changed = await showModalBottomSheet<bool>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    builder: (_) => _BulkEditSheet(parts: parts, categories: categories),
  );
  if (changed == true) {
    ref.invalidate(inventoryProvider);
    for (final part in parts) {
      ref.invalidate(partProvider(part.id));
    }
  }
  return changed ?? false;
}

Future<bool> confirmBulkDelete(
  BuildContext context,
  WidgetRef ref,
  List<Part> parts,
) async {
  final confirmed = await showDialog<bool>(
    context: context,
    builder: (dialogContext) => AlertDialog(
      title: const Text('Удалить запчасти?'),
      content: Text(
        'Будут безвозвратно удалены ${parts.length} выбранных запчастей.',
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(dialogContext, false),
          child: const Text('Отмена'),
        ),
        FilledButton(
          style: FilledButton.styleFrom(backgroundColor: Colors.red),
          onPressed: () => Navigator.pop(dialogContext, true),
          child: const Text('Удалить'),
        ),
      ],
    ),
  );
  if (confirmed != true || !context.mounted) return false;

  try {
    await apiClient.dio.post(
      '/api/v1/admin/parts/bulk-delete',
      data: {'ids': parts.map((part) => part.id).toList()},
    );
    ref.invalidate(inventoryProvider);
    return true;
  } catch (error) {
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Не удалось удалить запчасти: $error')),
      );
    }
    return false;
  }
}

class _SelectByCriteriaSheet extends StatefulWidget {
  final List<String> categories;

  const _SelectByCriteriaSheet({required this.categories});

  @override
  State<_SelectByCriteriaSheet> createState() => _SelectByCriteriaSheetState();
}

class _SelectByCriteriaSheetState extends State<_SelectByCriteriaSheet> {
  final _minPriceCtrl = TextEditingController();
  final _maxPriceCtrl = TextEditingController();
  String? _category;
  String _statusOption = 'all';

  @override
  void dispose() {
    _minPriceCtrl.dispose();
    _maxPriceCtrl.dispose();
    super.dispose();
  }

  void _apply() {
    final minPrice = double.tryParse(_minPriceCtrl.text.trim().replaceAll(',', '.'));
    final maxPrice = double.tryParse(_maxPriceCtrl.text.trim().replaceAll(',', '.'));
    bool? status;
    if (_statusOption == 'active') status = true;
    if (_statusOption == 'inactive') status = false;

    Navigator.pop(
      context,
      CriteriaSelectionResult(
        minPrice: minPrice,
        maxPrice: maxPrice,
        category: _category,
        status: status,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final bottomInset = MediaQuery.viewInsetsOf(context).bottom;
    return Padding(
      padding: EdgeInsets.fromLTRB(20, 16, 20, bottomInset + 20),
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(LucideIcons.list_filter, size: 22),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Выбрать по критерию',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                ),
                IconButton(
                  onPressed: () => Navigator.pop(context),
                  icon: const Icon(LucideIcons.x),
                ),
              ],
            ),
            const SizedBox(height: 14),
            Text(
              'Диапазон цен (₽)',
              style: Theme.of(context).textTheme.titleSmall,
            ),
            const SizedBox(height: 6),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _minPriceCtrl,
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                    inputFormatters: [
                      FilteringTextInputFormatter.allow(RegExp(r'[0-9.,]')),
                    ],
                    decoration: const InputDecoration(
                      labelText: 'От (₽)',
                      isDense: true,
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextField(
                    controller: _maxPriceCtrl,
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                    inputFormatters: [
                      FilteringTextInputFormatter.allow(RegExp(r'[0-9.,]')),
                    ],
                    decoration: const InputDecoration(
                      labelText: 'До (₽)',
                      isDense: true,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Text(
              'Категория',
              style: Theme.of(context).textTheme.titleSmall,
            ),
            const SizedBox(height: 6),
            DropdownButtonFormField<String>(
              initialValue: _category,
              isExpanded: true,
              decoration: const InputDecoration(
                labelText: 'Все категории',
                isDense: true,
              ),
              items: [
                const DropdownMenuItem<String>(
                  value: null,
                  child: Text('Все категории'),
                ),
                ...widget.categories.map(
                  (val) => DropdownMenuItem(value: val, child: Text(val)),
                ),
              ],
              onChanged: (val) => setState(() => _category = val),
            ),
            const SizedBox(height: 16),
            Text(
              'Статус активности',
              style: Theme.of(context).textTheme.titleSmall,
            ),
            const SizedBox(height: 6),
            SegmentedButton<String>(
              segments: const [
                ButtonSegment(value: 'all', label: Text('Все')),
                ButtonSegment(value: 'active', label: Text('Активные')),
                ButtonSegment(value: 'inactive', label: Text('Неактивные')),
              ],
              selected: {_statusOption},
              onSelectionChanged: (val) => setState(() => _statusOption = val.first),
            ),
            const SizedBox(height: 24),
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: _apply,
                icon: const Icon(LucideIcons.check),
                label: const Text('Выделить подходящие'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _BulkEditSheet extends StatefulWidget {
  final List<Part> parts;
  final List<String> categories;

  const _BulkEditSheet({required this.parts, required this.categories});

  @override
  State<_BulkEditSheet> createState() => _BulkEditSheetState();
}

class _BulkEditSheetState extends State<_BulkEditSheet> {
  final _formKey = GlobalKey<FormState>();
  final _price = TextEditingController();
  final _location = TextEditingController();
  final _salesman = TextEditingController();
  final Set<String> _enabled = {};
  PriceMode _priceMode = PriceMode.set;
  String? _category;
  bool _status = true;
  bool _submitting = false;

  @override
  void initState() {
    super.initState();
    _price.addListener(() => setState(() {}));
  }

  @override
  void dispose() {
    _price.dispose();
    _location.dispose();
    _salesman.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_enabled.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Выберите хотя бы одно поле для изменения')),
      );
      return;
    }
    if (!_formKey.currentState!.validate()) return;

    double? priceVal;
    if (_enabled.contains('price')) {
      priceVal = double.tryParse(_price.text.trim().replaceAll(',', '.'));
      if (priceVal == null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Укажите корректную цену')),
        );
        return;
      }
    }

    setState(() => _submitting = true);
    try {
      final payload = buildBulkUpdatePayload(
        ids: widget.parts.map((part) => part.id),
        priceMode: _enabled.contains('price') ? _priceMode : null,
        priceValue: priceVal,
        category: _enabled.contains('category') ? _category : null,
        location: _enabled.contains('location') ? _location.text.trim() : null,
        salesman: _enabled.contains('salesman') ? _salesman.text.trim() : null,
        status: _enabled.contains('status') ? _status : null,
      );

      await apiClient.dio.post(
        '/api/v1/admin/parts/bulk-update',
        data: payload,
      );
      if (mounted) Navigator.pop(context, true);
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось изменить запчасти: $error')),
        );
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  void _toggle(String field, bool value) {
    setState(() => value ? _enabled.add(field) : _enabled.remove(field));
  }

  List<Map<String, dynamic>> _calculatePricePreview() {
    if (!_enabled.contains('price') || _price.text.isEmpty) return const [];
    final val = double.tryParse(_price.text.trim().replaceAll(',', '.'));
    if (val == null) return const [];

    return widget.parts.take(2).map((part) {
      final oldPrice = part.price;
      double newPrice = oldPrice;
      if (_priceMode == PriceMode.set) {
        newPrice = val > 0 ? val : 0;
      } else if (_priceMode == PriceMode.percent) {
        newPrice = (oldPrice * (1 + val / 100)).clamp(0, double.infinity);
      } else if (_priceMode == PriceMode.offset) {
        newPrice = (oldPrice + val).clamp(0, double.infinity);
      }
      return {
        'name': part.name,
        'old': oldPrice,
        'new': newPrice,
        'diff': newPrice - oldPrice,
      };
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    final bottomInset = MediaQuery.viewInsetsOf(context).bottom;
    final pricePreviews = _calculatePricePreview();

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
                      'Изменить ${widget.parts.length} запчастей',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                  ),
                  IconButton(
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(LucideIcons.x),
                  ),
                ],
              ),

              // Изменение цены
              _fieldToggle('price', 'Изменение цены'),
              if (_enabled.contains('price')) ...[
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 6),
                  child: SegmentedButton<PriceMode>(
                    segments: const [
                      ButtonSegment(
                        value: PriceMode.set,
                        label: Text('Точная'),
                      ),
                      ButtonSegment(
                        value: PriceMode.percent,
                        label: Text('% Наценка'),
                      ),
                      ButtonSegment(
                        value: PriceMode.offset,
                        label: Text('Смещение ₽'),
                      ),
                    ],
                    selected: {_priceMode},
                    onSelectionChanged: (val) => setState(() => _priceMode = val.first),
                  ),
                ),
                TextFormField(
                  controller: _price,
                  keyboardType: const TextInputType.numberWithOptions(
                    decimal: true,
                    signed: true,
                  ),
                  inputFormatters: [
                    FilteringTextInputFormatter.allow(RegExp(r'[-0-9.,]')),
                  ],
                  decoration: InputDecoration(
                    labelText: _priceMode == PriceMode.set
                        ? 'Новая цена (₽) *'
                        : _priceMode == PriceMode.percent
                            ? 'Процент (+/- %) *'
                            : 'Смещение (+/- ₽) *',
                    hintText: _priceMode == PriceMode.set
                        ? 'Например: 2500'
                        : _priceMode == PriceMode.percent
                            ? 'Например: 10 или -15'
                            : 'Например: 500 или -300',
                  ),
                  validator: (value) =>
                      double.tryParse((value ?? '').replaceAll(',', '.')) == null
                          ? 'Укажите числовое значение'
                          : null,
                ),
                if (pricePreviews.isNotEmpty) ...[
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: Theme.of(context).colorScheme.surfaceContainerHighest.withValues(alpha: 0.5),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: Theme.of(context).colorScheme.outlineVariant,
                      ),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Row(
                          children: [
                            Icon(LucideIcons.calculator, size: 14),
                            SizedBox(width: 6),
                            Text(
                              'Предпросмотр расчета цен:',
                              style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold),
                            ),
                          ],
                        ),
                        const SizedBox(height: 4),
                        ...pricePreviews.map((p) => Padding(
                          padding: const EdgeInsets.symmetric(vertical: 2),
                          child: Text(
                            '${p['name']}: ${p['old'].toStringAsFixed(0)} ₽ ➔ ${p['new'].toStringAsFixed(0)} ₽',
                            style: const TextStyle(fontSize: 12),
                            overflow: TextOverflow.ellipsis,
                          ),
                        )),
                      ],
                    ),
                  ),
                ],
              ],

              // Категория
              _fieldToggle('category', 'Категория'),
              if (_enabled.contains('category'))
                DropdownButtonFormField<String>(
                  initialValue: _category,
                  isExpanded: true,
                  decoration: const InputDecoration(labelText: 'Категория *'),
                  items: widget.categories
                      .map(
                        (value) =>
                            DropdownMenuItem(value: value, child: Text(value)),
                      )
                      .toList(),
                  onChanged: (value) => setState(() => _category = value),
                  validator: (_) =>
                      _enabled.contains('category') &&
                              (_category == null || _category!.isEmpty)
                          ? 'Выберите категорию'
                          : null,
                ),

              // Местоположение
              _fieldToggle('location', 'Местоположение'),
              if (_enabled.contains('location'))
                TextFormField(
                  controller: _location,
                  decoration: const InputDecoration(labelText: 'Место / Склад *'),
                  validator: (value) => value == null || value.trim().isEmpty
                      ? 'Укажите место'
                      : null,
                ),

              // Продавец
              _fieldToggle('salesman', 'Продавец'),
              if (_enabled.contains('salesman'))
                TextFormField(
                  controller: _salesman,
                  decoration: const InputDecoration(labelText: 'Продавец *'),
                  validator: (value) => value == null || value.trim().isEmpty
                      ? 'Укажите продавца'
                      : null,
                ),

              // Статус
              _fieldToggle('status', 'Статус активности'),
              if (_enabled.contains('status'))
                Padding(
                  padding: const EdgeInsets.only(top: 6),
                  child: SegmentedButton<bool>(
                    segments: const [
                      ButtonSegment(value: true, label: Text('Активна')),
                      ButtonSegment(value: false, label: Text('Неактивна')),
                    ],
                    selected: {_status},
                    onSelectionChanged: (value) {
                      setState(() => _status = value.first);
                    },
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
                      : const Icon(LucideIcons.save),
                  label: const Text('Применить изменения'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _fieldToggle(String field, String label) => CheckboxListTile(
    contentPadding: EdgeInsets.zero,
    title: Text(label),
    value: _enabled.contains(field),
    onChanged: (value) => _toggle(field, value ?? false),
  );
}
