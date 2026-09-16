import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/api/api_client.dart';
import '../providers/inventory_provider.dart';
import '../providers/part_catalog_provider.dart';

class DefectReportScreen extends ConsumerStatefulWidget {
  const DefectReportScreen({super.key});

  @override
  ConsumerState<DefectReportScreen> createState() => _DefectReportScreenState();
}

class _DefectReportScreenState extends ConsumerState<DefectReportScreen> {
  final _formKey = GlobalKey<FormState>();
  bool _loading = false;

  // Form controllers
  final _modelCtrl = TextEditingController();
  final _yearCtrl = TextEditingController();
  final _vinCtrl = TextEditingController();
  final _carReleasePeriodCtrl = TextEditingController();
  final _mileageCtrl = TextEditingController();
  final _engineBrandCtrl = TextEditingController();
  final _bodyBrandCtrl = TextEditingController();
  final _transmissionModelCtrl = TextEditingController();
  final _driveCtrl = TextEditingController();
  final _descCtrl = TextEditingController(
    text:
        "В связи с изменением цены конечную стоимость товара узнавать по WhatsApp 89138538227",
  );

  String? _selectedBrand;
  String? _selectedTransmission;
  String? _selectedInteriorColor;
  String? _selectedBodyColor;

  final List<String> _brands = [
    "BMW",
    "Audi",
    "Mercedes",
    "Toyota",
    "Volkswagen",
    "Honda",
    "Ford",
  ];

  final List<String> _availableColors = [
    "Черный",
    "Белый",
    "Серебристый",
    "Серый",
    "Темно-серый",
    "Светло-серый",
    "Синий",
    "Темно-синий",
    "Светло-синий",
    "Красный",
    "Темно-красный",
    "Бордовый",
    "Зеленый",
    "Темно-зеленый",
    "Светло-зеленый",
    "Желтый",
    "Оранжевый",
    "Фиолетовый",
    "Коричневый",
    "Бежевый",
    "Золотой",
    "Бронзовый",
    "Перламутровый",
    "Металлик",
    "Матовый",
  ];

  // Preview management
  int _displayLimit = 10;
  String _previewFilter = '';

  @override
  void dispose() {
    _modelCtrl.dispose();
    _yearCtrl.dispose();
    _vinCtrl.dispose();
    _carReleasePeriodCtrl.dispose();
    _mileageCtrl.dispose();
    _engineBrandCtrl.dispose();
    _bodyBrandCtrl.dispose();
    _transmissionModelCtrl.dispose();
    _driveCtrl.dispose();
    _descCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (_selectedBrand == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Выберите бренд автомобиля')),
      );
      return;
    }

    setState(() => _loading = true);

    try {
      final catalog = await ref.read(partCatalogProvider.future);
      final values = <String, dynamic>{
        'body_brand': _bodyBrandCtrl.text.trim(),
        'engine_brand': _engineBrandCtrl.text.trim(),
        'year': int.tryParse(_yearCtrl.text) ?? 0,
        'vin': _vinCtrl.text.trim(),
        'car_release_period': _carReleasePeriodCtrl.text.trim(),
        'transmission': _selectedTransmission,
        'transmission_model': _transmissionModelCtrl.text.trim(),
        'drive': _driveCtrl.text.trim(),
        'interior_color': _selectedInteriorColor,
        'body_color': _selectedBodyColor,
      };

      final payload = {
        'brand': _selectedBrand,
        'model': _modelCtrl.text.trim(),
        'year': int.tryParse(_yearCtrl.text) ?? 0,
        'vin': _vinCtrl.text.trim(),
        'car_release_period': _carReleasePeriodCtrl.text.trim(),
        'mileage': int.tryParse(_mileageCtrl.text) ?? 0,
        'engine_brand': _engineBrandCtrl.text.trim(),
        'body_brand': _bodyBrandCtrl.text.trim(),
        'interior_color': _selectedInteriorColor,
        'body_color': _selectedBodyColor,
        'transmission': _selectedTransmission,
        'transmission_model': _transmissionModelCtrl.text.trim(),
        'drive': _driveCtrl.text.trim(),
        'description': _descCtrl.text.trim(),
        'catalog_version': catalog.version,
        'selectedParts': catalog.expandDefectReportParts(
          values,
          supplierCode: DateTime.now().millisecondsSinceEpoch.toString(),
        ),
      };

      await apiClient.dio.post('/api/defect-reports', data: payload);

      if (!mounted) return;

      ref.invalidate(inventoryProvider);
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Дефектная ведомость успешно создана!')),
      );
      context.go('/inventory');
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text('Ошибка создания: $e')));
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final catalogState = ref.watch(partCatalogProvider);
    final catalog = catalogState.asData?.value;
    if (catalog == null) {
      return Scaffold(
        appBar: AppBar(title: const Text('Дефектная ведомость')),
        body: Center(
          child: catalogState.hasError
              ? Padding(
                  padding: const EdgeInsets.all(24),
                  child: Text(
                    'Не удалось загрузить каталог: ${catalogState.error}',
                    textAlign: TextAlign.center,
                  ),
                )
              : const CircularProgressIndicator(),
        ),
      );
    }

    final allParts = catalog.parts.map((part) => part.toPreviewMap()).toList();
    final driveCategories = catalog.categoriesForBinding('drive');
    final transmissionModelCategories = catalog.categoriesForBinding(
      'transmission_model',
    );
    final transmissionOptions = catalog.optionsForAttribute('transmission');
    final driveOptions = catalog.optionsForAttribute('drive');

    // Filter parts for preview based on filter text
    final filteredParts = allParts.where((part) {
      if (_previewFilter.isEmpty) return true;
      final name = (part['name'] as String? ?? '').toLowerCase();
      final cat = (part['category'] as String? ?? '').toLowerCase();
      final query = _previewFilter.toLowerCase();
      return name.contains(query) || cat.contains(query);
    }).toList();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Дефектная ведомость'),
        actions: [
          if (_loading)
            const Padding(
              padding: EdgeInsets.all(16),
              child: SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
            )
          else
            TextButton(
              onPressed: _submit,
              child: const Text(
                'Создать',
                style: TextStyle(
                  color: Color(0xFF4F8EF7),
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            _sectionTitle('Информация об автомобиле'),
            const SizedBox(height: 8),
            _buildDropdown(
              value: _selectedBrand,
              label: 'Бренд *',
              items: _brands,
              onChanged: (val) => setState(() => _selectedBrand = val),
            ),
            _buildTextField(_modelCtrl, 'Модель *', required: true),
            _buildTextField(
              _bodyBrandCtrl,
              'Марка кузова',
              hint: 'Например: E90',
            ),
            _buildTextField(
              _engineBrandCtrl,
              'Марка двигателя',
              hint: 'Например: Toyota 1NZ-FE',
            ),
            _buildTextField(
              _yearCtrl,
              'Год выпуска *',
              required: true,
              keyboard: TextInputType.number,
            ),
            _buildTextField(_vinCtrl, 'VIN / Номер кузова'),
            _buildTextField(
              _carReleasePeriodCtrl,
              'Период выпуска автомобиля',
              hint: 'Например: 2001-2007',
            ),
            _buildTextField(
              _mileageCtrl,
              'Пробег (км) *',
              required: true,
              keyboard: TextInputType.number,
            ),
            _buildDropdown(
              value: _selectedTransmission,
              label: 'Тип трансмиссии',
              items: transmissionOptions,
              onChanged: (val) => setState(() => _selectedTransmission = val),
            ),
            _buildTextField(
              _transmissionModelCtrl,
              'Модель трансмиссии',
              hint:
                  'Применяется к подвеске ДВС/КПП, трансмиссии и передней подвеске',
            ),
            _buildDropdown(
              value: driveOptions.contains(_driveCtrl.text.trim())
                  ? _driveCtrl.text.trim()
                  : null,
              label: 'Привод',
              items: driveOptions,
              onChanged: (val) => setState(() => _driveCtrl.text = val ?? ''),
            ),
            const SizedBox(height: 12),
            _sectionTitle('Цвета деталей'),
            _buildDropdown(
              value: _selectedInteriorColor,
              label: 'Цвет салона',
              items: _availableColors,
              onChanged: (val) => setState(() => _selectedInteriorColor = val),
            ),
            _buildDropdown(
              value: _selectedBodyColor,
              label: 'Цвет кузовных деталей',
              items: _availableColors,
              onChanged: (val) => setState(() => _selectedBodyColor = val),
            ),

            const SizedBox(height: 12),
            _sectionTitle('Заметки'),
            _buildTextField(_descCtrl, 'Описание / Заметки', maxLines: 3),

            const SizedBox(height: 16),
            _buildPartsPreviewSection(
              filteredParts,
              totalParts: allParts.length,
              driveCategories: driveCategories,
              transmissionModelCategories: transmissionModelCategories,
            ),
            const SizedBox(height: 80),
          ],
        ),
      ),
    );
  }

  Widget _sectionTitle(String title) {
    return Padding(
      padding: const EdgeInsets.only(top: 8, bottom: 8),
      child: Text(
        title,
        style: const TextStyle(
          color: Colors.white70,
          fontSize: 14,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }

  Widget _buildTextField(
    TextEditingController ctrl,
    String label, {
    bool required = false,
    int maxLines = 1,
    TextInputType keyboard = TextInputType.text,
    String? hint,
    ValueChanged<String>? onChanged,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: TextFormField(
        controller: ctrl,
        maxLines: maxLines,
        keyboardType: keyboard,
        onChanged: onChanged,
        style: const TextStyle(color: Colors.white),
        decoration: InputDecoration(
          labelText: label,
          hintText: hint,
          hintStyle: const TextStyle(color: Colors.white30),
        ),
        validator: required
            ? (v) =>
                  (v == null || v.trim().isEmpty) ? 'Обязательное поле' : null
            : null,
      ),
    );
  }

  Widget _buildDropdown({
    required String? value,
    required String label,
    required List<String> items,
    required void Function(String?) onChanged,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: DropdownButtonFormField<String>(
        // ignore: deprecated_member_use
        value: value,
        decoration: InputDecoration(labelText: label),
        dropdownColor: const Color(0xFF16213E),
        style: const TextStyle(color: Colors.white, fontSize: 16),
        items: items
            .map((item) => DropdownMenuItem(value: item, child: Text(item)))
            .toList(),
        onChanged: onChanged,
      ),
    );
  }

  Widget _buildPartsPreviewSection(
    List<Map<String, dynamic>> filteredParts, {
    required int totalParts,
    required Set<String> driveCategories,
    required Set<String> transmissionModelCategories,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            const Text(
              'Создаваемые запчасти',
              style: TextStyle(
                color: Colors.white70,
                fontSize: 14,
                fontWeight: FontWeight.bold,
              ),
            ),
            Text(
              'Всего: $totalParts',
              style: const TextStyle(color: Colors.white38, fontSize: 12),
            ),
          ],
        ),
        const SizedBox(height: 8),
        TextField(
          style: const TextStyle(color: Colors.white, fontSize: 14),
          decoration: const InputDecoration(
            hintText: 'Поиск по превью деталей...',
            prefixIcon: Icon(Icons.search, size: 18),
            contentPadding: EdgeInsets.symmetric(vertical: 8),
          ),
          onChanged: (val) => setState(() => _previewFilter = val),
        ),
        const SizedBox(height: 8),
        Row(
          children: [
            const Text(
              'Показывать:',
              style: TextStyle(color: Colors.white54, fontSize: 12),
            ),
            const SizedBox(width: 8),
            _limitButton(10),
            _limitButton(30),
            _limitButton(50),
            _limitButton(100),
            _limitButton(filteredParts.length, label: 'Все'),
          ],
        ),
        const SizedBox(height: 8),
        Container(
          height: 250,
          decoration: BoxDecoration(
            color: const Color(0xFF16213E),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.white10),
          ),
          child: filteredParts.isEmpty
              ? const Center(
                  child: Text(
                    'Ничего не найдено',
                    style: TextStyle(color: Colors.white30),
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(8),
                  itemCount: filteredParts.length.clamp(0, _displayLimit),
                  itemBuilder: (ctx, i) {
                    final part = filteredParts[i];
                    return Container(
                      padding: const EdgeInsets.symmetric(
                        vertical: 6,
                        horizontal: 4,
                      ),
                      decoration: const BoxDecoration(
                        border: Border(
                          bottom: BorderSide(color: Colors.white10, width: 0.5),
                        ),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  part['name'] ?? '',
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 13,
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  part['category'] ?? '',
                                  style: const TextStyle(
                                    color: Colors.white38,
                                    fontSize: 11,
                                  ),
                                ),
                                if (driveCategories.contains(
                                      part['category'],
                                    ) &&
                                    _driveCtrl.text.trim().isNotEmpty)
                                  Text(
                                    'Привод: ${_driveCtrl.text.trim()}',
                                    style: const TextStyle(
                                      color: Colors.white54,
                                      fontSize: 11,
                                    ),
                                  ),
                                if (transmissionModelCategories.contains(
                                      part['category'],
                                    ) &&
                                    _transmissionModelCtrl.text
                                        .trim()
                                        .isNotEmpty)
                                  Text(
                                    'Модель трансмиссии: ${_transmissionModelCtrl.text.trim()}',
                                    style: const TextStyle(
                                      color: Colors.white54,
                                      fontSize: 11,
                                    ),
                                  ),
                              ],
                            ),
                          ),
                          Text(
                            '${part['price'] ?? 0} ₽',
                            style: const TextStyle(
                              color: Color(0xFF4F8EF7),
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }

  Widget _limitButton(int val, {String? label}) {
    final isSelected = _displayLimit == val;
    return Padding(
      padding: const EdgeInsets.only(right: 6),
      child: InkWell(
        onTap: () => setState(() => _displayLimit = val),
        borderRadius: BorderRadius.circular(4),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          decoration: BoxDecoration(
            color: isSelected
                ? const Color(0xFF4F8EF7).withValues(alpha: 0.2)
                : Colors.transparent,
            border: Border.all(
              color: isSelected ? const Color(0xFF4F8EF7) : Colors.white10,
              width: 1,
            ),
            borderRadius: BorderRadius.circular(4),
          ),
          child: Text(
            label ?? val.toString(),
            style: TextStyle(
              color: isSelected ? const Color(0xFF4F8EF7) : Colors.white70,
              fontSize: 11,
              fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
            ),
          ),
        ),
      ),
    );
  }
}
