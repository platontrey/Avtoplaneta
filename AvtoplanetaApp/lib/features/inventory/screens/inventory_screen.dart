import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../../../core/models/part.dart';
import '../../auth/providers/auth_provider.dart';
import '../providers/inventory_provider.dart';

class InventoryScreen extends ConsumerStatefulWidget {
  const InventoryScreen({super.key});

  @override
  ConsumerState<InventoryScreen> createState() => _InventoryScreenState();
}

class _InventoryScreenState extends ConsumerState<InventoryScreen> {
  final _searchCtrl = TextEditingController();
  Timer? _debounce;

  @override
  void dispose() {
    _searchCtrl.dispose();
    _debounce?.cancel();
    super.dispose();
  }

  void _onSearch(String value) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 400), () {
      ref.read(inventoryFilterProvider.notifier).update(
            (f) => f.copyWith(search: value, page: 1),
          );
    });
  }

  @override
  Widget build(BuildContext context) {
    final filter = ref.watch(inventoryFilterProvider);
    final inventoryAsync = ref.watch(inventoryProvider(filter));
    final user = ref.watch(authProvider).valueOrNull;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Инвентарь'),
        actions: [
          if (user?.isOperator == true)
            IconButton(
              icon: const Icon(Icons.add),
              onPressed: () => context.go('/inventory/add'),
            ),
          IconButton(
            icon: const Icon(Icons.person_outline),
            onPressed: () => _showUserMenu(context),
          ),
        ],
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(60),
          child: Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
            child: TextField(
              controller: _searchCtrl,
              onChanged: _onSearch,
              style: const TextStyle(color: Colors.white),
              decoration: InputDecoration(
                hintText: 'Поиск запчастей...',
                hintStyle: const TextStyle(color: Colors.white38),
                prefixIcon: const Icon(Icons.search, color: Colors.white38),
                suffixIcon: _searchCtrl.text.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.clear, color: Colors.white38),
                        onPressed: () {
                          _searchCtrl.clear();
                          _onSearch('');
                        },
                      )
                    : null,
                contentPadding: const EdgeInsets.symmetric(vertical: 8),
              ),
            ),
          ),
        ),
      ),
      body: inventoryAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.error_outline, size: 48, color: Colors.red),
              const SizedBox(height: 8),
              Text('Ошибка загрузки', style: Theme.of(context).textTheme.bodyLarge),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: () => ref.invalidate(inventoryProvider),
                child: const Text('Повторить'),
              ),
            ],
          ),
        ),
        data: (data) => Column(
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
              child: Row(
                children: [
                  Text(
                    'Всего: ${data.total} шт.',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Colors.white54,
                        ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: data.parts.isEmpty
                  ? const Center(
                      child: Text(
                        'Запчасти не найдены',
                        style: TextStyle(color: Colors.white54),
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: () async =>
                          ref.invalidate(inventoryProvider(filter)),
                      child: ListView.builder(
                        padding: const EdgeInsets.all(8),
                        itemCount: data.parts.length,
                        itemBuilder: (ctx, i) =>
                            _PartCard(part: data.parts[i]),
                      ),
                    ),
            ),
            // Пагинация
            if (data.total > 20)
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
    );
  }

  void _showUserMenu(BuildContext context) {
    showModalBottomSheet(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
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
}

class _PartCard extends StatelessWidget {
  final Part part;
  const _PartCard({required this.part});

  @override
  Widget build(BuildContext context) {
    const baseUrl = 'https://avtoplaneta.avtoplaneta.crazedns.ru';
    final photoUrl = part.photos.isNotEmpty ? '$baseUrl${part.photos.first}' : null;

    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 4),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () => context.go('/inventory/part/${part.id}'),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              // Фото
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: photoUrl != null
                    ? CachedNetworkImage(
                        imageUrl: photoUrl,
                        width: 64,
                        height: 64,
                        fit: BoxFit.cover,
                        placeholder: (ctx, url) => Container(
                          width: 64,
                          height: 64,
                          color: const Color(0xFF1A1A2E),
                          child: const Icon(Icons.image_outlined,
                              color: Colors.white24),
                        ),
                        errorWidget: (ctx, url, err) => _placeholder(),
                      )
                    : _placeholder(),
              ),
              const SizedBox(width: 12),
              // Инфо
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      part.name,
                      style: const TextStyle(
                          fontWeight: FontWeight.w600, color: Colors.white),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),
                    if (part.brand != null || part.model != null)
                      Text(
                        [part.brand, part.model]
                            .where((e) => e != null && e.isNotEmpty)
                            .join(' • '),
                        style: const TextStyle(
                            color: Colors.white54, fontSize: 12),
                      ),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        _chip(part.category, const Color(0xFF4F8EF7)),
                        const SizedBox(width: 6),
                        if (part.location.isNotEmpty)
                          _chip(part.location, const Color(0xFF43A047)),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              // Цена и количество
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text(
                    '${part.price.toStringAsFixed(0)} ₽',
                    style: const TextStyle(
                      color: Color(0xFF4F8EF7),
                      fontWeight: FontWeight.bold,
                      fontSize: 15,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '${part.quantity} шт.',
                    style: TextStyle(
                      color:
                          part.quantity > 0 ? Colors.white54 : Colors.red,
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
        width: 64,
        height: 64,
        color: const Color(0xFF1A1A2E),
        child: const Icon(Icons.directions_car_outlined, color: Colors.white24),
      );

  Widget _chip(String label, Color color) => Container(
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(4),
          border: Border.all(color: color.withValues(alpha: 0.4), width: 0.5),
        ),
        child: Text(
          label,
          style: TextStyle(color: color, fontSize: 11),
        ),
      );
}

class _Pagination extends StatelessWidget {
  final int current;
  final int total;
  final void Function(int) onPage;
  const _Pagination(
      {required this.current, required this.total, required this.onPage});

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
          Text('$current / $total', style: const TextStyle(color: Colors.white70)),
          IconButton(
            icon: const Icon(Icons.chevron_right),
            onPressed: current < total ? () => onPage(current + 1) : null,
          ),
        ],
      ),
    );
  }
}
