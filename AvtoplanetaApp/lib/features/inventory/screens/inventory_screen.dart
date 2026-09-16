import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import '../../../app/theme.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/part.dart';
import '../../../core/utils/qr_signer.dart';
import '../../../shared/widgets/app_states.dart';
import '../../auth/providers/auth_provider.dart';
import '../../orders/widgets/part_order_sheet.dart';
import '../providers/inventory_provider.dart';
import '../providers/part_catalog_provider.dart';
import '../widgets/bulk_part_actions.dart';

class InventoryScreen extends ConsumerStatefulWidget {
  const InventoryScreen({super.key});

  @override
  ConsumerState<InventoryScreen> createState() => _InventoryScreenState();
}

class _InventoryScreenState extends ConsumerState<InventoryScreen> {
  final _searchCtrl = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  Timer? _debounce;
  final Set<int> _selectedPartIds = {};
  bool _selectionMode = false;

  List<Part> _infiniteParts = [];
  int _nextInfinitePage = 2;
  bool _isLoadingMore = false;
  bool _hasMoreInfinite = true;
  InventoryFilter? _lastFilter;

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollController.removeListener(_onScroll);
    _scrollController.dispose();
    _searchCtrl.dispose();
    _debounce?.cancel();
    super.dispose();
  }

  void _onScroll() {
    if (!_scrollController.hasClients) return;
    final maxScroll = _scrollController.position.maxScrollExtent;
    final currentScroll = _scrollController.position.pixels;
    if (maxScroll - currentScroll <= 350) {
      _loadMoreInfinite();
    }
  }

  Future<void> _loadMoreInfinite() async {
    final filter = ref.read(inventoryFilterProvider);
    if (filter.pageSize != 'all' || _isLoadingMore || !_hasMoreInfinite) {
      return;
    }
    setState(() => _isLoadingMore = true);
    try {
      final params = filter.copyWith(page: _nextInfinitePage).toQueryParameters(overrideLimit: 20);
      final response = await apiClient.dio.get(
        '/api/v1/inventory',
        queryParameters: params,
      );
      final responseData = response.data;
      final list = responseData is List
          ? responseData
          : responseData is Map && responseData['parts'] is List
          ? responseData['parts'] as List
          : const [];
      final newParts = list
          .map((e) => Part.fromJson(e as Map<String, dynamic>))
          .toList();

      if (mounted) {
        setState(() {
          _infiniteParts.addAll(newParts);
          _nextInfinitePage++;
          _hasMoreInfinite = newParts.length >= 20;
          _isLoadingMore = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() => _isLoadingMore = false);
      }
    }
  }

  void _onSearch(String value) {
    setState(() {});
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 400), () {
      ref
          .read(inventoryFilterProvider.notifier)
          .update((f) => f.copyWith(search: value, page: 1));
    });
  }

  void _toggleSelection(Part part) {
    setState(() {
      _selectionMode = true;
      if (!_selectedPartIds.add(part.id)) _selectedPartIds.remove(part.id);
      if (_selectedPartIds.isEmpty) _selectionMode = false;
    });
  }

  void _clearSelection() {
    setState(() {
      _selectionMode = false;
      _selectedPartIds.clear();
    });
  }

  bool _sameStringList(List<String> a, List<String> b) {
    if (a.length != b.length) return false;
    for (var i = 0; i < a.length; i++) {
      if (a[i] != b[i]) return false;
    }
    return true;
  }

  bool _samePartContent(Part a, Part b) {
    return a.id == b.id &&
        a.name == b.name &&
        a.description == b.description &&
        a.category == b.category &&
        a.price == b.price &&
        a.quantity == b.quantity &&
        a.status == b.status &&
        a.brand == b.brand &&
        a.model == b.model &&
        a.location == b.location &&
        a.address == b.address &&
        a.salesman == b.salesman &&
        a.vin == b.vin &&
        a.carReleasePeriod == b.carReleasePeriod &&
        a.oemCode == b.oemCode &&
        a.supplierCode == b.supplierCode &&
        a.defect == b.defect &&
        a.markedForDeletion == b.markedForDeletion &&
        _sameStringList(a.photos, b.photos);
  }

  bool _samePartPrefix(List<Part> currentParts, List<Part> nextParts) {
    if (currentParts.length < nextParts.length) return false;
    for (var i = 0; i < nextParts.length; i++) {
      if (!_samePartContent(currentParts[i], nextParts[i])) return false;
    }
    return true;
  }

  @override
  Widget build(BuildContext context) {
    final filter = ref.watch(inventoryFilterProvider);
    final inventoryAsync = ref.watch(inventoryProvider(filter));
    final categories =
        ref
            .watch(partCatalogProvider)
            .valueOrNull
            ?.partFormCategories
            .map((category) => category.name)
            .toList() ??
        const <String>[];
    final user = ref.watch(authProvider).valueOrNull;

    final isOffline = inventoryAsync.valueOrNull?.isOffline ?? false;
    final visibleParts = inventoryAsync.valueOrNull?.parts ?? const <Part>[];

    if (filter.pageSize == 'all' && inventoryAsync.hasValue) {
      final data = inventoryAsync.value!;
      if (_lastFilter != filter) {
        _lastFilter = filter;
        _infiniteParts = List.from(data.parts);
        _nextInfinitePage = 2;
        _hasMoreInfinite = data.parts.length >= 20;
      } else if (!_samePartPrefix(_infiniteParts, data.parts)) {
        _infiniteParts = [
          ...data.parts,
          ..._infiniteParts.skip(data.parts.length),
        ];
        _hasMoreInfinite = data.parts.length >= 20;
      }
    }

    final displayParts = filter.pageSize == 'all'
        ? (_infiniteParts.isNotEmpty ? _infiniteParts : visibleParts)
        : visibleParts;

    final selectedParts = displayParts
        .where((part) => _selectedPartIds.contains(part.id))
        .toList();

    return Scaffold(
      appBar: AppBar(
        leading: _selectionMode
            ? IconButton(
                tooltip: 'Отменить выбор',
                onPressed: _clearSelection,
                icon: const Icon(Icons.close),
              )
            : null,
        title: Text(
          _selectionMode ? 'Выбрано: ${_selectedPartIds.length}' : 'Инвентарь',
        ),
        actions: [
          if (_selectionMode)
            IconButton(
              tooltip: _selectedPartIds.length == displayParts.length
                  ? 'Снять выбор со всех'
                  : 'Выбрать все на странице',
              icon: Icon(
                _selectedPartIds.length == displayParts.length
                    ? Icons.deselect_rounded
                    : Icons.select_all_rounded,
              ),
              onPressed: () {
                setState(() {
                  if (_selectedPartIds.length == displayParts.length) {
                    _selectedPartIds.clear();
                    _selectionMode = false;
                  } else {
                    _selectedPartIds.addAll(
                      displayParts.map((part) => part.id),
                    );
                  }
                });
              },
            ),
          if (!_selectionMode) ...[
            IconButton(
              tooltip: 'Сканировать код',
              icon: const Icon(Icons.qr_code_scanner_outlined),
              onPressed: () => _openScanner(context),
            ),
            if (user?.isOperator == true)
              IconButton(
                tooltip: 'Добавить',
                icon: const Icon(Icons.add_circle_outline_rounded),
                onPressed: isOffline
                    ? () => ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(
                          content: Text(
                            'В оффлайн-режиме добавление недоступно',
                          ),
                        ),
                      )
                    : () => _showAddMenu(context),
              ),
            IconButton(
              tooltip: 'Профиль',
              icon: const Icon(Icons.person_outline),
              onPressed: () => _showUserMenu(context),
            ),
          ],
        ],
        bottom: _selectionMode
            ? null
            : PreferredSize(
                preferredSize: const Size.fromHeight(68),
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(16, 4, 16, 12),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _searchCtrl,
                          onChanged: _onSearch,
                          decoration: InputDecoration(
                            hintText: 'Название, марка, модель или место',
                            prefixIcon: const Icon(Icons.search_rounded),
                            suffixIcon: _searchCtrl.text.isNotEmpty
                                ? IconButton(
                                    tooltip: 'Очистить поиск',
                                    icon: const Icon(Icons.close_rounded),
                                    onPressed: () {
                                      _searchCtrl.clear();
                                      _onSearch('');
                                    },
                                  )
                                : null,
                            contentPadding: const EdgeInsets.symmetric(
                              vertical: 12,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Badge(
                        isLabelVisible: filter.activeFilterCount > 0,
                        label: Text('${filter.activeFilterCount}'),
                        child: IconButton.filledTonal(
                          tooltip: 'Фильтры',
                          onPressed: () =>
                              _showFilters(context, filter, categories),
                          icon: const Icon(Icons.tune_rounded),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
      ),
      body: inventoryAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => AppEmptyState(
          icon: Icons.cloud_off_rounded,
          title: 'Не удалось загрузить склад',
          message: 'Проверьте подключение к сети и попробуйте ещё раз.',
          actionLabel: 'Повторить',
          onAction: () => ref.invalidate(inventoryProvider(filter)),
        ),
        data: (data) => Column(
          children: [
            if (data.isOffline)
              Container(
                margin: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                padding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  color: AppTheme.warningColor.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(
                    color: AppTheme.warningColor.withValues(alpha: 0.25),
                  ),
                ),
                child: const Row(
                  children: [
                    Icon(
                      Icons.cloud_off_rounded,
                      size: 18,
                      color: AppTheme.warningColor,
                    ),
                    SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        'Оффлайн-режим · доступен только просмотр',
                        style: TextStyle(
                          color: AppTheme.warningColor,
                          fontWeight: FontWeight.w600,
                          fontSize: 13,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            if (filter.activeFilterCount > 0) _activeFilters(filter),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 14, 16, 6),
              child: Row(
                children: [
                  Text(
                    'Запчасти',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const Spacer(),
                  PopupMenuButton<String>(
                    tooltip: 'Сколько показывать',
                    initialValue: filter.pageSize,
                    onSelected: (mode) {
                      ref.read(inventoryFilterProvider.notifier).update(
                            (f) => f.copyWith(pageSize: mode, page: 1),
                          );
                    },
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 5,
                      ),
                      decoration: BoxDecoration(
                        color: AppTheme.primaryColor.withValues(alpha: 0.12),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            filter.pageSize == 'all'
                                ? 'Все (${data.total})'
                                : '${filter.pageSize} / стр (${data.total})',
                            style: const TextStyle(
                              color: AppTheme.primaryColor,
                              fontSize: 12,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                          const SizedBox(width: 4),
                          const Icon(
                            Icons.arrow_drop_down,
                            size: 16,
                            color: AppTheme.primaryColor,
                          ),
                        ],
                      ),
                    ),
                    itemBuilder: (context) => const [
                      PopupMenuItem(
                        value: 'all',
                        child: Row(
                          children: [
                            Icon(
                              Icons.all_inclusive_rounded,
                              size: 18,
                              color: AppTheme.primaryColor,
                            ),
                            SizedBox(width: 8),
                            Text('Все (бесконечная лента)'),
                          ],
                        ),
                      ),
                      PopupMenuItem(
                        value: '20',
                        child: Text('20 на страницу'),
                      ),
                      PopupMenuItem(
                        value: '50',
                        child: Text('50 на страницу'),
                      ),
                      PopupMenuItem(
                        value: '100',
                        child: Text('100 на страницу'),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            Expanded(
              child: displayParts.isEmpty
                  ? const AppEmptyState(
                      icon: Icons.search_off_rounded,
                      title: 'Ничего не найдено',
                      message:
                          'Попробуйте изменить запрос или очистить строку поиска.',
                    )
                  : RefreshIndicator(
                      onRefresh: () async {
                        _lastFilter = null;
                        return ref.invalidate(inventoryProvider(filter));
                      },
                      child: ListView.builder(
                        controller: _scrollController,
                        padding: const EdgeInsets.fromLTRB(16, 6, 16, 16),
                        itemCount: displayParts.length +
                            (filter.pageSize == 'all' && _isLoadingMore ? 1 : 0),
                        itemBuilder: (ctx, i) {
                          if (i >= displayParts.length) {
                            return const Padding(
                              padding: EdgeInsets.symmetric(vertical: 16),
                              child: Center(
                                child: SizedBox(
                                  width: 24,
                                  height: 24,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2.5,
                                  ),
                                ),
                              ),
                            );
                          }
                          return Padding(
                            padding: const EdgeInsets.only(bottom: 10),
                            child: _PartCard(
                              part: displayParts[i],
                              selected: _selectedPartIds.contains(
                                displayParts[i].id,
                              ),
                              selectionMode: _selectionMode,
                              onTap: () {
                                if (_selectionMode) {
                                  _toggleSelection(displayParts[i]);
                                } else {
                                  context.go(
                                    '/inventory/part/${displayParts[i].id}',
                                  );
                                }
                              },
                              onLongPress: () =>
                                  _toggleSelection(displayParts[i]),
                            ),
                          );
                        },
                      ),
                    ),
            ),
            // Пагинация
            if (filter.pageSize != 'all' &&
                data.total > (int.tryParse(filter.pageSize) ?? 20) &&
                !_selectionMode)
              _Pagination(
                current: filter.page,
                total: (data.total / (int.tryParse(filter.pageSize) ?? 20)).ceil(),
                onPage: (p) => ref
                    .read(inventoryFilterProvider.notifier)
                    .update((f) => f.copyWith(page: p)),
              ),
          ],
        ),
      ),
      bottomNavigationBar: _selectionMode
          ? _BulkActionBar(
              count: _selectedPartIds.length,
              enabled: !isOffline && selectedParts.isNotEmpty,
              onOrder: () async {
                if (selectedParts.any((part) => part.quantity <= 0)) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text(
                        'Для заказа выберите только запчасти в наличии',
                      ),
                    ),
                  );
                  return;
                }
                final created = await showPartsOrderSheet(
                  context,
                  ref,
                  selectedParts,
                );
                if (!context.mounted) return;
                if (created) {
                  _clearSelection();
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Заказ оформлен')),
                  );
                }
              },
              onEdit: () async {
                final changed = await showBulkEditSheet(
                  context,
                  ref,
                  parts: selectedParts,
                  categories: categories,
                );
                if (!context.mounted) return;
                if (changed) {
                  _clearSelection();
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Запчасти обновлены')),
                  );
                }
              },
              onDelete: () async {
                final deleted = await confirmBulkDelete(
                  context,
                  ref,
                  selectedParts,
                );
                if (!context.mounted) return;
                if (deleted) {
                  _clearSelection();
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Запчасти удалены')),
                  );
                }
              },
            )
          : null,
    );
  }

  Future<void> _showFilters(
    BuildContext context,
    InventoryFilter current,
    List<String> categories,
  ) async {
    final result = await showModalBottomSheet<InventoryFilter>(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (_) => FractionallySizedBox(
        heightFactor: 0.92,
        child: _InventoryFilterSheet(current: current, categories: categories),
      ),
    );
    if (result != null && mounted) {
      ref.read(inventoryFilterProvider.notifier).state = result;
    }
  }

  Widget _activeFilters(InventoryFilter filter) {
    final chips = <Widget>[];

    void addChip(String label, InventoryFilter cleared) {
      chips.add(
        InputChip(
          label: Text(label),
          visualDensity: VisualDensity.compact,
          onDeleted: () {
            ref.read(inventoryFilterProvider.notifier).state = cleared;
          },
        ),
      );
    }

    if (filter.category.isNotEmpty) {
      addChip(
        'Категория: ${filter.category}',
        filter.copyWith(category: '', page: 1),
      );
    }
    if (filter.brand.isNotEmpty) {
      addChip('Бренд: ${filter.brand}', filter.copyWith(brand: '', page: 1));
    }
    if (filter.model.isNotEmpty) {
      addChip('Модель: ${filter.model}', filter.copyWith(model: '', page: 1));
    }
    if (filter.location.isNotEmpty) {
      addChip(
        'Место: ${filter.location}',
        filter.copyWith(location: '', page: 1),
      );
    }
    if (filter.salesman.isNotEmpty) {
      addChip(
        'Продавец: ${filter.salesman}',
        filter.copyWith(salesman: '', page: 1),
      );
    }
    if (filter.status.isNotEmpty) {
      addChip(
        filter.status == 'true' ? 'Доступно' : 'Недоступно',
        filter.copyWith(status: '', page: 1),
      );
    }
    if (filter.hasPhoto != 'all') {
      addChip(
        filter.hasPhoto == 'with' ? 'С фото' : 'Без фото',
        filter.copyWith(hasPhoto: 'all', page: 1),
      );
    }
    if (filter.pageSize != 'all') {
      addChip(
        'Показывать: ${filter.pageSize} / стр',
        filter.copyWith(pageSize: 'all', page: 1),
      );
    }
    if (filter.number.isNotEmpty) {
      addChip('Номер: ${filter.number}', filter.copyWith(number: '', page: 1));
    }
    if (filter.oemCode.isNotEmpty) {
      addChip('OEM: ${filter.oemCode}', filter.copyWith(oemCode: '', page: 1));
    }
    if (filter.vin.isNotEmpty) {
      addChip('VIN: ${filter.vin}', filter.copyWith(vin: '', page: 1));
    }
    if (filter.carReleasePeriod.isNotEmpty) {
      addChip('Период: ${filter.carReleasePeriod}', filter.copyWith(carReleasePeriod: '', page: 1));
    }
    if (filter.bodyBrand.isNotEmpty) {
      addChip('Кузов: ${filter.bodyBrand}', filter.copyWith(bodyBrand: '', page: 1));
    }
    if (filter.engineBrand.isNotEmpty) {
      addChip('Двигатель: ${filter.engineBrand}', filter.copyWith(engineBrand: '', page: 1));
    }
    if (filter.carReleaseDate.isNotEmpty) {
      addChip('Год: ${filter.carReleaseDate}', filter.copyWith(carReleaseDate: '', page: 1));
    }
    if (filter.transmission.isNotEmpty) {
      addChip('КПП: ${filter.transmission}', filter.copyWith(transmission: '', page: 1));
    }
    if (filter.drive.isNotEmpty) {
      addChip('Привод: ${filter.drive}', filter.copyWith(drive: '', page: 1));
    }
    if (filter.condition.isNotEmpty) {
      addChip('Состояние: ${filter.condition}', filter.copyWith(condition: '', page: 1));
    }
    if (filter.manufacturer.isNotEmpty) {
      addChip('Производитель: ${filter.manufacturer}', filter.copyWith(manufacturer: '', page: 1));
    }
    if (filter.defect.isNotEmpty) {
      addChip('Дефект: ${filter.defect}', filter.copyWith(defect: '', page: 1));
    }
    if (filter.color.isNotEmpty) {
      addChip('Цвет: ${filter.color}', filter.copyWith(color: '', page: 1));
    }
    if (filter.address.isNotEmpty) {
      addChip('Адрес: ${filter.address}', filter.copyWith(address: '', page: 1));
    }
    if (filter.minPrice.isNotEmpty) {
      addChip('Цена от: ${filter.minPrice} ₽', filter.copyWith(minPrice: '', page: 1));
    }
    if (filter.maxPrice.isNotEmpty) {
      addChip('Цена до: ${filter.maxPrice} ₽', filter.copyWith(maxPrice: '', page: 1));
    }
    if (filter.minQuantity.isNotEmpty) {
      addChip('Кол-во от: ${filter.minQuantity}', filter.copyWith(minQuantity: '', page: 1));
    }
    if (filter.maxQuantity.isNotEmpty) {
      addChip('Кол-во до: ${filter.maxQuantity}', filter.copyWith(maxQuantity: '', page: 1));
    }
    if (filter.frontRear.isNotEmpty) {
      addChip('Перед/зад: ${filter.frontRear}', filter.copyWith(frontRear: '', page: 1));
    }
    if (filter.leftRight.isNotEmpty) {
      addChip('Право/лево: ${filter.leftRight}', filter.copyWith(leftRight: '', page: 1));
    }
    if (filter.topBottom.isNotEmpty) {
      addChip('Верх/низ: ${filter.topBottom}', filter.copyWith(topBottom: '', page: 1));
    }
    if (filter.manufacturerCode.isNotEmpty) {
      addChip('Код произв.: ${filter.manufacturerCode}', filter.copyWith(manufacturerCode: '', page: 1));
    }
    if (filter.supplierCode.isNotEmpty) {
      addChip('Код поставщ.: ${filter.supplierCode}', filter.copyWith(supplierCode: '', page: 1));
    }
    if (filter.transmissionModel.isNotEmpty) {
      addChip('Модель КПП: ${filter.transmissionModel}', filter.copyWith(transmissionModel: '', page: 1));
    }
    if (filter.wearPercentage.isNotEmpty) {
      addChip('Износ: ${filter.wearPercentage}', filter.copyWith(wearPercentage: '', page: 1));
    }
    if (filter.season.isNotEmpty) {
      addChip('Сезон: ${filter.season}', filter.copyWith(season: '', page: 1));
    }
    if (filter.diameter.isNotEmpty) {
      addChip('Диаметр: ${filter.diameter}', filter.copyWith(diameter: '', page: 1));
    }
    if (filter.width.isNotEmpty) {
      addChip('Ширина: ${filter.width}', filter.copyWith(width: '', page: 1));
    }
    if (filter.profile.isNotEmpty) {
      addChip('Профиль: ${filter.profile}', filter.copyWith(profile: '', page: 1));
    }
    if (filter.tireQuantity.isNotEmpty) {
      addChip('Кол-во шин: ${filter.tireQuantity}', filter.copyWith(tireQuantity: '', page: 1));
    }
    if (filter.drilling.isNotEmpty) {
      addChip('Сверловка: ${filter.drilling}', filter.copyWith(drilling: '', page: 1));
    }
    if (filter.offset.isNotEmpty) {
      addChip('Вылет: ${filter.offset}', filter.copyWith(offset: '', page: 1));
    }
    if (filter.centerHoleDiameter.isNotEmpty) {
      addChip('Диаметр ЦО: ${filter.centerHoleDiameter}', filter.copyWith(centerHoleDiameter: '', page: 1));
    }
    if (filter.tireModel.isNotEmpty) {
      addChip('Модель шины: ${filter.tireModel}', filter.copyWith(tireModel: '', page: 1));
    }

    return SizedBox(
      height: 46,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
        itemCount: chips.length,
        separatorBuilder: (_, _) => const SizedBox(width: 6),
        itemBuilder: (_, index) => chips[index],
      ),
    );
  }

  void _showAddMenu(BuildContext context) {
    showModalBottomSheet(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(
                Icons.directions_car_outlined,
                color: AppTheme.primaryColor,
              ),
              title: const Text('Добавить одну запчасть'),
              onTap: () {
                Navigator.pop(context);
                context.go('/inventory/add');
              },
            ),
            ListTile(
              leading: const Icon(
                Icons.receipt_long_outlined,
                color: AppTheme.primaryColor,
              ),
              title: const Text('Создать дефектную ведомость'),
              onTap: () {
                Navigator.pop(context);
                context.go('/inventory/defect-report');
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showUserMenu(BuildContext context) {
    final user = ref.read(authProvider).valueOrNull;
    showModalBottomSheet(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (user?.isAdmin == true)
              ListTile(
                leading: const Icon(Icons.admin_panel_settings_outlined),
                title: const Text('Администрирование'),
                onTap: () {
                  Navigator.pop(context);
                  context.push('/admin');
                },
              ),
            ListTile(
              leading: const Icon(Icons.logout),
              title: const Text('Выйти'),
              onTap: () {
                Navigator.pop(context);
                ref.read(authProvider.notifier).logout();
              },
            ),
          ],
        ),
      ),
    );
  }

  void _openScanner(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.black,
      builder: (ctx) => FractionallySizedBox(
        heightFactor: 0.85,
        child: _ScannerModal(
          onScan: (code) {
            Navigator.pop(ctx);
            _handleScanResult(code);
          },
        ),
      ),
    );
  }

  void _handleScanResult(String code) {
    if (code.startsWith('ap:')) {
      final parts = code.split(':');
      if (parts.length >= 4) {
        final idStr = parts[1];
        final tsStr = parts[2];
        final hmacStr = parts[3];

        final id = int.tryParse(idStr);
        final ts = int.tryParse(tsStr);

        if (id != null && ts != null) {
          final isValid = QrSigner.verify(id, ts, hmacStr);
          if (isValid) {
            context.go('/inventory/part/$id');
            return;
          } else {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Неверная подпись QR-кода запчасти'),
                backgroundColor: Colors.red,
              ),
            );
            return;
          }
        }
      }
    }
    _searchCtrl.text = code;
    _onSearch(code);
  }
}

const _defaultPartCategories = [
  'Выхлопная система',
  'Двигатель',
  'Диски и шины',
  'Кузов',
  'Кузов внутри',
  'Кузов снаружи',
  'Оптика',
  'Пневмосистема',
  'Подвеска',
  'Подвеска ДВС/КПП',
  'Подвеска задних колес',
  'Подвеска передних колес',
  'Рулевое управление',
  'Система кондиционирования',
  'Система охлаждения и отопления',
  'Система выхлопа (Глушитель)',
  'Система рулевого управления',
  'Система фильтрации (Фильтры)',
  'Сопутствующие товары',
  'Стекла',
  'Тормоза',
  'Тормозная система',
  'Трансмиссия',
  'Шины и диски',
  'Электрика',
  'Электрооснащение',
  'Автохимия и масла',
  'Аксессуары и тюннинг',
  'Интерьер',
  'Другое',
];

class _InventoryFilterSheet extends StatefulWidget {
  final InventoryFilter current;
  final List<String> categories;

  const _InventoryFilterSheet({
    required this.current,
    required this.categories,
  });

  @override
  State<_InventoryFilterSheet> createState() => _InventoryFilterSheetState();
}

class _InventoryFilterSheetState extends State<_InventoryFilterSheet> {
  late final TextEditingController _brandController;
  late final TextEditingController _modelController;
  late final TextEditingController _locationController;
  late final TextEditingController _addressController;
  late final TextEditingController _salesmanController;
  late final TextEditingController _numberController;
  late final TextEditingController _oemCodeController;
  late final TextEditingController _vinController;
  late final TextEditingController _carReleasePeriodController;
  late final TextEditingController _bodyBrandController;
  late final TextEditingController _engineBrandController;
  late final TextEditingController _carReleaseDateController;
  late final TextEditingController _conditionController;
  late final TextEditingController _manufacturerController;
  late final TextEditingController _defectController;
  late final TextEditingController _colorController;
  late final TextEditingController _minPriceController;
  late final TextEditingController _maxPriceController;
  late final TextEditingController _minQuantityController;
  late final TextEditingController _maxQuantityController;
  late final TextEditingController _manufacturerCodeController;
  late final TextEditingController _supplierCodeController;
  late final TextEditingController _transmissionModelController;
  late final TextEditingController _wearPercentageController;
  late final TextEditingController _diameterController;
  late final TextEditingController _widthController;
  late final TextEditingController _profileController;
  late final TextEditingController _tireQuantityController;
  late final TextEditingController _drillingController;
  late final TextEditingController _offsetController;
  late final TextEditingController _centerHoleDiameterController;
  late final TextEditingController _tireModelController;
  late String _category;
  late String _transmission;
  late String _drive;
  late String _frontRear;
  late String _leftRight;
  late String _topBottom;
  late String _season;
  late String _status;
  late String _hasPhoto;
  late String _pageSize;

  @override
  void initState() {
    super.initState();
    _brandController = TextEditingController(text: widget.current.brand);
    _modelController = TextEditingController(text: widget.current.model);
    _locationController = TextEditingController(text: widget.current.location);
    _addressController = TextEditingController(text: widget.current.address);
    _salesmanController = TextEditingController(text: widget.current.salesman);
    _numberController = TextEditingController(text: widget.current.number);
    _oemCodeController = TextEditingController(text: widget.current.oemCode);
    _vinController = TextEditingController(text: widget.current.vin);
    _carReleasePeriodController = TextEditingController(text: widget.current.carReleasePeriod);
    _bodyBrandController = TextEditingController(text: widget.current.bodyBrand);
    _engineBrandController = TextEditingController(text: widget.current.engineBrand);
    _carReleaseDateController = TextEditingController(text: widget.current.carReleaseDate);
    _conditionController = TextEditingController(text: widget.current.condition);
    _manufacturerController = TextEditingController(text: widget.current.manufacturer);
    _defectController = TextEditingController(text: widget.current.defect);
    _colorController = TextEditingController(text: widget.current.color);
    _minPriceController = TextEditingController(text: widget.current.minPrice);
    _maxPriceController = TextEditingController(text: widget.current.maxPrice);
    _minQuantityController = TextEditingController(text: widget.current.minQuantity);
    _maxQuantityController = TextEditingController(text: widget.current.maxQuantity);
    _manufacturerCodeController = TextEditingController(text: widget.current.manufacturerCode);
    _supplierCodeController = TextEditingController(text: widget.current.supplierCode);
    _transmissionModelController = TextEditingController(text: widget.current.transmissionModel);
    _wearPercentageController = TextEditingController(text: widget.current.wearPercentage);
    _diameterController = TextEditingController(text: widget.current.diameter);
    _widthController = TextEditingController(text: widget.current.width);
    _profileController = TextEditingController(text: widget.current.profile);
    _tireQuantityController = TextEditingController(text: widget.current.tireQuantity);
    _drillingController = TextEditingController(text: widget.current.drilling);
    _offsetController = TextEditingController(text: widget.current.offset);
    _centerHoleDiameterController = TextEditingController(text: widget.current.centerHoleDiameter);
    _tireModelController = TextEditingController(text: widget.current.tireModel);
    _category = widget.current.category;
    _transmission = widget.current.transmission;
    _drive = widget.current.drive;
    _frontRear = widget.current.frontRear;
    _leftRight = widget.current.leftRight;
    _topBottom = widget.current.topBottom;
    _season = widget.current.season;
    _status = widget.current.status;
    _hasPhoto = widget.current.hasPhoto;
    _pageSize = widget.current.pageSize;
  }

  @override
  void dispose() {
    _brandController.dispose();
    _modelController.dispose();
    _locationController.dispose();
    _addressController.dispose();
    _salesmanController.dispose();
    _numberController.dispose();
    _oemCodeController.dispose();
    _vinController.dispose();
    _carReleasePeriodController.dispose();
    _bodyBrandController.dispose();
    _engineBrandController.dispose();
    _carReleaseDateController.dispose();
    _conditionController.dispose();
    _manufacturerController.dispose();
    _defectController.dispose();
    _colorController.dispose();
    _minPriceController.dispose();
    _maxPriceController.dispose();
    _minQuantityController.dispose();
    _maxQuantityController.dispose();
    _manufacturerCodeController.dispose();
    _supplierCodeController.dispose();
    _transmissionModelController.dispose();
    _wearPercentageController.dispose();
    _diameterController.dispose();
    _widthController.dispose();
    _profileController.dispose();
    _tireQuantityController.dispose();
    _drillingController.dispose();
    _offsetController.dispose();
    _centerHoleDiameterController.dispose();
    _tireModelController.dispose();
    super.dispose();
  }

  void _reset() {
    setState(() {
      _brandController.clear();
      _modelController.clear();
      _locationController.clear();
      _addressController.clear();
      _salesmanController.clear();
      _numberController.clear();
      _oemCodeController.clear();
      _vinController.clear();
      _carReleasePeriodController.clear();
      _bodyBrandController.clear();
      _engineBrandController.clear();
      _carReleaseDateController.clear();
      _conditionController.clear();
      _manufacturerController.clear();
      _defectController.clear();
      _colorController.clear();
      _minPriceController.clear();
      _maxPriceController.clear();
      _minQuantityController.clear();
      _maxQuantityController.clear();
      _manufacturerCodeController.clear();
      _supplierCodeController.clear();
      _transmissionModelController.clear();
      _wearPercentageController.clear();
      _diameterController.clear();
      _widthController.clear();
      _profileController.clear();
      _tireQuantityController.clear();
      _drillingController.clear();
      _offsetController.clear();
      _centerHoleDiameterController.clear();
      _tireModelController.clear();
      _category = '';
      _transmission = '';
      _drive = '';
      _frontRear = '';
      _leftRight = '';
      _topBottom = '';
      _season = '';
      _status = '';
      _hasPhoto = 'all';
      _pageSize = 'all';
    });
  }

  void _apply() {
    Navigator.pop(
      context,
      widget.current.copyWith(
        category: _category,
        brand: _brandController.text.trim(),
        model: _modelController.text.trim(),
        location: _locationController.text.trim(),
        address: _addressController.text.trim(),
        salesman: _salesmanController.text.trim(),
        number: _numberController.text.trim(),
        oemCode: _oemCodeController.text.trim(),
        vin: _vinController.text.trim(),
        carReleasePeriod: _carReleasePeriodController.text.trim(),
        bodyBrand: _bodyBrandController.text.trim(),
        engineBrand: _engineBrandController.text.trim(),
        carReleaseDate: _carReleaseDateController.text.trim(),
        transmission: _transmission,
        drive: _drive,
        condition: _conditionController.text.trim(),
        manufacturer: _manufacturerController.text.trim(),
        defect: _defectController.text.trim(),
        color: _colorController.text.trim(),
        minPrice: _minPriceController.text.trim(),
        maxPrice: _maxPriceController.text.trim(),
        minQuantity: _minQuantityController.text.trim(),
        maxQuantity: _maxQuantityController.text.trim(),
        frontRear: _frontRear,
        leftRight: _leftRight,
        topBottom: _topBottom,
        manufacturerCode: _manufacturerCodeController.text.trim(),
        supplierCode: _supplierCodeController.text.trim(),
        transmissionModel: _transmissionModelController.text.trim(),
        wearPercentage: _wearPercentageController.text.trim(),
        season: _season,
        diameter: _diameterController.text.trim(),
        width: _widthController.text.trim(),
        profile: _profileController.text.trim(),
        tireQuantity: _tireQuantityController.text.trim(),
        drilling: _drillingController.text.trim(),
        offset: _offsetController.text.trim(),
        centerHoleDiameter: _centerHoleDiameterController.text.trim(),
        tireModel: _tireModelController.text.trim(),
        status: _status,
        hasPhoto: _hasPhoto,
        pageSize: _pageSize,
        page: 1,
      ),
    );
  }

  int get _mainFiltersCount {
    var count = 0;
    if (_brandController.text.trim().isNotEmpty) count++;
    if (_modelController.text.trim().isNotEmpty) count++;
    if (_category.isNotEmpty) count++;
    if (_carReleaseDateController.text.trim().isNotEmpty) count++;
    if (_conditionController.text.trim().isNotEmpty) count++;
    if (_minPriceController.text.trim().isNotEmpty || _maxPriceController.text.trim().isNotEmpty) count++;
    if (_minQuantityController.text.trim().isNotEmpty || _maxQuantityController.text.trim().isNotEmpty) count++;
    if (_status.isNotEmpty) count++;
    if (_hasPhoto != 'all') count++;
    if (_pageSize != 'all') count++;
    return count;
  }

  int get _bodyEngineFiltersCount {
    var count = 0;
    if (_bodyBrandController.text.trim().isNotEmpty) count++;
    if (_engineBrandController.text.trim().isNotEmpty) count++;
    if (_vinController.text.trim().isNotEmpty) count++;
    if (_carReleasePeriodController.text.trim().isNotEmpty) count++;
    if (_numberController.text.trim().isNotEmpty) count++;
    if (_oemCodeController.text.trim().isNotEmpty) count++;
    if (_defectController.text.trim().isNotEmpty) count++;
    if (_colorController.text.trim().isNotEmpty) count++;
    return count;
  }

  int get _transmissionFiltersCount {
    var count = 0;
    if (_transmission.isNotEmpty) count++;
    if (_transmissionModelController.text.trim().isNotEmpty) count++;
    if (_drive.isNotEmpty) count++;
    if (_frontRear.isNotEmpty) count++;
    if (_leftRight.isNotEmpty) count++;
    if (_topBottom.isNotEmpty) count++;
    return count;
  }

  int get _warehouseFiltersCount {
    var count = 0;
    if (_locationController.text.trim().isNotEmpty) count++;
    if (_addressController.text.trim().isNotEmpty) count++;
    if (_salesmanController.text.trim().isNotEmpty) count++;
    if (_manufacturerController.text.trim().isNotEmpty) count++;
    if (_manufacturerCodeController.text.trim().isNotEmpty) count++;
    if (_supplierCodeController.text.trim().isNotEmpty) count++;
    if (_wearPercentageController.text.trim().isNotEmpty) count++;
    return count;
  }

  int get _wheelsFiltersCount {
    var count = 0;
    if (_season.isNotEmpty) count++;
    if (_diameterController.text.trim().isNotEmpty) count++;
    if (_widthController.text.trim().isNotEmpty) count++;
    if (_profileController.text.trim().isNotEmpty) count++;
    if (_tireQuantityController.text.trim().isNotEmpty) count++;
    if (_drillingController.text.trim().isNotEmpty) count++;
    if (_offsetController.text.trim().isNotEmpty) count++;
    if (_centerHoleDiameterController.text.trim().isNotEmpty) count++;
    if (_tireModelController.text.trim().isNotEmpty) count++;
    return count;
  }

  int get _totalFiltersCount =>
      _mainFiltersCount +
      _bodyEngineFiltersCount +
      _transmissionFiltersCount +
      _warehouseFiltersCount +
      _wheelsFiltersCount;

  Widget _buildTab(String label, IconData icon, int count) {
    return Tab(
      height: 44,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16),
          const SizedBox(width: 6),
          Text(label),
          if (count > 0) ...[
            const SizedBox(width: 6),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: AppTheme.primaryColor,
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                '$count',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  height: 1.1,
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final categories = {
      ..._defaultPartCategories,
      ...widget.categories.where((category) => category.isNotEmpty),
      if (_category.isNotEmpty) _category,
    }.toList()..sort();
    final keyboardInset = MediaQuery.viewInsetsOf(context).bottom;

    return DefaultTabController(
      length: 5,
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 8, 8, 4),
            child: Row(
              children: [
                Expanded(
                  child: Row(
                    children: [
                      Text(
                        'Фильтры запчастей',
                        style: Theme.of(context).textTheme.titleLarge,
                      ),
                      if (_totalFiltersCount > 0) ...[
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                          decoration: BoxDecoration(
                            color: AppTheme.primaryColor.withValues(alpha: 0.18),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: AppTheme.primaryColor.withValues(alpha: 0.5),
                              width: 1,
                            ),
                          ),
                          child: Text(
                            '$_totalFiltersCount',
                            style: const TextStyle(
                              color: AppTheme.primaryColor,
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                IconButton(
                  tooltip: 'Закрыть',
                  onPressed: () => Navigator.pop(context),
                  icon: const Icon(Icons.close),
                ),
              ],
            ),
          ),
          TabBar(
            isScrollable: true,
            tabAlignment: TabAlignment.start,
            dividerColor: Theme.of(context).dividerColor.withValues(alpha: 0.25),
            tabs: [
              _buildTab('Основные', Icons.tune_rounded, _mainFiltersCount),
              _buildTab('Кузов и ДВС', Icons.handyman_rounded, _bodyEngineFiltersCount),
              _buildTab('КПП и привод', Icons.alt_route_rounded, _transmissionFiltersCount),
              _buildTab('Склад', Icons.warehouse_rounded, _warehouseFiltersCount),
              _buildTab('Шины и диски', Icons.album_rounded, _wheelsFiltersCount),
            ],
          ),
          Expanded(
            child: TabBarView(
              children: [
                // 1. Основные
                ListView(
                  padding: EdgeInsets.fromLTRB(20, 16, 20, keyboardInset + 20),
                  children: [
                    DropdownButtonFormField<String>(
                      key: ValueKey('category-$_category'),
                      initialValue: _category,
                      isExpanded: true,
                      decoration: const InputDecoration(labelText: 'Категория'),
                      items: [
                        const DropdownMenuItem(
                          value: '',
                          child: Text('Все категории'),
                        ),
                        ...categories.map(
                          (category) => DropdownMenuItem(
                            value: category,
                            child: Text(category, overflow: TextOverflow.ellipsis),
                          ),
                        ),
                      ],
                      onChanged: (value) => setState(() => _category = value ?? ''),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _brandController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Бренд',
                        hintText: 'BMW, Toyota…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _modelController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Модель',
                        hintText: 'E90, A4…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _carReleaseDateController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Год выпуска',
                        hintText: '2015…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _conditionController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Состояние',
                        hintText: 'Контрактная, б/у, новая…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: TextField(
                            controller: _minPriceController,
                            keyboardType: TextInputType.number,
                            onChanged: (_) => setState(() {}),
                            decoration: const InputDecoration(
                              labelText: 'Цена от (₽)',
                              hintText: '0',
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: TextField(
                            controller: _maxPriceController,
                            keyboardType: TextInputType.number,
                            onChanged: (_) => setState(() {}),
                            decoration: const InputDecoration(
                              labelText: 'Цена до (₽)',
                              hintText: '100000',
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: TextField(
                            controller: _minQuantityController,
                            keyboardType: TextInputType.number,
                            onChanged: (_) => setState(() {}),
                            decoration: const InputDecoration(
                              labelText: 'Кол-во от',
                              hintText: '1',
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: TextField(
                            controller: _maxQuantityController,
                            keyboardType: TextInputType.number,
                            onChanged: (_) => setState(() {}),
                            decoration: const InputDecoration(
                              labelText: 'Кол-во до',
                              hintText: '100',
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      key: ValueKey('status-$_status'),
                      initialValue: _status,
                      decoration: const InputDecoration(labelText: 'Статус'),
                      items: const [
                        DropdownMenuItem(value: '', child: Text('Все статусы')),
                        DropdownMenuItem(value: 'true', child: Text('Доступно')),
                        DropdownMenuItem(value: 'false', child: Text('Недоступно')),
                      ],
                      onChanged: (value) => setState(() => _status = value ?? ''),
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      key: ValueKey('photo-$_hasPhoto'),
                      initialValue: _hasPhoto,
                      decoration: const InputDecoration(labelText: 'Фото'),
                      items: const [
                        DropdownMenuItem(value: 'all', child: Text('Все')),
                        DropdownMenuItem(value: 'with', child: Text('С фото')),
                        DropdownMenuItem(value: 'without', child: Text('Без фото')),
                      ],
                      onChanged: (value) {
                        setState(() => _hasPhoto = value ?? 'all');
                      },
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      key: ValueKey('pageSize-$_pageSize'),
                      initialValue: _pageSize,
                      decoration: const InputDecoration(
                        labelText: 'Отображение запчастей',
                      ),
                      items: const [
                        DropdownMenuItem(
                          value: 'all',
                          child: Text('Все (бесконечная лента)'),
                        ),
                        DropdownMenuItem(
                          value: '20',
                          child: Text('20 на страницу'),
                        ),
                        DropdownMenuItem(
                          value: '50',
                          child: Text('50 на страницу'),
                        ),
                        DropdownMenuItem(
                          value: '100',
                          child: Text('100 на страницу'),
                        ),
                      ],
                      onChanged: (value) {
                        setState(() => _pageSize = value ?? 'all');
                      },
                    ),
                  ],
                ),

                // 2. Кузов и ДВС
                ListView(
                  padding: EdgeInsets.fromLTRB(20, 16, 20, keyboardInset + 20),
                  children: [
                    TextField(
                      controller: _bodyBrandController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Марка кузова',
                        hintText: 'E90, W204, G05…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _engineBrandController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Марка двигателя',
                        hintText: 'N52B30, 2JZ…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _vinController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'VIN / Номер кузова',
                        hintText: 'WBA…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _carReleasePeriodController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Период выпуска автомобиля',
                        hintText: '2001-2007…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _numberController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Номер детали',
                        hintText: '51117188830…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _oemCodeController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'OEM код',
                        hintText: '17117559273…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _defectController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Дефект',
                        hintText: 'Царапины, трещина…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _colorController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Цвет',
                        hintText: 'Черный, серебристый…',
                      ),
                    ),
                  ],
                ),

                // 3. КПП и привод
                ListView(
                  padding: EdgeInsets.fromLTRB(20, 16, 20, keyboardInset + 20),
                  children: [
                    DropdownButtonFormField<String>(
                      key: ValueKey('transmission-$_transmission'),
                      initialValue: _transmission,
                      decoration: const InputDecoration(labelText: 'Трансмиссия'),
                      items: const [
                        DropdownMenuItem(value: '', child: Text('Все типы трансмиссии')),
                        DropdownMenuItem(value: 'АКПП', child: Text('АКПП (Автомат)')),
                        DropdownMenuItem(value: 'МКПП', child: Text('МКПП (Механика)')),
                        DropdownMenuItem(value: 'Вариатор', child: Text('Вариатор (CVT)')),
                        DropdownMenuItem(value: 'Робот', child: Text('Робот')),
                      ],
                      onChanged: (value) => setState(() => _transmission = value ?? ''),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _transmissionModelController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Модель КПП',
                        hintText: 'U140F, RE4F04B…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      key: ValueKey('drive-$_drive'),
                      initialValue: _drive,
                      decoration: const InputDecoration(labelText: 'Привод'),
                      items: const [
                        DropdownMenuItem(value: '', child: Text('Все приводы')),
                        DropdownMenuItem(value: 'Передний', child: Text('Передний привод')),
                        DropdownMenuItem(value: 'Задний', child: Text('Задний привод')),
                        DropdownMenuItem(value: 'Полный', child: Text('Полный привод')),
                      ],
                      onChanged: (value) => setState(() => _drive = value ?? ''),
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      key: ValueKey('frontRear-$_frontRear'),
                      initialValue: _frontRear,
                      decoration: const InputDecoration(labelText: 'Перед / зад'),
                      items: const [
                        DropdownMenuItem(value: '', child: Text('Все расположения')),
                        DropdownMenuItem(value: 'перед', child: Text('Перед')),
                        DropdownMenuItem(value: 'зад', child: Text('Зад')),
                        DropdownMenuItem(value: 'перед / зад', child: Text('Перед / зад')),
                      ],
                      onChanged: (value) => setState(() => _frontRear = value ?? ''),
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      key: ValueKey('leftRight-$_leftRight'),
                      initialValue: _leftRight,
                      decoration: const InputDecoration(labelText: 'Право / лево'),
                      items: const [
                        DropdownMenuItem(value: '', child: Text('Все стороны')),
                        DropdownMenuItem(value: 'лево', child: Text('Лево')),
                        DropdownMenuItem(value: 'право', child: Text('Право')),
                        DropdownMenuItem(value: 'лево / право', child: Text('Лево / право')),
                      ],
                      onChanged: (value) => setState(() => _leftRight = value ?? ''),
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      key: ValueKey('topBottom-$_topBottom'),
                      initialValue: _topBottom,
                      decoration: const InputDecoration(labelText: 'Верх / низ'),
                      items: const [
                        DropdownMenuItem(value: '', child: Text('Все положения')),
                        DropdownMenuItem(value: 'верх', child: Text('Верх')),
                        DropdownMenuItem(value: 'низ', child: Text('Низ')),
                        DropdownMenuItem(value: 'верх / низ', child: Text('Верх / низ')),
                      ],
                      onChanged: (value) => setState(() => _topBottom = value ?? ''),
                    ),
                  ],
                ),

                // 4. Склад
                ListView(
                  padding: EdgeInsets.fromLTRB(20, 16, 20, keyboardInset + 20),
                  children: [
                    TextField(
                      controller: _locationController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Местоположение',
                        hintText: 'Стеллаж A-12…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _addressController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Адрес склада',
                        hintText: 'Профсоюзная 2/11…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _salesmanController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(labelText: 'Продавец'),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _manufacturerController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Производитель',
                        hintText: 'Bosch, Denso, Lemforder…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _manufacturerCodeController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Код производителя',
                        hintText: 'Код производителя…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _supplierCodeController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Код поставщика',
                        hintText: 'Код поставщика…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _wearPercentageController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Процент износа',
                        hintText: '5%, 10%…',
                      ),
                    ),
                  ],
                ),

                // 5. Шины и диски
                ListView(
                  padding: EdgeInsets.fromLTRB(20, 16, 20, keyboardInset + 20),
                  children: [
                    DropdownButtonFormField<String>(
                      key: ValueKey('season-$_season'),
                      initialValue: _season,
                      decoration: const InputDecoration(labelText: 'Сезонность шин'),
                      items: const [
                        DropdownMenuItem(value: '', child: Text('Все сезоны')),
                        DropdownMenuItem(value: 'лето', child: Text('Лето')),
                        DropdownMenuItem(value: 'зима', child: Text('Зима')),
                        DropdownMenuItem(value: 'всесезонные', child: Text('Всесезонные')),
                        DropdownMenuItem(value: 'зима шипованные', child: Text('Зима шипованные')),
                        DropdownMenuItem(value: 'зима нешипованные (липучка)', child: Text('Зима нешипованные (липучка)')),
                      ],
                      onChanged: (value) => setState(() => _season = value ?? ''),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _diameterController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Диаметр',
                        hintText: 'R15, R16, R17…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _widthController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Ширина шины',
                        hintText: '205, 215, 225…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _profileController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Профиль шины',
                        hintText: '55, 60, 65…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _tireQuantityController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Количество шин',
                        hintText: '4 шт, пара…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _drillingController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Сверловка (PCD)',
                        hintText: '5x114.3, 4x100…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _offsetController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Вылет (ET)',
                        hintText: 'ET45, ET38…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _centerHoleDiameterController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Диаметр ЦО (DIA)',
                        hintText: '60.1, 67.1…',
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _tireModelController,
                      onChanged: (_) => setState(() {}),
                      decoration: const InputDecoration(
                        labelText: 'Модель шины',
                        hintText: 'Hakkapeliitta 8, Ice Cruiser…',
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          Container(
            padding: EdgeInsets.fromLTRB(16, 12, 16, 12 + MediaQuery.paddingOf(context).bottom),
            decoration: BoxDecoration(
              color: Theme.of(context).cardColor,
              border: Border(
                top: BorderSide(
                  color: Theme.of(context).dividerColor.withValues(alpha: 0.25),
                ),
              ),
            ),
            child: Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _reset,
                    icon: const Icon(Icons.clear_all_rounded),
                    label: const Text('Сбросить'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: FilledButton.icon(
                    onPressed: _apply,
                    icon: const Icon(Icons.check_rounded),
                    label: Text(
                      _totalFiltersCount > 0
                          ? 'Применить ($_totalFiltersCount)'
                          : 'Применить',
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _PartCard extends StatelessWidget {
  final Part part;
  final bool selected;
  final bool selectionMode;
  final VoidCallback onTap;
  final VoidCallback onLongPress;

  const _PartCard({
    required this.part,
    required this.selected,
    required this.selectionMode,
    required this.onTap,
    required this.onLongPress,
  });

  @override
  Widget build(BuildContext context) {
    final photoUrl = part.photos.isNotEmpty
        ? apiClient.resolveUrl(part.photos.first)
        : null;

    return Card(
      color: selected ? AppTheme.primaryColor.withValues(alpha: 0.18) : null,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(20),
        side: selected
            ? const BorderSide(color: AppTheme.primaryColor, width: 2)
            : BorderSide.none,
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(20),
        onTap: onTap,
        onLongPress: onLongPress,
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Row(
            children: [
              // Фото
              ClipRRect(
                borderRadius: BorderRadius.circular(14),
                child: photoUrl != null
                    ? CachedNetworkImage(
                        imageUrl: photoUrl,
                        width: 76,
                        height: 76,
                        fit: BoxFit.cover,
                        placeholder: (ctx, url) => Container(
                          width: 76,
                          height: 76,
                          color: AppTheme.surfaceColor,
                          child: const Icon(
                            Icons.image_outlined,
                            color: Colors.white24,
                          ),
                        ),
                        errorWidget: (ctx, url, err) => _placeholder(),
                      )
                    : _placeholder(),
              ),
              const SizedBox(width: 14),
              // Инфо
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      part.name,
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.w700,
                        color: Colors.white,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),
                    if (part.brand != null || part.model != null)
                      Text(
                        [
                          part.brand,
                          part.model,
                        ].where((e) => e != null && e.isNotEmpty).join(' • '),
                        style: const TextStyle(
                          color: AppTheme.mutedColor,
                          fontSize: 12,
                        ),
                      ),
                    const SizedBox(height: 4),
                    Wrap(
                      spacing: 6,
                      runSpacing: 5,
                      children: [
                        _chip(part.category, AppTheme.primaryColor),
                        if (part.location.isNotEmpty)
                          _chip(part.location, AppTheme.successColor),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              if (selectionMode) ...[
                Checkbox(value: selected, onChanged: (_) => onTap()),
                const SizedBox(width: 4),
              ],
              // Цена и количество
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text(
                    '${part.price.toStringAsFixed(0)} ₽',
                    style: const TextStyle(
                      color: AppTheme.primaryColor,
                      fontWeight: FontWeight.bold,
                      fontSize: 15,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '${part.quantity} шт.',
                    style: TextStyle(
                      color: part.quantity > 0
                          ? AppTheme.mutedColor
                          : AppTheme.dangerColor,
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _placeholder() => Container(
    width: 76,
    height: 76,
    color: AppTheme.surfaceColor,
    child: const Icon(Icons.directions_car_outlined, color: Colors.white24),
  );

  Widget _chip(String label, Color color) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
    decoration: BoxDecoration(
      color: color.withValues(alpha: 0.15),
      borderRadius: BorderRadius.circular(20),
      border: Border.all(color: color.withValues(alpha: 0.4), width: 0.5),
    ),
    child: Text(label, style: TextStyle(color: color, fontSize: 11)),
  );
}

class _BulkActionBar extends StatelessWidget {
  final int count;
  final bool enabled;
  final VoidCallback onOrder;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const _BulkActionBar({
    required this.count,
    required this.enabled,
    required this.onOrder,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) => SafeArea(
    child: Material(
      elevation: 12,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(8, 8, 8, 10),
        child: Row(
          children: [
            Expanded(
              child: _action(
                Icons.shopping_cart_checkout_rounded,
                'Заказать ($count)',
                enabled ? onOrder : null,
              ),
            ),
            Expanded(
              child: _action(
                Icons.edit_outlined,
                'Изменить',
                enabled ? onEdit : null,
              ),
            ),
            Expanded(
              child: _action(
                Icons.delete_outline_rounded,
                'Удалить',
                enabled ? onDelete : null,
                danger: true,
              ),
            ),
          ],
        ),
      ),
    ),
  );

  Widget _action(
    IconData icon,
    String label,
    VoidCallback? onPressed, {
    bool danger = false,
  }) => TextButton.icon(
    onPressed: onPressed,
    style: danger
        ? TextButton.styleFrom(foregroundColor: AppTheme.dangerColor)
        : null,
    icon: Icon(icon),
    label: Text(label),
  );
}

class _Pagination extends StatelessWidget {
  final int current;
  final int total;
  final void Function(int) onPage;
  const _Pagination({
    required this.current,
    required this.total,
    required this.onPage,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          IconButton(
            icon: const Icon(Icons.chevron_left),
            onPressed: current > 1 ? () => onPage(current - 1) : null,
          ),
          Text(
            '$current / $total',
            style: const TextStyle(color: Colors.white70),
          ),
          IconButton(
            icon: const Icon(Icons.chevron_right),
            onPressed: current < total ? () => onPage(current + 1) : null,
          ),
        ],
      ),
    );
  }
}

class _ScannerModal extends StatefulWidget {
  final void Function(String) onScan;
  const _ScannerModal({required this.onScan});

  @override
  State<_ScannerModal> createState() => _ScannerModalState();
}

class _ScannerModalState extends State<_ScannerModal> {
  final MobileScannerController _controller = MobileScannerController();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        title: const Text(
          'Сканирование кода',
          style: TextStyle(color: Colors.white),
        ),
        leading: IconButton(
          icon: const Icon(Icons.close, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
        actions: [
          IconButton(
            icon: ValueListenableBuilder(
              valueListenable: _controller,
              builder: (context, state, child) {
                switch (state.torchState) {
                  case TorchState.off:
                    return const Icon(Icons.flash_off, color: Colors.white54);
                  case TorchState.on:
                    return const Icon(Icons.flash_on, color: Colors.amber);
                  default:
                    return const Icon(Icons.flash_off, color: Colors.white54);
                }
              },
            ),
            onPressed: () => _controller.toggleTorch(),
          ),
          IconButton(
            icon: ValueListenableBuilder(
              valueListenable: _controller,
              builder: (context, state, child) {
                switch (state.cameraDirection) {
                  case CameraFacing.front:
                    return const Icon(Icons.camera_front, color: Colors.white);
                  case CameraFacing.back:
                    return const Icon(Icons.camera_rear, color: Colors.white);
                  default:
                    return const Icon(Icons.camera_rear, color: Colors.white);
                }
              },
            ),
            onPressed: () => _controller.switchCamera(),
          ),
        ],
      ),
      body: Stack(
        children: [
          MobileScanner(
            controller: _controller,
            onDetect: (capture) {
              final List<Barcode> barcodes = capture.barcodes;
              for (final barcode in barcodes) {
                final String? rawValue = barcode.rawValue;
                if (rawValue != null && rawValue.isNotEmpty) {
                  widget.onScan(rawValue);
                  break;
                }
              }
            },
          ),
          Center(
            child: Container(
              width: 250,
              height: 250,
              decoration: BoxDecoration(
                border: Border.all(color: AppTheme.primaryColor, width: 3),
                borderRadius: BorderRadius.circular(16),
                color: Colors.transparent,
              ),
            ),
          ),
          const Positioned(
            bottom: 40,
            left: 20,
            right: 20,
            child: Text(
              'Поместите QR-код или штрих-код в рамку',
              textAlign: TextAlign.center,
              style: TextStyle(color: Colors.white70, fontSize: 14),
            ),
          ),
        ],
      ),
    );
  }
}
