import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/theme.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/part.dart';
import '../../../core/utils/formatters.dart';
import '../providers/inventory_provider.dart';
import 'vehicle_picker_field.dart';

enum PriceMode {
  set('PRICE_OPERATION_MODE_SET', 'Фиксированная'),
  percent('PRICE_OPERATION_MODE_PERCENT', 'Процент %'),
  offset('PRICE_OPERATION_MODE_OFFSET', 'Смещение ₽');

  final String protoName;
  final String label;
  const PriceMode(this.protoName, this.label);
}

enum FieldSection {
  basic('Основные'),
  specs('Характеристики'),
  tires('Шины и диски');

  final String label;
  const FieldSection(this.label);
}

enum FieldInputType {
  price,
  number,
  text,
  multiline,
  period,
  select,
  status,
  shelfAutocomplete,
}

class BulkFieldMeta {
  final String key;
  final String label;
  final FieldSection section;
  final String keywords;
  final FieldInputType type;
  final List<String> options;
  final String? hint;

  const BulkFieldMeta({
    required this.key,
    required this.label,
    required this.section,
    required this.keywords,
    this.type = FieldInputType.text,
    this.options = const [],
    this.hint,
  });
}

const List<String> defaultCategories = [
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

const List<String> defaultColors = [
  'Черный',
  'Белый',
  'Серебристый',
  'Серый',
  'Темно-серый',
  'Светло-серый',
  'Синий',
  'Темно-синий',
  'Светло-синий',
  'Красный',
  'Темно-красный',
  'Бордовый',
  'Зеленый',
  'Темно-зеленый',
  'Светло-зеленый',
  'Желтый',
  'Оранжевый',
  'Фиолетовый',
  'Коричневый',
  'Бежевый',
  'Золотой',
  'Бронзовый',
  'Перламутровый',
  'Металлик',
  'Матовый',
];

const List<String> defaultWarehouses = [
  'Томск, Профсоюзная 2/11',
  'г. Томск, пос. Кузовлево, ул. Заречная, 36',
];

// ignore: constant_identifier_names
const DEFAULT_CATEGORIES = defaultCategories;
// ignore: constant_identifier_names
const DEFAULT_COLORS = defaultColors;
// ignore: constant_identifier_names
const DEFAULT_WAREHOUSES = defaultWarehouses;

const List<BulkFieldMeta> allBulkFields = [
  // ─── 1. Основные (14 полей) ────────────────────────────────────────────────
  BulkFieldMeta(
    key: 'price',
    label: 'Изменение цены',
    section: FieldSection.basic,
    keywords: 'цена стоимость прайс скидка наценка руб',
    type: FieldInputType.price,
  ),
  BulkFieldMeta(
    key: 'quantity',
    label: 'Количество (шт.)',
    section: FieldSection.basic,
    keywords: 'количество остаток шт склад',
    type: FieldInputType.number,
    hint: 'Например: 1 (или 0)',
  ),
  BulkFieldMeta(
    key: 'name',
    label: 'Название детали',
    section: FieldSection.basic,
    keywords: 'название наименование имя деталь',
    hint: 'Например: Диск тормозной передний',
  ),
  BulkFieldMeta(
    key: 'brand',
    label: 'Бренд автомобиля',
    section: FieldSection.basic,
    keywords: 'бренд марка авто машина тойота',
    hint: 'Например: Toyota, BMW, Nissan',
  ),
  BulkFieldMeta(
    key: 'model',
    label: 'Модель автомобиля',
    section: FieldSection.basic,
    keywords: 'модель авто машина камри солярис',
    hint: 'Например: Camry, Corolla, X5',
  ),
  BulkFieldMeta(
    key: 'vin',
    label: 'VIN / Номер кузова',
    section: FieldSection.basic,
    keywords: 'вин vin кузов фрейм frame',
    hint: 'VIN или номер кузова',
  ),
  BulkFieldMeta(
    key: 'car_release_date',
    label: 'Год выпуска',
    section: FieldSection.basic,
    keywords: 'год выпуска дата',
    type: FieldInputType.number,
    hint: 'Например: 2012',
  ),
  BulkFieldMeta(
    key: 'car_release_period',
    label: 'Период выпуска',
    section: FieldSection.basic,
    keywords: 'период выпуска поколение года',
    type: FieldInputType.period,
    hint: 'Например: 2008-2015',
  ),
  BulkFieldMeta(
    key: 'category',
    label: 'Категория',
    section: FieldSection.basic,
    keywords: 'категория раздел группа',
    type: FieldInputType.select,
    options: DEFAULT_CATEGORIES,
  ),
  BulkFieldMeta(
    key: 'description',
    label: 'Описание / Заметки',
    section: FieldSection.basic,
    keywords: 'описание заметки комментарий примечание',
    type: FieldInputType.multiline,
    hint: 'Дополнительные примечания к запчастям',
  ),
  BulkFieldMeta(
    key: 'location',
    label: 'Местоположение (полка / стеллаж)',
    section: FieldSection.basic,
    keywords: 'местоположение локация стеллаж полка ряд ячейка',
    type: FieldInputType.shelfAutocomplete,
    hint: 'Стеллаж / ряд / полка',
  ),
  BulkFieldMeta(
    key: 'address',
    label: 'Адрес склада',
    section: FieldSection.basic,
    keywords: 'адрес склад томск кузовлево профсоюзная',
    type: FieldInputType.select,
    options: DEFAULT_WAREHOUSES,
  ),
  BulkFieldMeta(
    key: 'salesman',
    label: 'Ответственный продавец',
    section: FieldSection.basic,
    keywords: 'продавец менеджер сотрудник',
    hint: 'Имя ответственного продавца',
  ),
  BulkFieldMeta(
    key: 'status',
    label: 'Статус активности',
    section: FieldSection.basic,
    keywords: 'статус активность активный неактивный видимость',
    type: FieldInputType.status,
  ),

  // ─── 2. Характеристики (17 полей) ──────────────────────────────────────────
  BulkFieldMeta(
    key: 'body_brand',
    label: 'Марка кузова',
    section: FieldSection.specs,
    keywords: 'марка кузова кузов код ae100 w204 e39',
    hint: 'Например: AE100, W204, E39',
  ),
  BulkFieldMeta(
    key: 'engine_brand',
    label: 'Марка двигателя',
    section: FieldSection.specs,
    keywords: 'марка двигателя двс мотор 1jz 2gr m54',
    hint: 'Например: 1JZ-GTE, 2GR-FE, M54',
  ),
  BulkFieldMeta(
    key: 'transmission',
    label: 'Тип трансмиссии (КПП)',
    section: FieldSection.specs,
    keywords: 'трансмиссия кпп акпп мкпп вариатор робот редуктор',
    type: FieldInputType.select,
    options: ['АКПП', 'МКПП', 'Вариатор', 'Роботизированная', 'Редуктор'],
  ),
  BulkFieldMeta(
    key: 'transmission_model',
    label: 'Модель трансмиссии',
    section: FieldSection.specs,
    keywords: 'модель трансмиссии номер кпп коробка',
    hint: 'Модель / номер КПП',
  ),
  BulkFieldMeta(
    key: 'drive',
    label: 'Привод',
    section: FieldSection.specs,
    keywords: 'привод передний задний полный 4wd awd',
    type: FieldInputType.select,
    options: ['Передний', 'Задний', 'Полный (4WD)'],
  ),
  BulkFieldMeta(
    key: 'color',
    label: 'Цвет деталей',
    section: FieldSection.specs,
    keywords: 'цвет краска оттенок колор',
    type: FieldInputType.select,
    options: DEFAULT_COLORS,
  ),
  BulkFieldMeta(
    key: 'condition',
    label: 'Состояние',
    section: FieldSection.specs,
    keywords: 'состояние бу б/у новое новый',
    type: FieldInputType.select,
    options: ['Б/У', 'Новое'],
  ),
  BulkFieldMeta(
    key: 'defect',
    label: 'Дефект',
    section: FieldSection.specs,
    keywords: 'дефект повреждение царапина вмятина скол брак',
    hint: 'Описание дефекта (или пусто для очистки)',
  ),
  BulkFieldMeta(
    key: 'supplier_code',
    label: 'Код поставщика / партии',
    section: FieldSection.specs,
    keywords: 'код поставщика партия ведомость поставка',
    hint: 'Код партии или поставщика',
  ),
  BulkFieldMeta(
    key: 'front_rear',
    label: 'Перед / зад',
    section: FieldSection.specs,
    keywords: 'перед зад расположение f r',
    type: FieldInputType.select,
    options: ['F (Перед)', 'R (Зад)'],
  ),
  BulkFieldMeta(
    key: 'left_right',
    label: 'Право / лево',
    section: FieldSection.specs,
    keywords: 'право лево сторона l r',
    type: FieldInputType.select,
    options: ['L (Лево)', 'R (Право)'],
  ),
  BulkFieldMeta(
    key: 'top_bottom',
    label: 'Верх / низ',
    section: FieldSection.specs,
    keywords: 'верх низ расположение',
    hint: 'Например: Верх / Низ',
  ),
  BulkFieldMeta(
    key: 'number',
    label: 'Номер детали',
    section: FieldSection.specs,
    keywords: 'номер детали артикул партномер',
    hint: 'Номер детали / артикул',
  ),
  BulkFieldMeta(
    key: 'manufacturer',
    label: 'Производитель',
    section: FieldSection.specs,
    keywords: 'производитель завод фирма бренд',
    hint: 'Производитель (фирма, завод)',
  ),
  BulkFieldMeta(
    key: 'manufacturer_code',
    label: 'Код производителя',
    section: FieldSection.specs,
    keywords: 'код производителя кросс артикул',
    hint: 'Код производителя (артикул)',
  ),
  BulkFieldMeta(
    key: 'oem_code',
    label: 'OEM код',
    section: FieldSection.specs,
    keywords: 'oem оем оригинальный номер код оригинал',
    hint: 'Оригинальный OEM номер',
  ),
  BulkFieldMeta(
    key: 'wear_percentage',
    label: 'Процент износа (%)',
    section: FieldSection.specs,
    keywords: 'процент износа износ остаток выработка',
    type: FieldInputType.number,
    hint: 'Процент износа (0-100)',
  ),

  // ─── 3. Шины и диски (9 полей) ─────────────────────────────────────────────
  BulkFieldMeta(
    key: 'season',
    label: 'Сезон шин',
    section: FieldSection.tires,
    keywords: 'сезон зимние летние всесезонные шины резина',
    type: FieldInputType.select,
    options: [
      'Зимние (шипованные)',
      'Зимние (нешипованные)',
      'Летние',
      'Всесезонные',
    ],
  ),
  BulkFieldMeta(
    key: 'diameter',
    label: 'Диаметр (R)',
    section: FieldSection.tires,
    keywords: 'диаметр радиус r15 r16 r17 r18 диски шины',
    hint: 'Например: R16',
  ),
  BulkFieldMeta(
    key: 'width',
    label: 'Ширина',
    section: FieldSection.tires,
    keywords: 'ширина профиль диск шина 205 215 225',
    hint: 'Например: 205',
  ),
  BulkFieldMeta(
    key: 'profile',
    label: 'Профиль (%)',
    section: FieldSection.tires,
    keywords: 'профиль высота шина 55 60 65',
    hint: 'Например: 55',
  ),
  BulkFieldMeta(
    key: 'tire_quantity',
    label: 'Количество шин / дисков в комплекте',
    section: FieldSection.tires,
    keywords: 'количество шин дисков комплект пара 4шт',
    type: FieldInputType.number,
    hint: 'Например: 4',
  ),
  BulkFieldMeta(
    key: 'drilling',
    label: 'Сверловка (PCD)',
    section: FieldSection.tires,
    keywords: 'сверловка pcd разболтовка 5x114.3 4x100',
    hint: 'Например: 5x114.3',
  ),
  BulkFieldMeta(
    key: 'offset',
    label: 'Вылет (ET)',
    section: FieldSection.tires,
    keywords: 'вылет et et45 et35',
    hint: 'Например: ET45',
  ),
  BulkFieldMeta(
    key: 'center_hole_diameter',
    label: 'Диаметр ЦО (DIA)',
    section: FieldSection.tires,
    keywords: 'диаметр цо dia ступица центральное отверстие 67.1 60.1',
    hint: 'Например: 67.1',
  ),
  BulkFieldMeta(
    key: 'tire_model',
    label: 'Модель шины / диска',
    section: FieldSection.tires,
    keywords: 'модель шины резины диска nokian hakkapeliitta',
    hint: 'Например: Hakkapeliitta 8',
  ),
];

// ignore: constant_identifier_names
const ALL_BULK_FIELDS = allBulkFields;

Map<String, dynamic> buildBulkUpdatePayload({
  required Iterable<int> ids,
  String scopeType = 'BULK_SCOPE_TYPE_IDS',
  Map<String, dynamic>? filterQuery,
  bool confirmEntireCatalog = false,
  PriceMode? priceMode,
  double? priceValue,
  int? quantity,
  String? category,
  String? location,
  String? address,
  String? salesman,
  bool? status,
  String? bodyBrand,
  String? engineBrand,
  String? name,
  String? brand,
  String? model,
  String? vin,
  String? carReleaseDate,
  String? carReleasePeriod,
  String? description,
  String? transmission,
  String? transmissionModel,
  String? drive,
  String? color,
  String? condition,
  String? defect,
  String? supplierCode,
  String? frontRear,
  String? leftRight,
  String? topBottom,
  String? number,
  String? manufacturer,
  String? manufacturerCode,
  String? oemCode,
  String? wearPercentage,
  String? season,
  String? diameter,
  String? width,
  String? profile,
  String? tireQuantity,
  String? drilling,
  String? offset,
  String? centerHoleDiameter,
  String? tireModel,
  Map<String, dynamic>? customOperations,
}) {
  final operations = <String, dynamic>{};

  if (priceMode != null && priceValue != null) {
    operations['price'] = {
      'mode': priceMode.protoName,
      'value': priceValue,
    };
  }
  if (quantity != null) {
    operations['quantity'] = quantity;
  }
  if (status != null) {
    operations['status'] = status;
  }

  void addField(String key, String? val) {
    if (val != null && val.isNotEmpty) {
      operations[key] = val;
    }
  }

  addField('category', category);
  addField('location', location);
  addField('address', address);
  addField('salesman', salesman);
  addField('body_brand', bodyBrand);
  addField('engine_brand', engineBrand);
  addField('name', name);
  addField('brand', brand);
  addField('model', model);
  addField('vin', vin);
  addField('car_release_date', carReleaseDate);
  addField('car_release_period', carReleasePeriod);
  addField('description', description);
  addField('transmission', transmission);
  addField('transmission_model', transmissionModel);
  addField('drive', drive);
  addField('color', color);
  addField('condition', condition);
  addField('defect', defect);
  addField('supplier_code', supplierCode);
  addField('front_rear', frontRear);
  addField('left_right', leftRight);
  addField('top_bottom', topBottom);
  addField('number', number);
  addField('manufacturer', manufacturer);
  addField('manufacturer_code', manufacturerCode);
  addField('oem_code', oemCode);
  addField('wear_percentage', wearPercentage);
  addField('season', season);
  addField('diameter', diameter);
  addField('width', width);
  addField('profile', profile);
  addField('tire_quantity', tireQuantity);
  addField('drilling', drilling);
  addField('offset', offset);
  addField('center_hole_diameter', centerHoleDiameter);
  addField('tire_model', tireModel);

  if (customOperations != null) {
    customOperations.forEach((k, v) {
      if (v != null) {
        if (v is String) {
          if (v.isNotEmpty) operations[k] = v;
        } else {
          operations[k] = v;
        }
      }
    });
  }

  final scope = <String, dynamic>{
    'type': scopeType,
  };
  if (scopeType == 'BULK_SCOPE_TYPE_FILTER') {
    if (filterQuery != null) {
      final cleanFilter = Map<String, dynamic>.from(filterQuery)
        ..remove('page')
        ..remove('limit')
        ..remove('hasPhoto');
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

class _BulkEditSheet extends ConsumerStatefulWidget {
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
  ConsumerState<_BulkEditSheet> createState() => _BulkEditSheetState();
}

class _BulkEditSheetState extends ConsumerState<_BulkEditSheet> {
  final _formKey = GlobalKey<FormState>();
  final Map<String, TextEditingController> _controllers = {};
  final Set<String> _enabled = {};

  PriceMode _priceMode = PriceMode.set;
  bool _status = true;
  late String _scopeMode;
  bool _confirmEntireCatalog = false;
  bool _submitting = false;

  FieldSection _selectedSection = FieldSection.basic;
  String _searchQuery = '';
  bool _showOnlyChecked = false;
  final TextEditingController _searchCtrl = TextEditingController();

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

    for (final meta in ALL_BULK_FIELDS) {
      final ctrl = TextEditingController();
      if (meta.key == 'quantity') ctrl.text = '1';
      ctrl.addListener(() {
        if (meta.key == 'price' && mounted) setState(() {});
      });
      _controllers[meta.key] = ctrl;
    }
  }

  @override
  void dispose() {
    _searchCtrl.dispose();
    for (final ctrl in _controllers.values) {
      ctrl.dispose();
    }
    super.dispose();
  }

  void _toggle(String key, bool enabled) {
    setState(() {
      if (enabled) {
        _enabled.add(key);
      } else {
        _enabled.remove(key);
      }
    });
  }

  void _resetAll() {
    setState(() {
      _enabled.clear();
      for (final meta in ALL_BULK_FIELDS) {
        if (meta.key == 'quantity') {
          _controllers[meta.key]?.text = '1';
        } else {
          _controllers[meta.key]?.clear();
        }
      }
      _status = true;
      _priceMode = PriceMode.set;
      _showOnlyChecked = false;
    });
  }

  int _countCheckedInSection(FieldSection section) {
    return ALL_BULK_FIELDS
        .where((f) => f.section == section && _enabled.contains(f.key))
        .length;
  }

  List<BulkFieldMeta> get _visibleFields {
    final query = _searchQuery.trim().toLowerCase();
    return ALL_BULK_FIELDS.where((meta) {
      if (_showOnlyChecked && !_enabled.contains(meta.key)) return false;

      if (query.isNotEmpty) {
        return meta.label.toLowerCase().contains(query) ||
            meta.keywords.toLowerCase().contains(query) ||
            meta.key.toLowerCase().contains(query);
      }

      return meta.section == _selectedSection;
    }).toList(growable: false);
  }

  List<Map<String, dynamic>> _calculatePricePreview() {
    if (!_enabled.contains('price')) return const [];
    final text = _controllers['price']?.text.trim() ?? '';
    if (text.isEmpty) return const [];
    final val = double.tryParse(text.replaceAll(',', '.'));
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
      final raw = _controllers['price']?.text.trim().replaceAll(',', '.') ?? '';
      priceVal = double.tryParse(raw);
      if (priceVal == null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Укажите корректную цену')),
        );
        return;
      }
    }

    int? quantityVal;
    if (_enabled.contains('quantity')) {
      final raw = _controllers['quantity']?.text.trim() ?? '';
      quantityVal = int.tryParse(raw);
      if (quantityVal == null || quantityVal < 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Укажите корректное количество (>= 0)')),
        );
        return;
      }
    }

    final customOps = <String, dynamic>{};
    for (final meta in ALL_BULK_FIELDS) {
      if (!_enabled.contains(meta.key)) continue;
      if (meta.key == 'price' || meta.key == 'quantity' || meta.key == 'status') {
        continue;
      }
      final rawVal = _controllers[meta.key]?.text.trim() ?? '';

      // Нормализация селектов сторон и привода под формат API
      if (meta.key == 'front_rear') {
        if (rawVal.startsWith('F')) {
          customOps['front_rear'] = 'F';
        } else if (rawVal.startsWith('R')) {
          customOps['front_rear'] = 'R';
        } else {
          customOps['front_rear'] = rawVal;
        }
      } else if (meta.key == 'left_right') {
        if (rawVal.startsWith('L')) {
          customOps['left_right'] = 'L';
        } else if (rawVal.startsWith('R')) {
          customOps['left_right'] = 'R';
        } else {
          customOps['left_right'] = rawVal;
        }
      } else if (meta.key == 'drive') {
        if (rawVal.contains('Полный')) {
          customOps['drive'] = 'Полный';
        } else {
          customOps['drive'] = rawVal;
        }
      } else {
        customOps[meta.key] = rawVal;
      }
    }

    setState(() => _submitting = true);
    try {
      final payload = buildBulkUpdatePayload(
        ids: widget.parts.map((p) => p.id),
        scopeType: isFilterScope ? 'BULK_SCOPE_TYPE_FILTER' : 'BULK_SCOPE_TYPE_IDS',
        filterQuery: isFilterScope ? widget.activeFilter?.toQueryParameters() : null,
        confirmEntireCatalog: isFilterScope && _isFilterEmpty && _confirmEntireCatalog,
        priceMode: _enabled.contains('price') ? _priceMode : null,
        priceValue: priceVal,
        quantity: quantityVal,
        status: _enabled.contains('status') ? _status : null,
        customOperations: customOps,
      );

      final response = await apiClient.dio.post(
        '/api/v1/admin/parts/bulk-update',
        data: payload,
      );

      int updatedCount = widget.parts.length;
      if (response.data is Map && response.data['updated_count'] != null) {
        updatedCount = response.data['updated_count'] as int;
      }

      if (mounted) {
        Navigator.pop(context, true);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Успешно обновлено запчастей: $updatedCount')),
        );
      }
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

  @override
  Widget build(BuildContext context) {
    final bottomInset = MediaQuery.viewInsetsOf(context).bottom;
    final filterEmpty = _isFilterEmpty;
    final visibleFields = _visibleFields;
    final shelvesAsync = ref.watch(shelvesProvider);
    final shelves = shelvesAsync.valueOrNull ?? const <String>[];

    final combinedCategories = {
      ...widget.categories.where((c) => c.isNotEmpty),
      ...DEFAULT_CATEGORIES,
    }.toList()..sort((a, b) => a.compareTo(b));

    return Padding(
      padding: EdgeInsets.fromLTRB(16, 12, 16, bottomInset + 16),
      child: SizedBox(
        height: MediaQuery.of(context).size.height * 0.88,
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Верхний заголовок и крестик
              Row(
                children: [
                  Expanded(
                    child: Text(
                      _scopeMode == 'filter'
                          ? (widget.totalMatchingCount != null
                              ? 'Изменить все по запросу (${widget.totalMatchingCount})'
                              : 'Изменить все по запросу')
                          : 'Изменить ${widget.parts.length} запчастей',
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                  ),
                  IconButton(
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(LucideIcons.x),
                  ),
                ],
              ),
              const SizedBox(height: 6),

              // Переключатель области применения
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

              // Предупреждение при пустом фильтре
              if (_scopeMode == 'filter' && filterEmpty) ...[
                Container(
                  margin: const EdgeInsets.only(bottom: 6),
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
                const SizedBox(height: 6),
              ],

              // Строка поиска по полям
              TextField(
                controller: _searchCtrl,
                decoration: InputDecoration(
                  hintText: 'Поиск поля для изменения...',
                  prefixIcon: const Icon(LucideIcons.search, size: 18),
                  suffixIcon: _searchQuery.isNotEmpty
                      ? IconButton(
                          icon: const Icon(LucideIcons.x, size: 16),
                          onPressed: () {
                            _searchCtrl.clear();
                            setState(() => _searchQuery = '');
                          },
                        )
                      : null,
                  isDense: true,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                ),
                onChanged: (val) => setState(() => _searchQuery = val.trim().toLowerCase()),
              ),
              const SizedBox(height: 8),

              // Вкладки секций (показываются если нет поискового запроса)
              if (_searchQuery.isEmpty) ...[
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: FieldSection.values.map((section) {
                      final count = _countCheckedInSection(section);
                      final isSelected = _selectedSection == section;
                      return Padding(
                        padding: const EdgeInsets.only(right: 6),
                        child: FilterChip(
                          selected: isSelected,
                          showCheckmark: false,
                          avatar: count > 0
                              ? Container(
                                  padding: const EdgeInsets.all(4),
                                  decoration: BoxDecoration(
                                    color: isSelected
                                        ? Theme.of(context).colorScheme.onPrimary
                                        : Theme.of(context).colorScheme.primary,
                                    shape: BoxShape.circle,
                                  ),
                                  child: Text(
                                    '$count',
                                    style: TextStyle(
                                      fontSize: 10,
                                      fontWeight: FontWeight.bold,
                                      color: isSelected
                                          ? Theme.of(context).colorScheme.primary
                                          : Theme.of(context).colorScheme.onPrimary,
                                    ),
                                  ),
                                )
                              : null,
                          label: Text(section.label),
                          onSelected: (_) => setState(() => _selectedSection = section),
                        ),
                      );
                    }).toList(),
                  ),
                ),
                const SizedBox(height: 6),
              ],

              // Панель быстрого управления (только выбранные, сброс)
              Row(
                children: [
                  FilterChip(
                    label: Text('Только выбранные (${_enabled.length})'),
                    selected: _showOnlyChecked,
                    onSelected: (val) => setState(() => _showOnlyChecked = val),
                    avatar: const Icon(LucideIcons.check, size: 14),
                  ),
                  const Spacer(),
                  if (_enabled.isNotEmpty)
                    TextButton.icon(
                      style: TextButton.styleFrom(
                        padding: const EdgeInsets.symmetric(horizontal: 8),
                        visualDensity: VisualDensity.compact,
                      ),
                      onPressed: _resetAll,
                      icon: const Icon(LucideIcons.rotate_ccw, size: 14),
                      label: const Text('Сбросить', style: TextStyle(fontSize: 12)),
                    ),
                ],
              ),
              const SizedBox(height: 4),

              // Список полей
              Expanded(
                child: visibleFields.isEmpty
                    ? Center(
                        child: Text(
                          _searchQuery.isNotEmpty
                              ? 'Поля не найдены'
                              : 'Нет выбранных полей',
                          style: TextStyle(color: Theme.of(context).colorScheme.outline),
                        ),
                      )
                    : ListView.builder(
                        itemCount: visibleFields.length,
                        itemBuilder: (context, index) {
                          final meta = visibleFields[index];
                          return _buildFieldCard(
                            meta,
                            shelves: shelves,
                            categories: combinedCategories,
                          );
                        },
                      ),
              ),

              const SizedBox(height: 10),

              // Кнопка подтверждения и отправки
              SizedBox(
                width: double.infinity,
                child: FilledButton.icon(
                  onPressed: (_submitting ||
                          (_scopeMode == 'filter' && filterEmpty && !_confirmEntireCatalog) ||
                          _enabled.isEmpty)
                      ? null
                      : _submit,
                  icon: _submitting
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                        )
                      : const Icon(LucideIcons.save),
                  label: Text(
                    _submitting
                        ? 'Применение изменений...'
                        : 'Применить изменения (${_enabled.length})',
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFieldCard(
    BulkFieldMeta meta, {
    required List<String> shelves,
    required List<String> categories,
  }) {
    final isChecked = _enabled.contains(meta.key);
    final theme = Theme.of(context);

    return Container(
      margin: const EdgeInsets.only(bottom: 6),
      decoration: BoxDecoration(
        color: isChecked
            ? theme.colorScheme.primary.withValues(alpha: 0.08)
            : theme.colorScheme.surfaceContainerHighest.withValues(alpha: 0.25),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(
          color: isChecked
              ? theme.colorScheme.primary
              : theme.colorScheme.outlineVariant.withValues(alpha: 0.4),
          width: isChecked ? 1.5 : 1.0,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          CheckboxListTile(
            contentPadding: const EdgeInsets.symmetric(horizontal: 10, vertical: 0),
            dense: true,
            title: Text(
              meta.label,
              style: TextStyle(
                fontSize: 14,
                fontWeight: isChecked ? FontWeight.w600 : FontWeight.normal,
                color: isChecked ? theme.colorScheme.primary : null,
              ),
            ),
            value: isChecked,
            onChanged: (val) => _toggle(meta.key, val ?? false),
          ),
          if (isChecked)
            Padding(
              padding: const EdgeInsets.fromLTRB(14, 0, 14, 12),
              child: _buildFieldInput(
                meta,
                shelves: shelves,
                categories: categories,
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildFieldInput(
    BulkFieldMeta meta, {
    required List<String> shelves,
    required List<String> categories,
  }) {
    final ctrl = _controllers[meta.key]!;

    switch (meta.type) {
      case FieldInputType.price:
        final pricePreviews = _calculatePricePreview();
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SegmentedButton<PriceMode>(
              segments: const [
                ButtonSegment(value: PriceMode.set, label: Text('Точная')),
                ButtonSegment(value: PriceMode.percent, label: Text('%')),
                ButtonSegment(value: PriceMode.offset, label: Text('Смещение ₽')),
              ],
              selected: {_priceMode},
              onSelectionChanged: (val) => setState(() => _priceMode = val.first),
            ),
            const SizedBox(height: 8),
            TextFormField(
              controller: ctrl,
              keyboardType: const TextInputType.numberWithOptions(decimal: true, signed: true),
              inputFormatters: [FilteringTextInputFormatter.allow(RegExp(r'[-0-9.,]'))],
              decoration: InputDecoration(
                labelText: _priceMode == PriceMode.set
                    ? 'Новая цена для всех (₽) *'
                    : _priceMode == PriceMode.percent
                        ? 'Процент изменения (+/- %) *'
                        : 'Сумма изменения (+/- ₽) *',
                hintText: _priceMode == PriceMode.set
                    ? 'Например: 2500'
                    : _priceMode == PriceMode.percent
                        ? 'Например: 10 или -15'
                        : 'Например: 500 или -300',
              ),
              validator: (v) => double.tryParse((v ?? '').replaceAll(',', '.')) == null
                  ? 'Укажите числовое значение'
                  : null,
            ),
            if (pricePreviews.isNotEmpty) ...[
              const SizedBox(height: 8),
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.surfaceContainerHighest.withValues(alpha: 0.5),
                  borderRadius: BorderRadius.circular(6),
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
                          style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    ...pricePreviews.map(
                      (p) => Text(
                        '${p['name']}: ${p['old'].toStringAsFixed(0)} ₽ ➔ ${p['new'].toStringAsFixed(0)} ₽',
                        style: const TextStyle(fontSize: 11),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ],
        );

      case FieldInputType.number:
        return TextFormField(
          controller: ctrl,
          keyboardType: TextInputType.number,
          inputFormatters: [FilteringTextInputFormatter.digitsOnly],
          decoration: InputDecoration(
            labelText: '${meta.label} *',
            hintText: meta.hint,
          ),
          validator: (v) => (v == null || v.isEmpty) ? 'Заполните поле' : null,
        );

      case FieldInputType.period:
        return TextFormField(
          controller: ctrl,
          keyboardType: TextInputType.number,
          inputFormatters: const [CarReleasePeriodFormatter()],
          decoration: InputDecoration(
            labelText: meta.label,
            hintText: meta.hint ?? '2001-2007',
          ),
        );

      case FieldInputType.multiline:
        return TextFormField(
          controller: ctrl,
          maxLines: 3,
          decoration: InputDecoration(
            labelText: meta.label,
            hintText: meta.hint,
          ),
        );

      case FieldInputType.status:
        return SegmentedButton<bool>(
          segments: const [
            ButtonSegment(value: true, label: Text('Активна')),
            ButtonSegment(value: false, label: Text('Неактивна')),
          ],
          selected: {_status},
          onSelectionChanged: (val) => setState(() => _status = val.first),
        );

      case FieldInputType.shelfAutocomplete:
        return VehiclePickerField(
          controller: ctrl,
          label: meta.label,
          options: shelves,
          hint: meta.hint ?? 'Выберите или введите полку',
        );

      case FieldInputType.select:
        final options = meta.key == 'category' ? categories : meta.options;
        return DropdownButtonFormField<String>(
          initialValue: ctrl.text.isNotEmpty && options.contains(ctrl.text) ? ctrl.text : null,
          isExpanded: true,
          decoration: InputDecoration(labelText: meta.label),
          items: [
            const DropdownMenuItem<String>(
              value: '__clear__',
              child: Text('— Очистить значение —', style: TextStyle(color: Colors.grey)),
            ),
            ...options.map(
              (opt) => DropdownMenuItem(
                value: opt,
                child: Text(opt, overflow: TextOverflow.ellipsis),
              ),
            ),
          ],
          onChanged: (val) {
            if (val == '__clear__' || val == null) {
              ctrl.clear();
            } else {
              ctrl.text = val;
            }
          },
        );

      case FieldInputType.text:
        return TextFormField(
          controller: ctrl,
          decoration: InputDecoration(
            labelText: meta.label,
            hintText: meta.hint,
          ),
        );
    }
  }
}
