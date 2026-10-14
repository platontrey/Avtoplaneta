import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/theme.dart';
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
  String scopeType = 'BULK_SCOPE_TYPE_IDS',
  Map<String, dynamic>? filterQuery,
  bool confirmEntireCatalog = false,
  PriceMode? priceMode,
  double? priceValue,
  String? category,
  String? location,
  String? salesman,
  bool? status,
  String? bodyBrand,
  String? engineBrand,
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
  if (bodyBrand != null) {
    operations['body_brand'] = bodyBrand;
  }
  if (engineBrand != null) {
    operations['engine_brand'] = engineBrand;
  }

  final scope = <String, dynamic>{
    'type': scopeType,
  };
  if (scopeType == 'BULK_SCOPE_TYPE_FILTER') {
    if (filterQuery != null) {
      final cleanFilter = Map<String, dynamic>.from(filterQuery)
        ..remove('page')
        ..remove('limit');
      scope['filter'] = cleanFilter;
    }
  } else {
    scope['ids'] = ids.toList();
  }

  final payload = <String, dynamic>{
    'scope': scope,
    'operations': operations,
  };
  if (confirmEntireCatalog) {
    payload['confirm_entire_catalog'] = true;
  }

  return payload;
}

Future<bool> showBulkEditSheet(
  BuildContext context,
  WidgetRef ref, {
  required List<Part> parts,
  required List<String> categories,
  InventoryFilter? activeFilter,
  bool initialSelectAllByQuery = false,
  int? totalMatchingCount,
}) async {
  final changed = await showModalBottomSheet<bool>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    builder: (_) => _BulkEditSheet(
      parts: parts,
      categories: categories,
      activeFilter: activeFilter,
      initialSelectAllByQuery: initialSelectAllByQuery,
      totalMatchingCount: totalMatchingCount,
    ),
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
  WidgetRef ref, {
  required List<Part> parts,
  bool selectAllByQuery = false,
  InventoryFilter? activeFilter,
  int? totalMatchingCount,
}) async {
  // Проверяем пустой ли фильтр
  final cleanFilter = activeFilter != null
      ? (Map<String, dynamic>.from(activeFilter.toQueryParameters())
          ..remove('page')
          ..remove('limit')
          ..remove('hasPhoto'))
      : <String, dynamic>{};
  final isFilterEmpty = selectAllByQuery && cleanFilter.isEmpty;

  final confirmed = await showDialog<bool>(
    context: context,
    builder: (dialogContext) {
      bool confirmEntire = false;
      return StatefulBuilder(
        builder: (ctx, setState) {
          return AlertDialog(
            title: Row(
              children: [
                Icon(
                  LucideIcons.triangle_alert,
                  color: isFilterEmpty ? AppTheme.dangerColor : Colors.amber.shade700,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    selectAllByQuery
                        ? (isFilterEmpty
                            ? 'УДАЛЕНИЕ ВСЕГО СКЛАДА!'
                            : 'Удалить все по запросу?')
                        : 'Удалить запчасти?',
                  ),
                ),
              ],
            ),
            content: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (selectAllByQuery) ...[
                  if (isFilterEmpty) ...[
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: AppTheme.dangerColor.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: AppTheme.dangerColor),
                      ),
                      child: const Text(
                        'Внимание! Поисковый запрос и фильтры пусты. Это действие удалит ВСЕ запчасти из базы данных склада без возможности восстановления!',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          color: AppTheme.dangerColor,
                          fontSize: 13,
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),
                    CheckboxListTile(
                      contentPadding: EdgeInsets.zero,
                      title: const Text(
                        'Я подтверждаю безвозвратное удаление ВСЕГО склада',
                        style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
                      ),
                      value: confirmEntire,
                      onChanged: (val) => setState(() => confirmEntire = val ?? false),
                    ),
                  ] else ...[
                    Text(
                      'Будут безвозвратно удалены все запчасти, соответствующие текущему поисковому запросу / фильтрам'
                      '${totalMatchingCount != null ? ' (найдено: $totalMatchingCount)' : ''}.',
                    ),
                  ],
                ] else ...[
                  Text(
                    'Будут безвозвратно удалены ${parts.length} выбранных запчастей.',
                  ),
                ],
              ],
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: const Text('Отмена'),
              ),
              FilledButton(
                style: FilledButton.styleFrom(backgroundColor: AppTheme.dangerColor),
                onPressed: (isFilterEmpty && !confirmEntire)
                    ? null
                    : () => Navigator.pop(dialogContext, true),
                child: Text(
                  selectAllByQuery ? 'Удалить всё' : 'Удалить',
                ),
              ),
            ],
          );
        },
      );
    },
  );

  if (confirmed != true || !context.mounted) return false;

  try {
    final Map<String, dynamic> body;
    if (selectAllByQuery) {
      final query = activeFilter != null
          ? (Map<String, dynamic>.from(activeFilter.toQueryParameters())
              ..remove('page')
              ..remove('limit'))
          : <String, dynamic>{};
      body = {
        'scope': {
          'type': 'BULK_SCOPE_TYPE_FILTER',
          'filter': query,
        },
        if (isFilterEmpty) 'confirm_entire_catalog': true,
      };
    } else {
      final idsList = parts.map((part) => part.id).toList();
      body = {
        'scope': {
          'type': 'BULK_SCOPE_TYPE_IDS',
          'ids': idsList,
        },
        'ids': idsList,
      };
    }

    await apiClient.dio.post(
      '/api/v1/admin/parts/bulk-delete',
      data: body,
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

class _BulkEditSheet extends StatefulWidget {
  final List<Part> parts;
  final List<String> categories;
  final InventoryFilter? activeFilter;
  final bool initialSelectAllByQuery;
  final int? totalMatchingCount;

  const _BulkEditSheet({
    required this.parts,
    required this.categories,
    this.activeFilter,
    this.initialSelectAllByQuery = false,
    this.totalMatchingCount,
  });

  @override
  State<_BulkEditSheet> createState() => _BulkEditSheetState();
}

class _BulkEditSheetState extends State<_BulkEditSheet> {
  final _formKey = GlobalKey<FormState>();
  final _price = TextEditingController();
  final _location = TextEditingController();
  final _salesman = TextEditingController();
  final _bodyBrand = TextEditingController();
  final _engineBrand = TextEditingController();
  final Set<String> _enabled = {};
  PriceMode _priceMode = PriceMode.set;
  late String _scopeMode;
  bool _confirmEntireCatalog = false;
  String? _category;
  bool _status = true;
  bool _submitting = false;

  bool get _isFilterEmpty {
    if (widget.activeFilter == null) return true;
    final clean = Map<String, dynamic>.from(widget.activeFilter!.toQueryParameters())
      ..remove('page')
      ..remove('limit')
      ..remove('hasPhoto');
    return clean.isEmpty;
  }

  @override
  void initState() {
    super.initState();
    _scopeMode = widget.initialSelectAllByQuery ? 'filter' : 'ids';
    _price.addListener(() => setState(() {}));
  }

  @override
  void dispose() {
    _price.dispose();
    _location.dispose();
    _salesman.dispose();
    _bodyBrand.dispose();
    _engineBrand.dispose();
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

    final isFilterScope = _scopeMode == 'filter';
    if (isFilterScope && _isFilterEmpty && !_confirmEntireCatalog) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Необходимо подтвердить массовое изменение всего склада'),
        ),
      );
      return;
    }

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
        scopeType: isFilterScope ? 'BULK_SCOPE_TYPE_FILTER' : 'BULK_SCOPE_TYPE_IDS',
        filterQuery: isFilterScope ? widget.activeFilter?.toQueryParameters() : null,
        confirmEntireCatalog: isFilterScope && _isFilterEmpty && _confirmEntireCatalog,
        priceMode: _enabled.contains('price') ? _priceMode : null,
        priceValue: priceVal,
        category: _enabled.contains('category') ? _category : null,
        location: _enabled.contains('location') ? _location.text.trim() : null,
        salesman: _enabled.contains('salesman') ? _salesman.text.trim() : null,
        status: _enabled.contains('status') ? _status : null,
        bodyBrand: _enabled.contains('body_brand') ? _bodyBrand.text.trim() : null,
        engineBrand: _enabled.contains('engine_brand') ? _engineBrand.text.trim() : null,
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
    final filterEmpty = _isFilterEmpty;

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
                      _scopeMode == 'filter'
                          ? (widget.totalMatchingCount != null
                              ? 'Изменить все по запросу (${widget.totalMatchingCount})'
                              : 'Изменить все по запросу')
                          : 'Изменить ${widget.parts.length} запчастей',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                  ),
                  IconButton(
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(LucideIcons.x),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              SegmentedButton<String>(
                segments: [
                  ButtonSegment(
                    value: 'ids',
                    label: Text('Выбранные (${widget.parts.length})'),
                  ),
                  ButtonSegment(
                    value: 'filter',
                    label: Text(
                      widget.totalMatchingCount != null
                          ? 'Все по запросу (${widget.totalMatchingCount})'
                          : 'Все по запросу',
                    ),
                  ),
                ],
                selected: {_scopeMode},
                onSelectionChanged: (val) => setState(() => _scopeMode = val.first),
              ),
              const SizedBox(height: 8),

              if (_scopeMode == 'filter' && filterEmpty) ...[
                Container(
                  margin: const EdgeInsets.only(bottom: 8),
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: Colors.amber.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.amber.shade700),
                  ),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Icon(LucideIcons.triangle_alert, color: Colors.amber.shade800, size: 20),
                      const SizedBox(width: 8),
                      const Expanded(
                        child: Text(
                          'Внимание! Поисковый запрос и фильтры пусты. Изменения затронут ВСЕ запчасти на складе.',
                          style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold),
                        ),
                      ),
                    ],
                  ),
                ),
                CheckboxListTile(
                  contentPadding: EdgeInsets.zero,
                  title: const Text(
                    'Подтверждаю массовое изменение всего каталога',
                    style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
                  ),
                  value: _confirmEntireCatalog,
                  onChanged: (val) => setState(() => _confirmEntireCatalog = val ?? false),
                ),
                const SizedBox(height: 8),
              ],

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

              // Марка кузова
              _fieldToggle('body_brand', 'Марка кузова'),
              if (_enabled.contains('body_brand'))
                TextFormField(
                  controller: _bodyBrand,
                  decoration: const InputDecoration(
                    labelText: 'Марка кузова',
                    hintText: 'Например: AE100, W204, E39',
                  ),
                ),

              // Марка двигателя
              _fieldToggle('engine_brand', 'Марка двигателя'),
              if (_enabled.contains('engine_brand'))
                TextFormField(
                  controller: _engineBrand,
                  decoration: const InputDecoration(
                    labelText: 'Марка двигателя',
                    hintText: 'Например: 1JZ-GTE, 2GR-FE, M54',
                  ),
                ),

              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: FilledButton.icon(
                  onPressed: (_submitting || (_scopeMode == 'filter' && filterEmpty && !_confirmEntireCatalog))
                      ? null
                      : _submit,
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
