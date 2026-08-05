import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/part.dart';
import '../../../core/storage/cache_storage.dart';

// Параметры фильтрации
class InventoryFilter {
  final String search;
  final String category;
  final String brand;
  final String location;
  final int page;

  const InventoryFilter({
    this.search = '',
    this.category = '',
    this.brand = '',
    this.location = '',
    this.page = 1,
  });

  InventoryFilter copyWith({
    String? search,
    String? category,
    String? brand,
    String? location,
    int? page,
  }) =>
      InventoryFilter(
        search: search ?? this.search,
        category: category ?? this.category,
        brand: brand ?? this.brand,
        location: location ?? this.location,
        page: page ?? this.page,
      );
}

final inventoryFilterProvider =
    StateProvider<InventoryFilter>((ref) => const InventoryFilter());

final inventoryProvider =
    FutureProvider.family<InventoryResponse, InventoryFilter>(
  (ref, filter) async {
    final params = <String, dynamic>{
      'page': filter.page,
      'limit': 20,
    };
    if (filter.search.isNotEmpty) params['search'] = filter.search;
    if (filter.category.isNotEmpty) params['category'] = filter.category;
    if (filter.brand.isNotEmpty) params['brand'] = filter.brand;
    if (filter.location.isNotEmpty) params['location'] = filter.location;

    try {
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
      final parts = list
          .map((e) => Part.fromJson(e as Map<String, dynamic>))
          .toList();

      final hasMore = parts.length == 20;
      final calculatedTotal = hasMore
          ? (filter.page + 1) * 20
          : (filter.page - 1) * 20 + parts.length;
      final total = responseData is Map
          ? (responseData['total'] as num?)?.toInt() ?? calculatedTotal
          : calculatedTotal;

      return InventoryResponse(
        parts: parts,
        total: total,
        page: filter.page,
        limit: 20,
        isOffline: false,
      );
    } catch (e) {
      // Попытка взять данные из локального оффлайн-кэша
      final cachedList = CacheStorage.getPartsCache();
      if (cachedList != null) {
        var filteredList = cachedList
            .map((e) => Part.fromJson(Map<String, dynamic>.from(e as Map)))
            .toList();

        // Локальная фильтрация в кэше
        if (filter.search.isNotEmpty) {
          final query = filter.search.toLowerCase();
          filteredList = filteredList.where((p) =>
              p.name.toLowerCase().contains(query) ||
              (p.brand != null && p.brand!.toLowerCase().contains(query)) ||
              (p.model != null && p.model!.toLowerCase().contains(query)) ||
              (p.oemCode != null && p.oemCode!.toLowerCase().contains(query))).toList();
        }
        if (filter.category.isNotEmpty) {
          filteredList = filteredList.where((p) => p.category == filter.category).toList();
        }
        if (filter.brand.isNotEmpty) {
          filteredList = filteredList.where((p) => p.brand == filter.brand).toList();
        }
        if (filter.location.isNotEmpty) {
          filteredList = filteredList.where((p) => p.location == filter.location).toList();
        }

        // Локальная пагинация
        final start = (filter.page - 1) * 20;
        final end = start + 20;
        final partsPage = start >= filteredList.length
            ? <Part>[]
            : filteredList.sublist(
                start,
                end > filteredList.length ? filteredList.length : end,
              );

        return InventoryResponse(
          parts: partsPage,
          total: filteredList.length,
          page: filter.page,
          limit: 20,
          isOffline: true,
        );
      }
      rethrow;
    }
  },
);

// Провайдер для одной запчасти с поддержкой локального поиска при оффлайне
final partProvider = FutureProvider.family<Part, int>((ref, id) async {
  try {
    final response = await apiClient.dio.get('/api/v1/parts/item/$id');
    return Part.fromJson(response.data as Map<String, dynamic>);
  } catch (e) {
    // В оффлайне пробуем найти деталь в сохраненном кэше
    final cachedList = CacheStorage.getPartsCache();
    if (cachedList != null) {
      try {
        final match = cachedList
            .map((e) => Part.fromJson(Map<String, dynamic>.from(e as Map)))
            .firstWhere((p) => p.id == id);
        return match;
      } catch (_) {
        // не нашли деталь с таким id в кэше
      }
    }
    rethrow;
  }
});
