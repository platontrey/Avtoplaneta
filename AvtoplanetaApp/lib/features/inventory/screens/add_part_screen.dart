import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:image_cropper/image_cropper.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import '../../../core/api/api_client.dart';
import '../data/part_catalog.dart';
import '../providers/inventory_provider.dart';
import '../providers/part_catalog_provider.dart';

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
  final _addressCtrl = TextEditingController();
  final _salesmanCtrl = TextEditingController();

  // Управление фото
  final List<String> _existingPhotos = [];
  final List<File> _newPhotos = [];
  final List<String> _photosToDelete = [];

  final ImagePicker _picker = ImagePicker();
  bool _selectingImage = false;

  @override
  void initState() {
    super.initState();
    _categoryCtrl.addListener(_onCategoryChanged);
    _loadCatalog();
    if (widget.editId != null) {
      _loadPart();
    }
    _recoverLostImages();
  }

  PartCatalog? _partCatalog;
  String? _catalogError;

  List<String> get _categories =>
      _partCatalog?.partFormCategories.map((item) => item.name).toList() ??
      const [];
  List<String> get _transmissionOptions =>
      _partCatalog?.optionsForAttribute('transmission') ?? const [];
  List<String> get _driveOptions =>
      _partCatalog?.optionsForAttribute('drive') ?? const [];

  Set<String> get _visibleSpecificationFields {
    final catalog = _partCatalog;
    if (catalog == null) {
      return const {};
    }
    // Editing must expose the same complete set of existing fields as a new
    // part form; category templates only control the compact add flow.
    return catalog.attributes.map((attribute) => attribute.code).toSet();
  }

  Future<void> _loadCatalog() async {
    try {
      final catalog = await ref.read(partCatalogProvider.future);
      if (mounted) {
        setState(() => _partCatalog = catalog);
      }
    } catch (error) {
      if (mounted) {
        setState(() => _catalogError = error.toString());
      }
    }
  }

  bool _shows(String field) => _visibleSpecificationFields.contains(field);

  void _onCategoryChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  Future<void> _loadPart() async {
    try {
      final response = await apiClient.dio.get(
        '/api/v1/parts/item/${widget.editId}',
      );
      final data = response.data as Map<String, dynamic>;
      if (!mounted) {
        return;
      }
      setState(() {
        _nameCtrl.text = data['name'] ?? '';
        _descCtrl.text = data['description'] ?? '';
        _categoryCtrl.text = data['category'] ?? '';
        _priceCtrl.text = (data['price'] ?? '').toString();
        _quantityCtrl.text = (data['quantity'] ?? '1').toString();
        _brandCtrl.text = data['brand'] ?? '';
        _modelCtrl.text = data['model'] ?? '';
        _bodyBrandCtrl.text = data['body_brand'] ?? data['bodyBrand'] ?? '';
        _engineBrandCtrl.text = data['engine_brand'] ?? data['engineBrand'] ?? '';
        _carReleaseDateCtrl.text = data['car_release_date'] ?? data['carReleaseDate'] ?? '';
        _frontRearCtrl.text = data['front_rear'] ?? data['frontRear'] ?? '';
        _leftRightCtrl.text = data['left_right'] ?? data['leftRight'] ?? '';
        _topBottomCtrl.text = data['top_bottom'] ?? data['topBottom'] ?? '';
        _numberCtrl.text = data['number'] ?? '';
        _manufacturerCtrl.text = data['manufacturer'] ?? '';
        _manufacturerCodeCtrl.text = data['manufacturer_code'] ?? data['manufacturerCode'] ?? '';
        _colorCtrl.text = data['color'] ?? '';
        _conditionCtrl.text = data['condition'] ?? '';
        _defectCtrl.text = data['defect'] ?? '';
        _transmissionCtrl.text = data['transmission'] ?? '';
        _transmissionModelCtrl.text = data['transmission_model'] ?? data['transmissionModel'] ?? '';
        _driveCtrl.text = data['drive'] ?? '';
        _wearPercentageCtrl.text = data['wear_percentage'] ?? data['wearPercentage'] ?? '';
        _seasonCtrl.text = data['season'] ?? '';
        _diameterCtrl.text = data['diameter'] ?? '';
        _widthCtrl.text = data['width'] ?? '';
        _profileCtrl.text = data['profile'] ?? '';
        _tireQuantityCtrl.text = data['tire_quantity'] ?? data['tireQuantity'] ?? '';
        _drillingCtrl.text = data['drilling'] ?? '';
        _offsetCtrl.text = data['offset'] ?? '';
        _centerHoleDiameterCtrl.text = data['center_hole_diameter'] ?? data['centerHoleDiameter'] ?? '';
        _tireModelCtrl.text = data['tire_model'] ?? data['tireModel'] ?? '';
        _oemCtrl.text = data['oem_code'] ?? data['oemCode'] ?? '';
        _supplierCtrl.text = data['supplier_code'] ?? data['supplierCode'] ?? '';
        _vinCtrl.text = data['vin'] ?? '';
        _locationCtrl.text = data['location'] ?? '';
        _addressCtrl.text = data['address'] ?? '';
        _salesmanCtrl.text = data['salesman'] ?? '';

        _existingPhotos.clear();
        final photosRaw = data['photos'];
        if (photosRaw is List) {
          _existingPhotos.addAll(photosRaw.map((e) => e.toString()));
        } else if (data['photo']?.toString().isNotEmpty == true) {
          _existingPhotos.add(data['photo'].toString());
        }
      });
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось загрузить запчасть: $error')),
        );
      }
    }
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
      _addressCtrl,
      _salesmanCtrl,
    ]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _pickImage(ImageSource source) async {
    if (_selectingImage) {
      return;
    }
    setState(() => _selectingImage = true);
    try {
      final XFile? image = await _picker.pickImage(
        source: source,
        imageQuality: 85,
        maxWidth: 2048,
        maxHeight: 2048,
      );
      if (image == null || !mounted) {
        return;
      }
      await _cropAndAddImage(image);
    } catch (e) {
      _showImageError(e);
    } finally {
      if (mounted) {
        setState(() => _selectingImage = false);
      }
    }
  }

  Future<void> _recoverLostImages() async {
    try {
      final response = await _picker.retrieveLostData();
      if (response.isEmpty || !mounted) {
        return;
      }
      if (response.exception != null) {
        _showImageError(response.exception!);
        return;
      }
      final files = response.files;
      if (files == null || files.isEmpty) {
        return;
      }
      await _cropAndAddImage(files.first);
    } catch (error) {
      _showImageError(error);
    }
  }

  Future<void> _cropAndAddImage(XFile image) async {
    final cropped = await ImageCropper().cropImage(
      sourcePath: image.path,
      compressQuality: 85,
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
    if (cropped == null || !mounted) {
      return;
    }
    setState(() => _newPhotos.add(File(cropped.path)));
  }

  void _showImageError(Object error) {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text('Ошибка при выборе фото: $error')));
  }

  Future<void> _showPhotoOptions() async {
    final source = await showModalBottomSheet<ImageSource>(
      context: context,
      builder: (_) => SafeArea(
        child: Wrap(
          children: [
            ListTile(
              leading: const Icon(Icons.camera_alt_outlined),
              title: const Text('Сделать снимок (Камера)'),
              onTap: () => Navigator.pop(context, ImageSource.camera),
            ),
            ListTile(
              leading: const Icon(Icons.photo_library_outlined),
              title: const Text('Выбрать из галереи'),
              onTap: () => Navigator.pop(context, ImageSource.gallery),
            ),
          ],
        ),
      ),
    );
    if (source != null && mounted) {
      await _pickImage(source);
    }
  }

  String _parseDioError(dynamic e) {
    if (e is DioException) {
      final res = e.response?.data;
      if (res is Map) {
        if (res['error'] != null) return res['error'].toString();
        if (res['message'] != null) return res['message'].toString();
      } else if (res is String && res.isNotEmpty) {
        return res;
      }
      if (e.message != null && e.message!.isNotEmpty) {
        return e.message!;
      }
    }
    return e.toString();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }
    setState(() => _loading = true);
    try {
      final data = {
        'name': _nameCtrl.text.trim(),
        'description': _descCtrl.text.trim(),
        'category': _categoryCtrl.text.trim(),
        'price': double.tryParse(_priceCtrl.text.replaceAll(',', '.')) ?? 0.0,
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
        'address': _addressCtrl.text.trim(),
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
        final fileName = file.path.split(RegExp(r'[/\\]')).last;
        final extension = fileName.contains('.')
            ? fileName.split('.').last.toLowerCase()
            : 'jpg';
        final mimeType = extension == 'png'
            ? 'png'
            : (extension == 'webp' ? 'webp' : 'jpeg');

        final formData = FormData.fromMap({
          'photo': await MultipartFile.fromFile(
            file.path,
            filename: fileName,
            contentType: DioMediaType('image', mimeType),
          ),
        });
        await apiClient.dio.post(
          '/api/uploadpartphoto/$partId',
          data: formData,
        );
      }

      ref.invalidate(inventoryProvider);
      ref.invalidate(partProvider(partId));
      if (mounted) {
        context.go('/inventory');
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Ошибка сохранения: ${_parseDioError(e)}')));
      }
    } finally {
      if (mounted) {
        setState(() => _loading = false);
      }
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
            _field(_bodyBrandCtrl, 'Марка кузова'),
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
            if (_shows('drive')) _driveField(),
            if (_shows('color')) _field(_colorCtrl, 'Цвет кузовных деталей'),
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
            _field(_vinCtrl, 'VIN / Номер кузова'),

            _section('Расположение'),
            _field(_locationCtrl, 'Место хранения'),
            _field(_addressCtrl, 'Адрес склада'),
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
                  onTap: _selectingImage ? null : _showPhotoOptions,
                  borderRadius: BorderRadius.circular(12),
                  child: SizedBox(
                    width: 100,
                    height: 100,
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        if (_selectingImage)
                          const SizedBox(
                            width: 22,
                            height: 22,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        else
                          const Icon(
                            Icons.add_a_photo_outlined,
                            color: Colors.white70,
                          ),
                        const SizedBox(height: 4),
                        Text(
                          _selectingImage ? 'Обработка' : 'Добавить',
                          style: const TextStyle(
                            color: Colors.white70,
                            fontSize: 12,
                          ),
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
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          DropdownButtonFormField<String>(
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
          if (_catalogError != null)
            Padding(
              padding: const EdgeInsets.only(top: 6),
              child: Text(
                'Не удалось загрузить каталог: $_catalogError',
                style: const TextStyle(color: Colors.redAccent, fontSize: 12),
              ),
            ),
        ],
      ),
    );
  }

  Widget _transmissionField() {
    final current = _transmissionCtrl.text.trim();
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: DropdownButtonFormField<String>(
        key: ValueKey('transmission-$current'),
        initialValue: _transmissionOptions.contains(current) ? current : null,
        isExpanded: true,
        decoration: const InputDecoration(labelText: 'Тип трансмиссии'),
        items: _transmissionOptions
            .map((type) => DropdownMenuItem(value: type, child: Text(type)))
            .toList(),
        onChanged: (value) => _transmissionCtrl.text = value ?? '',
      ),
    );
  }

  Widget _driveField() {
    final current = _driveCtrl.text.trim();
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: DropdownButtonFormField<String>(
        key: ValueKey('drive-$current'),
        initialValue: _driveOptions.contains(current) ? current : null,
        isExpanded: true,
        decoration: const InputDecoration(labelText: 'Привод'),
        items: _driveOptions
            .map((type) => DropdownMenuItem(value: type, child: Text(type)))
            .toList(),
        onChanged: (value) => _driveCtrl.text = value ?? '',
      ),
    );
  }
}
