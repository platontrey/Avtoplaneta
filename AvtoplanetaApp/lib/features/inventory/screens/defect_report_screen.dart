import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/utils/formatters.dart';
import '../data/defect_report_api.dart';
import '../data/vehicle_catalog.dart';
import '../providers/inventory_provider.dart';
import '../providers/part_catalog_provider.dart';
import '../providers/vehicle_catalog_provider.dart';
import '../widgets/vehicle_picker_field.dart';

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

  String? get _selectedBrand {
    final value = _brandCtrl.text.trim();
    return value.isEmpty ? null : value;
  }
  String? _selectedTransmission;
  String? _selectedInteriorColor;
  String? _selectedBodyColor;

  // Марки и модели приходят из серверного справочника (/api/vehicle-catalog),
  // общего с веб-клиентом. Раньше здесь лежал свой короткий список из семи марок.
  VehicleCatalog? _vehicleCatalog;
  final _brandCtrl = TextEditingController();

  List<String> get _brandOptions => _vehicleCatalog?.brandNames ?? const [];
  List<String> get _modelOptions =>
      _vehicleCatalog?.modelsOf(_brandCtrl.text) ?? const [];

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
  List<Map<String, dynamic>> _previewParts = const [];
  bool _previewLoading = false;
  String? _previewError;
  String? _catalogVersion;
  Timer? _previewDebounce;
  int _previewRequestId = 0;

  @override
  void initState() {
    super.initState();
    for (final controller in [
      _yearCtrl,
      _vinCtrl,
      _carReleasePeriodCtrl,
      _engineBrandCtrl,
      _bodyBrandCtrl,
      _transmissionModelCtrl,
      _driveCtrl,
    ]) {
      controller.addListener(_schedulePreview);
    }
    _loadVehicleCatalog();
  }

  /// Справочник марок не блокирует форму: если он не загрузился, поля марки и
  /// модели работают как обычный текстовый ввод.
  Future<void> _loadVehicleCatalog() async {
    try {
      final catalog = await ref.read(vehicleCatalogProvider.future);
      if (mounted) {
        setState(() => _vehicleCatalog = catalog);
      }
    } catch (_) {
      // Ничего: поля деградируют до ввода руками.
    }
  }

  @override
  void dispose() {
    _previewDebounce?.cancel();
    _brandCtrl.dispose();
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

  Map<String, dynamic> _buildPayload(String catalogVersion) => {
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
    'catalog_version': catalogVersion,
  };

  void _schedulePreview() {
    final catalogVersion = _catalogVersion;
    if (catalogVersion == null) return;

    _previewDebounce?.cancel();
    _previewDebounce = Timer(
      const Duration(milliseconds: 350),
      () => _loadPreview(catalogVersion),
    );
  }

  Future<void> _loadPreview(String catalogVersion) async {
    final requestId = ++_previewRequestId;
    if (mounted) {
      setState(() {
        _previewLoading = true;
        _previewError = null;
      });
    }

    try {
      final preview = await DefectReportApi.preview(
        _buildPayload(catalogVersion),
      );
      if (!mounted || requestId != _previewRequestId) return;
      setState(() {
        _previewParts = preview.parts;
        _previewLoading = false;
      });
    } catch (error) {
      if (!mounted || requestId != _previewRequestId) return;
      setState(() {
        _previewParts = const [];
        _previewLoading = false;
        _previewError = 'Не удалось построить превью: $error';
      });
    }
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
      await DefectReportApi.create(_buildPayload(catalog.version));

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

    if (_catalogVersion != catalog.version) {
      _catalogVersion = catalog.version;
      WidgetsBinding.instance.addPostFrameCallback((_) => _schedulePreview());
    }

    final allParts = _previewParts;
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
              onPressed: _previewError == null ? _submit : null,
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
            VehiclePickerField(
              controller: _brandCtrl,
              label: 'Бренд *',
              options: _brandOptions,
              required: true,
              onSelected: (_) {
                // Модель принадлежит марке: пересобираем список моделей и превью.
                setState(() {});
                _schedulePreview();
              },
            ),
            VehiclePickerField(
              controller: _modelCtrl,
              label: 'Модель *',
              options: _modelOptions,
              required: true,
              hint: _brandCtrl.text.trim().isEmpty ? 'Сначала выберите бренд' : null,
            ),
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
              keyboard: TextInputType.number,
              inputFormatters: const [CarReleasePeriodFormatter()],
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
              onChanged: (val) {
                setState(() => _selectedTransmission = val);
                _schedulePreview();
              },
            ),
            _buildTextField(
              _transmissionModelCtrl,
              'Модель трансмиссии',
              hint: 'Применяется ко всем запчастям ведомости',
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
              onChanged: (val) {
                setState(() => _selectedInteriorColor = val);
                _schedulePreview();
              },
            ),
            _buildDropdown(
              value: _selectedBodyColor,
              label: 'Цвет кузовных деталей',
              items: _availableColors,
              onChanged: (val) {
                setState(() => _selectedBodyColor = val);
                _schedulePreview();
              },
            ),

            const SizedBox(height: 12),
            _sectionTitle('Заметки'),
            _buildTextField(_descCtrl, 'Описание / Заметки', maxLines: 3),

            const SizedBox(height: 16),
            _buildPartsPreviewSection(
              filteredParts,
              totalParts: allParts.length,
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
    List<TextInputFormatter>? inputFormatters,
    String? hint,
    ValueChanged<String>? onChanged,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: TextFormField(
        controller: ctrl,
        maxLines: maxLines,
        keyboardType: keyboard,
        inputFormatters: inputFormatters,
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
        if (_previewLoading)
          const Padding(
            padding: EdgeInsets.only(bottom: 8),
            child: LinearProgressIndicator(),
          ),
        if (_previewError != null)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Text(
              _previewError!,
              style: const TextStyle(color: Colors.redAccent, fontSize: 12),
            ),
          ),
        TextField(
          style: const TextStyle(color: Colors.white, fontSize: 14),
          decoration: const InputDecoration(
            hintText: 'Поиск по превью деталей...',
            prefixIcon: Icon(LucideIcons.search, size: 18),
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
              ? Center(
                  child: Text(
                    _previewLoading
                        ? 'Формируем превью...'
                        : _previewError != null
                        ? 'Превью временно недоступно'
                        : 'Ничего не найдено',
                    style: const TextStyle(color: Colors.white30),
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
                                if ((part['car_release_period'] as String?)
                                        ?.isNotEmpty ??
                                    false)
                                  Text(
                                    'Период выпуска: ${part['car_release_period']}',
                                    style: const TextStyle(
                                      color: Colors.white54,
                                      fontSize: 11,
                                    ),
                                  ),
                                if ((part['transmission'] as String?)
                                        ?.isNotEmpty ??
                                    false)
                                  Text(
                                    'Трансмиссия: ${part['transmission']}',
                                    style: const TextStyle(
                                      color: Colors.white54,
                                      fontSize: 11,
                                    ),
                                  ),
                                if ((part['drive'] as String?)?.isNotEmpty ??
                                    false)
                                  Text(
                                    'Привод: ${part['drive']}',
                                    style: const TextStyle(
                                      color: Colors.white54,
                                      fontSize: 11,
                                    ),
                                  ),
                                if ((part['transmission_model'] as String?)
                                        ?.isNotEmpty ??
                                    false)
                                  Text(
                                    'Модель трансмиссии: ${part['transmission_model']}',
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
