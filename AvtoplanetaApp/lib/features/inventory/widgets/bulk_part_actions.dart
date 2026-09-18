import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/part.dart';
import '../providers/inventory_provider.dart';

Map<String, dynamic> buildBulkUpdatePayload(
  Iterable<int> ids,
  Map<String, String> fields,
) => {
  'parts': ids.map((id) => {'id': id, 'fields': fields}).toList(),
};

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
  String? _category;
  bool _status = true;
  bool _submitting = false;

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
        const SnackBar(content: Text('Выберите хотя бы одно поле')),
      );
      return;
    }
    if (!_formKey.currentState!.validate()) return;

    final fields = <String, String>{
      if (_enabled.contains('category')) 'category': _category!,
      if (_enabled.contains('price'))
        'price': _price.text.trim().replaceAll(',', '.'),
      if (_enabled.contains('location')) 'location': _location.text.trim(),
      if (_enabled.contains('salesman')) 'salesman': _salesman.text.trim(),
      if (_enabled.contains('status')) 'status': _status.toString(),
    };
    setState(() => _submitting = true);
    try {
      await apiClient.dio.post(
        '/api/v1/admin/parts/bulk-update',
        data: buildBulkUpdatePayload(
          widget.parts.map((part) => part.id),
          fields,
        ),
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

  @override
  Widget build(BuildContext context) {
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
              _fieldToggle('price', 'Цена'),
              if (_enabled.contains('price'))
                TextFormField(
                  controller: _price,
                  keyboardType: const TextInputType.numberWithOptions(
                    decimal: true,
                  ),
                  inputFormatters: [
                    FilteringTextInputFormatter.allow(RegExp(r'[0-9.,]')),
                  ],
                  decoration: const InputDecoration(labelText: 'Цена *'),
                  validator: (value) =>
                      double.tryParse((value ?? '').replaceAll(',', '.')) ==
                          null
                      ? 'Укажите цену'
                      : null,
                ),
              _fieldToggle('location', 'Местоположение'),
              if (_enabled.contains('location'))
                TextFormField(
                  controller: _location,
                  decoration: const InputDecoration(labelText: 'Место *'),
                  validator: (value) => value == null || value.trim().isEmpty
                      ? 'Укажите место'
                      : null,
                ),
              _fieldToggle('salesman', 'Продавец'),
              if (_enabled.contains('salesman'))
                TextFormField(
                  controller: _salesman,
                  decoration: const InputDecoration(labelText: 'Продавец *'),
                  validator: (value) => value == null || value.trim().isEmpty
                      ? 'Укажите продавца'
                      : null,
                ),
              _fieldToggle('status', 'Статус'),
              if (_enabled.contains('status'))
                SegmentedButton<bool>(
                  segments: const [
                    ButtonSegment(value: true, label: Text('Активна')),
                    ButtonSegment(value: false, label: Text('Неактивна')),
                  ],
                  selected: {_status},
                  onSelectionChanged: (value) {
                    setState(() => _status = value.first);
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
