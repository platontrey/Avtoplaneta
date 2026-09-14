import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/part.dart';
import '../../../core/storage/cache_storage.dart';

// Параметры фильтрации
class InventoryFilter {
  final String search;
  final String category;
  final String brand;
  final String model;
  final String location;
  final String salesman;
  final String status;
  final String hasPhoto;
  final String number;
  final String oemCode;
  final String vin;
  final String bodyBrand;
  final String engineBrand;
  final String carReleaseDate;
  final String transmission;
  final String drive;
  final String condition;
  final String manufacturer;
  final String defect;
  final String color;
  final String pageSize; // 'all' (бесконечная лента), '20', '50', '100'
  final int page;

  const InventoryFilter({
    this.search = '',
    this.category = '',
    this.brand = '',
    this.model = '',
    this.location = '',
    this.salesman = '',
    this.status = '',
    this.hasPhoto = 'all',
    this.number = '',
    this.oemCode = '',
    this.vin = '',
    this.bodyBrand = '',
    this.engineBrand = '',
    this.carReleaseDate = '',
    this.transmission = '',
    this.drive = '',
    this.condition = '',
    this.manufacturer = '',
    this.defect = '',
    this.color = '',
    this.pageSize = 'all',
    this.page = 1,
  });

  InventoryFilter copyWith({
    String? search,
    String? category,
    String? brand,
    String? model,
    String? location,
    String? salesman,
    String? status,
    String? hasPhoto,
    String? number,
    String? oemCode,
    String? vin,
    String? bodyBrand,
    String? engineBrand,
    String? carReleaseDate,
    String? transmission,
    String? drive,
    String? condition,
    String? manufacturer,
    String? defect,
    String? color,
    String? pageSize,
    int? page,
  }) => InventoryFilter(
    search: search ?? this.search,
    category: category ?? this.category,
    brand: brand ?? this.brand,
    model: model ?? this.model,
    location: location ?? this.location,
    salesman: salesman ?? this.salesman,
    status: status ?? this.status,
    hasPhoto: hasPhoto ?? this.hasPhoto,
    number: number ?? this.number,
    oemCode: oemCode ?? this.oemCode,
    vin: vin ?? this.vin,
    bodyBrand: bodyBrand ?? this.bodyBrand,
    engineBrand: engineBrand ?? this.engineBrand,
    carReleaseDate: carReleaseDate ?? this.carReleaseDate,
    transmission: transmission ?? this.transmission,
    drive: drive ?? this.drive,
    condition: condition ?? this.condition,
    manufacturer: manufacturer ?? this.manufacturer,
    defect: defect ?? this.defect,
    color: color ?? this.color,
    pageSize: pageSize ?? this.pageSize,
    page: page ?? this.page,
  );

  int get activeFilterCount => [
    category,
    brand,
    model,
    location,
    salesman,
    status,
    hasPhoto == 'all' ? '' : hasPhoto,
    number,
    oemCode,
    vin,
    bodyBrand,
    engineBrand,
    carReleaseDate,
    transmission,
    drive,
    condition,
    manufacturer,
    defect,
    color,
    pageSize == 'all' ? '' : pageSize,
  ].where((value) => value.isNotEmpty).length;

  Map<String, dynamic> toQueryParameters({int? overrideLimit}) {
    final limitVal = overrideLimit ?? (pageSize == 'all' ? 20 : (int.tryParse(pageSize) ?? 20));
    return {
      'page': page,
      'limit': limitVal,
      if (search.isNotEmpty) 'search': search,
      if (category.isNotEmpty) 'category': category,
      if (brand.isNotEmpty) 'brand': brand,
      if (model.isNotEmpty) 'model': model,
      if (location.isNotEmpty) 'location': location,
      if (salesman.isNotEmpty) 'salesman': salesman,
      if (status.isNotEmpty) 'status': status,
      if (hasPhoto != 'all') 'hasPhoto': hasPhoto,
      if (number.isNotEmpty) 'number': number,
      if (oemCode.isNotEmpty) 'oem_code': oemCode,
      if (vin.isNotEmpty) 'vin': vin,
      if (bodyBrand.isNotEmpty) 'body_brand': bodyBrand,
      if (engineBrand.isNotEmpty) 'engine_brand': engineBrand,
      if (carReleaseDate.isNotEmpty) 'car_release_date': carReleaseDate,
      if (transmission.isNotEmpty) 'transmission': transmission,
      if (drive.isNotEmpty) 'drive': drive,
      if (condition.isNotEmpty) 'condition': condition,
      if (manufacturer.isNotEmpty) 'manufacturer': manufacturer,
      if (defect.isNotEmpty) 'defect': defect,
      if (color.isNotEmpty) 'color': color,
    };
  }
}

final inventoryFilterProvider = StateProvider<InventoryFilter>(
  (ref) => const InventoryFilter(),
);

final inventoryProvider =
    FutureProvider.family<InventoryResponse, InventoryFilter>((
      ref,
      filter,
    ) async {
      final params = filter.toQueryParameters();

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

        final effectiveLimit = filter.pageSize == 'all'
            ? 20
            : (int.tryParse(filter.pageSize) ?? 20);
        final hasMore = parts.length == effectiveLimit;
        final calculatedTotal = hasMore
            ? (filter.page + 1) * effectiveLimit
            : (filter.page - 1) * effectiveLimit + parts.length;
        final total = responseData is Map
            ? (responseData['total'] as num?)?.toInt() ?? calculatedTotal
            : calculatedTotal;

        return InventoryResponse(
          parts: parts,
          total: total,
          page: filter.page,
          limit: effectiveLimit,
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
            filteredList = filteredList
                .where(
                  (p) =>
                      p.name.toLowerCase().contains(query) ||
                      (p.brand != null &&
                          p.brand!.toLowerCase().contains(query)) ||
                      (p.model != null &&
                          p.model!.toLowerCase().contains(query)) ||
                      (p.oemCode != null &&
                          p.oemCode!.toLowerCase().contains(query)),
                )
                .toList();
          }
          if (filter.category.isNotEmpty) {
            final value = filter.category.toLowerCase();
            filteredList = filteredList
                .where((part) => part.category.toLowerCase().contains(value))
                .toList();
          }
          if (filter.brand.isNotEmpty) {
            final value = filter.brand.toLowerCase();
            filteredList = filteredList
                .where(
                  (part) => part.brand?.toLowerCase().contains(value) == true,
                )
                .toList();
          }
          if (filter.model.isNotEmpty) {
            final value = filter.model.toLowerCase();
            filteredList = filteredList
                .where(
                  (part) => part.model?.toLowerCase().contains(value) == true,
                )
                .toList();
          }
          if (filter.location.isNotEmpty) {
            final value = filter.location.toLowerCase();
            filteredList = filteredList
                .where((part) => part.location.toLowerCase().contains(value))
                .toList();
          }
          if (filter.salesman.isNotEmpty) {
            final value = filter.salesman.toLowerCase();
            filteredList = filteredList
                .where(
                  (part) =>
                      part.salesman?.toLowerCase().contains(value) == true,
                )
                .toList();
          }
          if (filter.status.isNotEmpty) {
            final isAvailable = filter.status == 'true';
            filteredList = filteredList
                .where((part) => part.isAvailable == isAvailable)
                .toList();
          }
          if (filter.hasPhoto == 'with') {
            filteredList = filteredList
                .where((part) => part.photos.isNotEmpty)
                .toList();
          } else if (filter.hasPhoto == 'without') {
            filteredList = filteredList
                .where((part) => part.photos.isEmpty)
                .toList();
          }
          if (filter.number.isNotEmpty) {
            final value = filter.number.toLowerCase();
            filteredList = filteredList
                .where((part) => part.number?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.oemCode.isNotEmpty) {
            final value = filter.oemCode.toLowerCase();
            filteredList = filteredList
                .where((part) => part.oemCode?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.vin.isNotEmpty) {
            final value = filter.vin.toLowerCase();
            filteredList = filteredList
                .where((part) => part.vin?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.bodyBrand.isNotEmpty) {
            final value = filter.bodyBrand.toLowerCase();
            filteredList = filteredList
                .where((part) => part.bodyBrand?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.engineBrand.isNotEmpty) {
            final value = filter.engineBrand.toLowerCase();
            filteredList = filteredList
                .where((part) => part.engineBrand?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.carReleaseDate.isNotEmpty) {
            final value = filter.carReleaseDate.toLowerCase();
            filteredList = filteredList
                .where((part) => part.carReleaseDate?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.transmission.isNotEmpty) {
            final value = filter.transmission.toLowerCase();
            filteredList = filteredList
                .where((part) => part.transmission?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.drive.isNotEmpty) {
            final value = filter.drive.toLowerCase();
            filteredList = filteredList
                .where((part) => part.drive?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.condition.isNotEmpty) {
            final value = filter.condition.toLowerCase();
            filteredList = filteredList
                .where((part) => part.condition?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.manufacturer.isNotEmpty) {
            final value = filter.manufacturer.toLowerCase();
            filteredList = filteredList
                .where((part) => part.manufacturer?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.defect.isNotEmpty) {
            final value = filter.defect.toLowerCase();
            filteredList = filteredList
                .where((part) => part.defect?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.color.isNotEmpty) {
            final value = filter.color.toLowerCase();
            filteredList = filteredList
                .where((part) => part.color?.toLowerCase().contains(value) == true)
                .toList();
          }

          // Локальная пагинация
          final effectiveLimit = filter.pageSize == 'all'
              ? 20
              : (int.tryParse(filter.pageSize) ?? 20);
          final start = (filter.page - 1) * effectiveLimit;
          final end = start + effectiveLimit;
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
            limit: effectiveLimit,
            isOffline: true,
          );
        }
        rethrow;
      }
    });

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
