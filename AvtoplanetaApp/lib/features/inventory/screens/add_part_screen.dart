import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/api/api_client.dart';
import '../providers/inventory_provider.dart';

class AddPartScreen extends ConsumerStatefulWidget {
  final int? editId;
  const AddPartScreen({super.key, this.editId});

  @override
  ConsumerState<AddPartScreen> createState() => _AddPartScreenState();
}

class _AddPartScreenState extends ConsumerState<AddPartScreen> {
  final _formKey = GlobalKey<FormState>();
  bool _loading = false;

  final _nameCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  final _categoryCtrl = TextEditingController();
  final _priceCtrl = TextEditingController();
  final _quantityCtrl = TextEditingController(text: '1');
  final _brandCtrl = TextEditingController();
  final _modelCtrl = TextEditingController();
  final _colorCtrl = TextEditingController();
  final _conditionCtrl = TextEditingController();
  final _oemCtrl = TextEditingController();
  final _supplierCtrl = TextEditingController();
  final _vinCtrl = TextEditingController();
  final _locationCtrl = TextEditingController();
  final _salesmanCtrl = TextEditingController();

  @override
  void initState() {
    super.initState();
    if (widget.editId != null) _loadPart();
  }

  Future<void> _loadPart() async {
    try {
      final response =
          await apiClient.dio.get('/api/inventory/${widget.editId}');
      final data = response.data as Map<String, dynamic>;
      setState(() {
        _nameCtrl.text = data['name'] ?? '';
        _descCtrl.text = data['description'] ?? '';
        _categoryCtrl.text = data['category'] ?? '';
        _priceCtrl.text = (data['price'] ?? '').toString();
        _quantityCtrl.text = (data['quantity'] ?? '1').toString();
        _brandCtrl.text = data['brand'] ?? '';
        _modelCtrl.text = data['model'] ?? '';
        _colorCtrl.text = data['color'] ?? '';
        _conditionCtrl.text = data['condition'] ?? '';
        _oemCtrl.text = data['oem_code'] ?? '';
        _supplierCtrl.text = data['supplier_code'] ?? '';
        _vinCtrl.text = data['vin'] ?? '';
        _locationCtrl.text = data['location'] ?? '';
        _salesmanCtrl.text = data['salesman'] ?? '';
      });
    } catch (_) {}
  }

  @override
  void dispose() {
    for (final c in [
      _nameCtrl, _descCtrl, _categoryCtrl, _priceCtrl, _quantityCtrl,
      _brandCtrl, _modelCtrl, _colorCtrl, _conditionCtrl, _oemCtrl,
      _supplierCtrl, _vinCtrl, _locationCtrl, _salesmanCtrl,
    ]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _loading = true);
    try {
      final data = {
        'name': _nameCtrl.text.trim(),
        'description': _descCtrl.text.trim(),
        'category': _categoryCtrl.text.trim(),
        'price': double.tryParse(_priceCtrl.text) ?? 0,
        'quantity': int.tryParse(_quantityCtrl.text) ?? 1,
        'brand': _brandCtrl.text.trim(),
        'model': _modelCtrl.text.trim(),
        'color': _colorCtrl.text.trim(),
        'condition': _conditionCtrl.text.trim(),
        'oem_code': _oemCtrl.text.trim(),
        'supplier_code': _supplierCtrl.text.trim(),
        'vin': _vinCtrl.text.trim(),
        'location': _locationCtrl.text.trim(),
        'salesman': _salesmanCtrl.text.trim(),
      };

      if (widget.editId != null) {
        await apiClient.dio.put('/api/updatepart/${widget.editId}', data: data);
      } else {
        await apiClient.dio.post('/api/addpart', data: data);
      }

      ref.invalidate(inventoryProvider);
      if (mounted) context.go('/inventory');
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.editId != null ? 'Редактировать' : 'Добавить запчасть'),
        actions: [
          if (_loading)
            const Padding(
              padding: EdgeInsets.all(16),
              child: SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2)),
            )
          else
            TextButton(
              onPressed: _submit,
              child: const Text('Сохранить',
                  style: TextStyle(color: Color(0xFF4F8EF7))),
            ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            _section('Основное'),
            _field(_nameCtrl, 'Название *', required: true),
            _field(_categoryCtrl, 'Категория *', required: true),
            _field(_priceCtrl, 'Цена (₽) *',
                required: true,
                keyboard: TextInputType.number),
            _field(_quantityCtrl, 'Количество *',
                required: true,
                keyboard: TextInputType.number),
            _field(_descCtrl, 'Описание', maxLines: 3),

            _section('Характеристики'),
            _field(_brandCtrl, 'Бренд'),
            _field(_modelCtrl, 'Модель'),
            _field(_colorCtrl, 'Цвет'),
            _field(_conditionCtrl, 'Состояние'),

            _section('Коды'),
            _field(_oemCtrl, 'OEM код'),
            _field(_supplierCtrl, 'Код поставщика'),
            _field(_vinCtrl, 'VIN'),

            _section('Расположение'),
            _field(_locationCtrl, 'Место хранения'),
            _field(_salesmanCtrl, 'Продавец'),

            const SizedBox(height: 80),
          ],
        ),
      ),
    );
  }

  Widget _section(String title) => Padding(
        padding: const EdgeInsets.only(top: 16, bottom: 8),
        child: Text(
          title,
          style: const TextStyle(
              color: Colors.white54,
              fontSize: 12,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.5),
        ),
      );

  Widget _field(
    TextEditingController ctrl,
    String label, {
    bool required = false,
    int maxLines = 1,
    TextInputType keyboard = TextInputType.text,
  }) =>
      Padding(
        padding: const EdgeInsets.only(bottom: 12),
        child: TextFormField(
          controller: ctrl,
          maxLines: maxLines,
          keyboardType: keyboard,
          style: const TextStyle(color: Colors.white),
          decoration: InputDecoration(labelText: label),
          validator: required
              ? (v) =>
                  (v == null || v.trim().isEmpty) ? 'Обязательное поле' : null
              : null,
        ),
      );
}
