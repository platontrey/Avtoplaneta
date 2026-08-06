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
  Timer? _debounce;
  final Set<int> _selectedPartIds = {};
  bool _selectionMode = false;

  @override
  void dispose() {
    _searchCtrl.dispose();
    _debounce?.cancel();
    super.dispose();
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
    final selectedParts = visibleParts
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
              tooltip: _selectedPartIds.length == visibleParts.length
                  ? 'Снять выбор со всех'
                  : 'Выбрать все на странице',
              icon: Icon(
                _selectedPartIds.length == visibleParts.length
                    ? Icons.deselect_rounded
                    : Icons.select_all_rounded,
              ),
              onPressed: () {
                setState(() {
                  if (_selectedPartIds.length == visibleParts.length) {
                    _selectedPartIds.clear();
                    _selectionMode = false;
                  } else {
                    _selectedPartIds.addAll(
                      visibleParts.map((part) => part.id),
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
          onAction: () => ref.invalidate(inventoryProvider),
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
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 5,
                    ),
                    decoration: BoxDecoration(
                      color: AppTheme.primaryColor.withValues(alpha: 0.12),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Text(
                      '${data.total} позиций',
                      style: const TextStyle(
                        color: AppTheme.primaryColor,
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: data.parts.isEmpty
                  ? const AppEmptyState(
                      icon: Icons.search_off_rounded,
                      title: 'Ничего не найдено',
                      message:
                          'Попробуйте изменить запрос или очистить строку поиска.',
                    )
                  : RefreshIndicator(
                      onRefresh: () async =>
                          ref.invalidate(inventoryProvider(filter)),
                      child: ListView.builder(
                        padding: const EdgeInsets.fromLTRB(16, 6, 16, 16),
                        itemCount: data.parts.length,
                        itemBuilder: (ctx, i) => Padding(
                          padding: const EdgeInsets.only(bottom: 10),
                          child: _PartCard(
                            part: data.parts[i],
                            selected: _selectedPartIds.contains(
                              data.parts[i].id,
                            ),
                            selectionMode: _selectionMode,
                            onTap: () {
                              if (_selectionMode) {
                                _toggleSelection(data.parts[i]);
                              } else {
                                context.go(
                                  '/inventory/part/${data.parts[i].id}',
                                );
                              }
                            },
                            onLongPress: () => _toggleSelection(data.parts[i]),
                          ),
                        ),
                      ),
                    ),
            ),
            // Пагинация
            if (data.total > 20 && !_selectionMode)
              _Pagination(
                current: filter.page,
                total: (data.total / 20).ceil(),
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
                  context.go('/admin');
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
  late final TextEditingController _salesmanController;
  late String _category;
  late String _status;
  late String _hasPhoto;

  @override
  void initState() {
    super.initState();
    _brandController = TextEditingController(text: widget.current.brand);
    _modelController = TextEditingController(text: widget.current.model);
    _locationController = TextEditingController(text: widget.current.location);
    _salesmanController = TextEditingController(text: widget.current.salesman);
    _category = widget.current.category;
    _status = widget.current.status;
    _hasPhoto = widget.current.hasPhoto;
  }

  @override
  void dispose() {
    _brandController.dispose();
    _modelController.dispose();
    _locationController.dispose();
    _salesmanController.dispose();
    super.dispose();
  }

  void _reset() {
    setState(() {
      _brandController.clear();
      _modelController.clear();
      _locationController.clear();
      _salesmanController.clear();
      _category = '';
      _status = '';
      _hasPhoto = 'all';
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
        salesman: _salesmanController.text.trim(),
        status: _status,
        hasPhoto: _hasPhoto,
        page: 1,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final categories = {
      ...widget.categories.where((category) => category.isNotEmpty),
      if (_category.isNotEmpty) _category,
    }.toList()..sort();
    final keyboardInset = MediaQuery.viewInsetsOf(context).bottom;

    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 8, 8, 8),
          child: Row(
            children: [
              Expanded(
                child: Text(
                  'Фильтры запчастей',
                  style: Theme.of(context).textTheme.titleLarge,
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
        const Divider(height: 1),
        Expanded(
          child: ListView(
            padding: EdgeInsets.fromLTRB(20, 20, 20, keyboardInset + 20),
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
                decoration: const InputDecoration(
                  labelText: 'Бренд',
                  hintText: 'BMW, Toyota…',
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _modelController,
                decoration: const InputDecoration(
                  labelText: 'Модель',
                  hintText: 'E90, A4…',
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _locationController,
                decoration: const InputDecoration(
                  labelText: 'Местоположение',
                  hintText: 'Стеллаж A-12…',
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _salesmanController,
                decoration: const InputDecoration(labelText: 'Продавец'),
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
              const SizedBox(height: 24),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: _reset,
                      icon: const Icon(Icons.clear_all_rounded),
                      label: const Text('Сбросить'),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: FilledButton.icon(
                      onPressed: _apply,
                      icon: const Icon(Icons.check_rounded),
                      label: const Text('Применить'),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
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
