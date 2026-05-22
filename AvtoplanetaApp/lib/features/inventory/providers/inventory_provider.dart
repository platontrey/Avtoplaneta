import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/part.dart';

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

    final response = await apiClient.dio.get(
      '/api/inventory',
      queryParameters: params,
    );
    return InventoryResponse.fromJson(response.data as Map<String, dynamic>);
  },
);

// Провайдер для одной запчасти
final partProvider = FutureProvider.family<Part, int>((ref, id) async {
  final response = await apiClient.dio.get('/api/inventory/$id');
  return Part.fromJson(response.data as Map<String, dynamic>);
});
