import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:image_cropper/image_cropper.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
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
  final _bodyBrandCtrl = TextEditingController();
  final _engineBrandCtrl = TextEditingController();
  final _carReleaseDateCtrl = TextEditingController();
  final _frontRearCtrl = TextEditingController();
  final _leftRightCtrl = TextEditingController();
  final _topBottomCtrl = TextEditingController();
  final _numberCtrl = TextEditingController();
  final _manufacturerCtrl = TextEditingController();
  final _manufacturerCodeCtrl = TextEditingController();
  final _colorCtrl = TextEditingController();
  final _conditionCtrl = TextEditingController();
  final _defectCtrl = TextEditingController();
  final _transmissionCtrl = TextEditingController();
  final _transmissionModelCtrl = TextEditingController();
  final _driveCtrl = TextEditingController();
  final _wearPercentageCtrl = TextEditingController();
  final _seasonCtrl = TextEditingController();
  final _diameterCtrl = TextEditingController();
  final _widthCtrl = TextEditingController();
  final _profileCtrl = TextEditingController();
  final _tireQuantityCtrl = TextEditingController();
  final _drillingCtrl = TextEditingController();
  final _offsetCtrl = TextEditingController();
  final _centerHoleDiameterCtrl = TextEditingController();
  final _tireModelCtrl = TextEditingController();
  final _oemCtrl = TextEditingController();
  final _supplierCtrl = TextEditingController();
  final _vinCtrl = TextEditingController();
  final _locationCtrl = TextEditingController();
  final _salesmanCtrl = TextEditingController();

  // Управление фото
  final List<String> _existingPhotos = [];
  final List<File> _newPhotos = [];
  final List<String> _photosToDelete = [];

  final ImagePicker _picker = ImagePicker();

  @override
  void initState() {
    super.initState();
    _categoryCtrl.addListener(_onCategoryChanged);
    if (widget.editId != null) _loadPart();
  }

  static const Set<String> _allSpecificationFields = {
    'body_brand',
    'engine_brand',
    'car_release_date',
    'front_rear',
    'left_right',
    'top_bottom',
    'number',
    'manufacturer',
    'manufacturer_code',
    'oem_code',
    'color',
    'condition',
    'supplier_code',
    'defect',
    'transmission',
    'transmission_model',
    'drive',
    'wear_percentage',
    'season',
    'diameter',
    'width',
    'profile',
    'tire_quantity',
    'drilling',
    'offset',
    'center_hole_diameter',
    'tire_model',
  };

  static const List<String> _categories = [
    'Тормоза',
    'Двигатель',
    'Подвеска',
    'Подвеска ДВС/КПП',
    'Подвеска передних колес',
    'Подвеска задних колес',
    'Электрика',
    'Кузов',
    'Кузов снаружи',
    'Интерьер',
    'Трансмиссия',
    'Система охлаждения и отопления',
    'Система выхлопа (Глушитель)',
    'Система рулевого управления',
    'Рулевое управление',
    'Система фильтрации (Фильтры)',
    'Шины и диски',
    'Автохимия и масла',
    'Аксессуары и тюннинг',
    'Другое',
  ];

  static const List<String> _transmissionTypes = [
    'МКПП',
    'АКПП',
    'Роботизированная',
    'Вариатор',
  ];

  static const Map<String, Set<String>> _specificationFieldsByCategory = {
    'Тормоза': {
      'front_rear',
      'left_right',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
      'drive',
      'wear_percentage',
    },
    'Двигатель': {
      'engine_brand',
      'car_release_date',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
      'drive',
    },
    'Подвеска': {
      'front_rear',
      'left_right',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
    },
    'Подвеска ДВС/КПП': {
      'front_rear',
      'left_right',
      'top_bottom',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
      'transmission_model',
      'drive',
    },
    'Подвеска передних колес': {
      'front_rear',
      'left_right',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
      'transmission_model',
      'drive',
    },
    'Подвеска задних колес': {
      'front_rear',
      'left_right',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
      'drive',
    },
    'Электрика': {
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
      'drive',
    },
    'Кузов': {
      'body_brand',
      'front_rear',
      'left_right',
      'top_bottom',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'color',
      'condition',
      'supplier_code',
      'defect',
    },
    'Кузов снаружи': {
      'body_brand',
      'front_rear',
      'left_right',
      'top_bottom',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'color',
      'condition',
      'supplier_code',
      'defect',
    },
    'Интерьер': {
      'top_bottom',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'color',
      'condition',
      'supplier_code',
      'defect',
    },
    'Трансмиссия': {
      'front_rear',
      'left_right',
      'transmission',
      'transmission_model',
      'drive',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
    },
    'Система охлаждения и отопления': {
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
    },
    'Система выхлопа (Глушитель)': {
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
      'drive',
    },
    'Система рулевого управления': {
      'front_rear',
      'left_right',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
      'drive',
    },
    'Рулевое управление': {
      'front_rear',
      'left_right',
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
      'drive',
    },
    'Система фильтрации (Фильтры)': {
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
    },
    'Шины и диски': {
      'front_rear',
      'left_right',
      'diameter',
      'width',
      'profile',
      'tire_quantity',
      'drilling',
      'offset',
      'center_hole_diameter',
      'tire_model',
      'season',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
      'defect',
      'wear_percentage',
    },
    'Автохимия и масла': {
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'condition',
      'supplier_code',
    },
    'Аксессуары и тюннинг': {
      'number',
      'manufacturer',
      'manufacturer_code',
      'oem_code',
      'color',
      'condition',
      'supplier_code',
      'defect',
    },
  };

  Set<String> get _visibleSpecificationFields {
    final category = _categoryCtrl.text.trim();
    if (category.isEmpty || category == 'Другое') {
      return _allSpecificationFields;
    }
    return _specificationFieldsByCategory[category] ?? const <String>{};
  }

  bool _shows(String field) => _visibleSpecificationFields.contains(field);

  void _onCategoryChanged() {
    if (mounted) setState(() {});
  }

  Future<void> _loadPart() async {
    try {
      final response = await apiClient.dio.get(
        '/api/v1/parts/item/${widget.editId}',
      );
      final data = response.data as Map<String, dynamic>;
      setState(() {
        _nameCtrl.text = data['name'] ?? '';
        _descCtrl.text = data['description'] ?? '';
        _categoryCtrl.text = data['category'] ?? '';
        _priceCtrl.text = (data['price'] ?? '').toString();
        _quantityCtrl.text = (data['quantity'] ?? '1').toString();
        _brandCtrl.text = data['brand'] ?? '';
        _modelCtrl.text = data['model'] ?? '';
        _bodyBrandCtrl.text = data['body_brand'] ?? '';
        _engineBrandCtrl.text = data['engine_brand'] ?? '';
        _carReleaseDateCtrl.text = data['car_release_date'] ?? '';
        _frontRearCtrl.text = data['front_rear'] ?? '';
        _leftRightCtrl.text = data['left_right'] ?? '';
        _topBottomCtrl.text = data['top_bottom'] ?? '';
        _numberCtrl.text = data['number'] ?? '';
        _manufacturerCtrl.text = data['manufacturer'] ?? '';
        _manufacturerCodeCtrl.text = data['manufacturer_code'] ?? '';
        _colorCtrl.text = data['color'] ?? '';
        _conditionCtrl.text = data['condition'] ?? '';
        _defectCtrl.text = data['defect'] ?? '';
        _transmissionCtrl.text = data['transmission'] ?? '';
        _transmissionModelCtrl.text = data['transmission_model'] ?? '';
        _driveCtrl.text = data['drive'] ?? '';
        _wearPercentageCtrl.text = data['wear_percentage'] ?? '';
        _seasonCtrl.text = data['season'] ?? '';
        _diameterCtrl.text = data['diameter'] ?? '';
        _widthCtrl.text = data['width'] ?? '';
        _profileCtrl.text = data['profile'] ?? '';
        _tireQuantityCtrl.text = data['tire_quantity'] ?? '';
        _drillingCtrl.text = data['drilling'] ?? '';
        _offsetCtrl.text = data['offset'] ?? '';
        _centerHoleDiameterCtrl.text = data['center_hole_diameter'] ?? '';
        _tireModelCtrl.text = data['tire_model'] ?? '';
        _oemCtrl.text = data['oem_code'] ?? '';
        _supplierCtrl.text = data['supplier_code'] ?? '';
        _vinCtrl.text = data['vin'] ?? '';
        _locationCtrl.text = data['location'] ?? '';
        _salesmanCtrl.text = data['salesman'] ?? '';

        final photosRaw = data['photos'];
        if (photosRaw is List) {
          _existingPhotos.addAll(photosRaw.map((e) => e.toString()));
        } else if (data['photo']?.toString().isNotEmpty == true) {
          _existingPhotos.add(data['photo'].toString());
        }
      });
    } catch (_) {}
  }

  @override
  void dispose() {
    _categoryCtrl.removeListener(_onCategoryChanged);
    for (final c in [
      _nameCtrl,
      _descCtrl,
      _categoryCtrl,
      _priceCtrl,
      _quantityCtrl,
      _brandCtrl,
      _modelCtrl,
      _bodyBrandCtrl,
      _engineBrandCtrl,
      _carReleaseDateCtrl,
      _frontRearCtrl,
      _leftRightCtrl,
      _topBottomCtrl,
      _numberCtrl,
      _manufacturerCtrl,
      _manufacturerCodeCtrl,
      _colorCtrl,
      _conditionCtrl,
      _defectCtrl,
      _transmissionCtrl,
      _transmissionModelCtrl,
      _driveCtrl,
      _wearPercentageCtrl,
      _seasonCtrl,
      _diameterCtrl,
      _widthCtrl,
      _profileCtrl,
      _tireQuantityCtrl,
      _drillingCtrl,
      _offsetCtrl,
      _centerHoleDiameterCtrl,
      _tireModelCtrl,
      _oemCtrl,
      _supplierCtrl,
      _vinCtrl,
      _locationCtrl,
      _salesmanCtrl,
    ]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _pickImage(ImageSource source) async {
    try {
      final XFile? image = await _picker.pickImage(
        source: source,
        imageQuality: 85,
      );
      if (image == null) return;

      final cropped = await ImageCropper().cropImage(
        sourcePath: image.path,
        uiSettings: [
          AndroidUiSettings(
            toolbarTitle: 'Редактор фото',
            toolbarColor: const Color(0xFF1A1A2E),
            toolbarWidgetColor: Colors.white,
            initAspectRatio: CropAspectRatioPreset.square,
            lockAspectRatio: true,
          ),
          IOSUiSettings(title: 'Редактор фото', aspectRatioLockEnabled: true),
        ],
      );

      if (cropped != null) {
        setState(() {
          _newPhotos.add(File(cropped.path));
        });
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text('Ошибка при выборе фото: $e')));
    }
  }

  void _showPhotoOptions() {
    showModalBottomSheet(
      context: context,
      builder: (_) => SafeArea(
        child: Wrap(
          children: [
            ListTile(
              leading: const Icon(Icons.camera_alt_outlined),
              title: const Text('Сделать снимок (Камера)'),
              onTap: () {
                Navigator.pop(context);
                _pickImage(ImageSource.camera);
              },
            ),
            ListTile(
              leading: const Icon(Icons.photo_library_outlined),
              title: const Text('Выбрать из галереи'),
              onTap: () {
                Navigator.pop(context);
                _pickImage(ImageSource.gallery);
              },
            ),
          ],
        ),
      ),
    );
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
        'body_brand': _bodyBrandCtrl.text.trim(),
        'engine_brand': _engineBrandCtrl.text.trim(),
        'car_release_date': _carReleaseDateCtrl.text.trim(),
        'front_rear': _frontRearCtrl.text.trim(),
        'left_right': _leftRightCtrl.text.trim(),
        'top_bottom': _topBottomCtrl.text.trim(),
        'number': _numberCtrl.text.trim(),
        'manufacturer': _manufacturerCtrl.text.trim(),
        'manufacturer_code': _manufacturerCodeCtrl.text.trim(),
        'color': _colorCtrl.text.trim(),
        'condition': _conditionCtrl.text.trim(),
        'defect': _defectCtrl.text.trim(),
        'transmission': _transmissionCtrl.text.trim(),
        'transmission_model': _transmissionModelCtrl.text.trim(),
        'drive': _driveCtrl.text.trim(),
        'wear_percentage': _wearPercentageCtrl.text.trim(),
        'season': _seasonCtrl.text.trim(),
        'diameter': _diameterCtrl.text.trim(),
        'width': _widthCtrl.text.trim(),
        'profile': _profileCtrl.text.trim(),
        'tire_quantity': _tireQuantityCtrl.text.trim(),
        'drilling': _drillingCtrl.text.trim(),
        'offset': _offsetCtrl.text.trim(),
        'center_hole_diameter': _centerHoleDiameterCtrl.text.trim(),
        'tire_model': _tireModelCtrl.text.trim(),
        'oem_code': _oemCtrl.text.trim(),
        'supplier_code': _supplierCtrl.text.trim(),
        'vin': _vinCtrl.text.trim(),
        'location': _locationCtrl.text.trim(),
        'salesman': _salesmanCtrl.text.trim(),
      };

      int partId;
      if (widget.editId != null) {
        partId = widget.editId!;
        await apiClient.dio.put('/api/v1/parts/$partId', data: data);

        for (final photoPath in _photosToDelete) {
          await apiClient.dio.delete(
            '/api/deletepartphoto/$partId',
            queryParameters: {'photo': photoPath},
          );
        }
      } else {
        final response = await apiClient.dio.post('/api/v1/parts', data: data);
        partId = (response.data['id'] as num).toInt();
      }

      for (final file in _newPhotos) {
        final formData = FormData.fromMap({
          'photo': await MultipartFile.fromFile(
            file.path,
            filename: file.path.split('/').last,
          ),
        });
        await apiClient.dio.post(
          '/api/uploadpartphoto/$partId',
          data: formData,
        );
      }

      ref.invalidate(inventoryProvider);
      if (mounted) context.go('/inventory');
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Ошибка сохранения: $e')));
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
          widget.editId != null ? 'Редактировать' : 'Добавить запчасть',
        ),
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
                'Сохранить',
                style: TextStyle(color: Color(0xFF4F8EF7)),
              ),
            ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            _photoListSection(),
            _section('Основное'),
            _field(_nameCtrl, 'Название *', required: true),
            _categoryField(),
            _field(
              _priceCtrl,
              'Цена (₽) *',
              required: true,
              keyboard: TextInputType.number,
            ),
            _field(
              _quantityCtrl,
              'Количество *',
              required: true,
              keyboard: TextInputType.number,
            ),
            _field(_descCtrl, 'Описание', maxLines: 3),

            _section('Характеристики'),
            _field(_brandCtrl, 'Бренд'),
            _field(_modelCtrl, 'Модель'),
            if (_shows('body_brand')) _field(_bodyBrandCtrl, 'Марка кузова'),
            if (_shows('engine_brand'))
              _field(_engineBrandCtrl, 'Марка двигателя'),
            if (_shows('car_release_date'))
              _field(_carReleaseDateCtrl, 'Год выпуска'),
            if (_shows('front_rear')) _field(_frontRearCtrl, 'Перед / зад'),
            if (_shows('left_right')) _field(_leftRightCtrl, 'Лево / право'),
            if (_shows('top_bottom')) _field(_topBottomCtrl, 'Верх / низ'),
            if (_shows('number')) _field(_numberCtrl, 'Номер детали'),
            if (_shows('manufacturer'))
              _field(_manufacturerCtrl, 'Производитель'),
            if (_shows('manufacturer_code'))
              _field(_manufacturerCodeCtrl, 'Код производителя'),
            if (_shows('oem_code')) _field(_oemCtrl, 'OEM код'),
            if (_shows('color')) _field(_colorCtrl, 'Цвет'),
            if (_shows('condition')) _field(_conditionCtrl, 'Состояние'),
            if (_shows('supplier_code'))
              _field(_supplierCtrl, 'Код поставщика'),
            if (_shows('defect')) _field(_defectCtrl, 'Дефект'),
            if (_shows('transmission')) _transmissionField(),
            if (_shows('transmission_model'))
              _field(
                _transmissionModelCtrl,
                'Модель трансмиссии',
                hint: 'Введите номер трансмиссии',
              ),
            if (_shows('drive')) _field(_driveCtrl, 'Привод'),
            if (_shows('wear_percentage'))
              _field(
                _wearPercentageCtrl,
                'Процент износа (%)',
                keyboard: TextInputType.number,
              ),
            if (_shows('season')) _field(_seasonCtrl, 'Сезон'),
            if (_shows('diameter')) _field(_diameterCtrl, 'Диаметр'),
            if (_shows('width')) _field(_widthCtrl, 'Ширина'),
            if (_shows('profile')) _field(_profileCtrl, 'Профиль'),
            if (_shows('tire_quantity'))
              _field(
                _tireQuantityCtrl,
                'Количество шин',
                keyboard: TextInputType.number,
              ),
            if (_shows('drilling')) _field(_drillingCtrl, 'Сверловка'),
            if (_shows('offset')) _field(_offsetCtrl, 'Вылет'),
            if (_shows('center_hole_diameter'))
              _field(_centerHoleDiameterCtrl, 'Диаметр центрального отверстия'),
            if (_shows('tire_model')) _field(_tireModelCtrl, 'Модель шины'),

            _section('Идентификация'),
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

  Widget _photoListSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _section('Фотографии'),
        SizedBox(
          height: 100,
          child: ListView(
            scrollDirection: Axis.horizontal,
            children: [
              Card(
                color: const Color(0xFF16213E),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                  side: const BorderSide(color: Colors.white24, width: 1),
                ),
                child: InkWell(
                  onTap: _showPhotoOptions,
                  borderRadius: BorderRadius.circular(12),
                  child: const SizedBox(
                    width: 100,
                    height: 100,
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.add_a_photo_outlined, color: Colors.white70),
                        SizedBox(height: 4),
                        Text(
                          'Добавить',
                          style: TextStyle(color: Colors.white70, fontSize: 12),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 8),

              ..._newPhotos.asMap().entries.map((entry) {
                final idx = entry.key;
                final file = entry.value;
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: Stack(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(12),
                        child: Image.file(
                          file,
                          width: 100,
                          height: 100,
                          fit: BoxFit.cover,
                        ),
                      ),
                      Positioned(
                        top: 2,
                        right: 2,
                        child: CircleAvatar(
                          radius: 12,
                          backgroundColor: Colors.black54,
                          child: IconButton(
                            padding: EdgeInsets.zero,
                            icon: const Icon(
                              Icons.close,
                              size: 14,
                              color: Colors.white,
                            ),
                            onPressed: () {
                              setState(() {
                                _newPhotos.removeAt(idx);
                              });
                            },
                          ),
                        ),
                      ),
                    ],
                  ),
                );
              }),

              ..._existingPhotos.asMap().entries.map((entry) {
                final idx = entry.key;
                final path = entry.value;
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: Stack(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(12),
                        child: CachedNetworkImage(
                          imageUrl: apiClient.resolveUrl(path),
                          width: 100,
                          height: 100,
                          fit: BoxFit.cover,
                          placeholder: (context, url) => Container(
                            width: 100,
                            height: 100,
                            color: const Color(0xFF16213E),
                            child: const Center(
                              child: CircularProgressIndicator(strokeWidth: 2),
                            ),
                          ),
                          errorWidget: (context, url, error) => Container(
                            width: 100,
                            height: 100,
                            color: const Color(0xFF16213E),
                            child: const Icon(
                              Icons.broken_image_outlined,
                              color: Colors.white24,
                            ),
                          ),
                        ),
                      ),
                      Positioned(
                        top: 2,
                        right: 2,
                        child: CircleAvatar(
                          radius: 12,
                          backgroundColor: Colors.black54,
                          child: IconButton(
                            padding: EdgeInsets.zero,
                            icon: const Icon(
                              Icons.close,
                              size: 14,
                              color: Colors.white,
                            ),
                            onPressed: () {
                              setState(() {
                                _photosToDelete.add(path);
                                _existingPhotos.removeAt(idx);
                              });
                            },
                          ),
                        ),
                      ),
                    ],
                  ),
                );
              }),
            ],
          ),
        ),
      ],
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
        letterSpacing: 0.5,
      ),
    ),
  );

  Widget _field(
    TextEditingController ctrl,
    String label, {
    bool required = false,
    int maxLines = 1,
    TextInputType keyboard = TextInputType.text,
    String? hint,
  }) => Padding(
    padding: const EdgeInsets.only(bottom: 12),
    child: TextFormField(
      controller: ctrl,
      maxLines: maxLines,
      keyboardType: keyboard,
      style: const TextStyle(color: Colors.white),
      decoration: InputDecoration(labelText: label, hintText: hint),
      validator: required
          ? (v) => (v == null || v.trim().isEmpty) ? 'Обязательное поле' : null
          : null,
    ),
  );

  Widget _categoryField() {
    final current = _categoryCtrl.text.trim();
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: DropdownButtonFormField<String>(
        key: ValueKey(current),
        initialValue: _categories.contains(current) ? current : null,
        isExpanded: true,
        decoration: const InputDecoration(labelText: 'Категория *'),
        items: _categories
            .map(
              (category) => DropdownMenuItem(
                value: category,
                child: Text(category, overflow: TextOverflow.ellipsis),
              ),
            )
            .toList(),
        onChanged: (value) => _categoryCtrl.text = value ?? '',
        validator: (value) =>
            value == null || value.isEmpty ? 'Обязательное поле' : null,
      ),
    );
  }

  Widget _transmissionField() {
    final current = _transmissionCtrl.text.trim();
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: DropdownButtonFormField<String>(
        key: ValueKey('transmission-$current'),
        initialValue: _transmissionTypes.contains(current) ? current : null,
        isExpanded: true,
        decoration: const InputDecoration(labelText: 'Тип трансмиссии'),
        items: _transmissionTypes
            .map((type) => DropdownMenuItem(value: type, child: Text(type)))
            .toList(),
        onChanged: (value) => _transmissionCtrl.text = value ?? '',
      ),
    );
  }
}
