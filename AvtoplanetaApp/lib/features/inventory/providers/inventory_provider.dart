import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/part.dart';
import '../../../core/storage/cache_storage.dart';
import '../data/inventory_filter.dart';
export '../data/inventory_filter.dart';

final inventoryFilterProvider = StateProvider<InventoryFilter>(
  (ref) => const InventoryFilter(),
);

bool _matchesFrontRear(String? partVal, String filterVal) {
  if (partVal == null || partVal.isEmpty) return false;
  final p = partVal.trim().toLowerCase();
  final f = filterVal.trim().toLowerCase();
  if (p == f || p.contains(f)) return true;

  if (f == 'перед' || f == 'передний' || f == 'f' || f == 'front') {
    return p == 'f' || p.contains('перед') || p.contains('front') || p.contains('f/r');
  }
  if (f == 'зад' || f == 'задний' || f == 'r' || f == 'rear') {
    return p == 'r' || p.contains('зад') || p.contains('rear') || p.contains('f/r');
  }
  if (f == 'перед / зад' || f == 'перед/зад' || f == 'f/r') {
    return p.contains('перед') || p.contains('зад') || p == 'f' || p == 'r' || p.contains('f/r');
  }
  return false;
}

bool _matchesLeftRight(String? partVal, String filterVal) {
  if (partVal == null || partVal.isEmpty) return false;
  final p = partVal.trim().toLowerCase();
  final f = filterVal.trim().toLowerCase();
  if (p == f || p.contains(f)) return true;

  if (f == 'право' || f == 'правый' || f == 'r' || f == 'right') {
    return p == 'r' || p.contains('прав') || p.contains('right') || p.contains('l/r');
  }
  if (f == 'лево' || f == 'левый' || f == 'l' || f == 'left') {
    return p == 'l' || p.contains('лев') || p.contains('left') || p.contains('l/r');
  }
  if (f == 'лево / право' || f == 'лево/право' || f == 'l/r') {
    return p.contains('лев') || p.contains('прав') || p == 'l' || p == 'r' || p.contains('l/r');
  }
  return false;
}

bool _matchesTopBottom(String? partVal, String filterVal) {
  if (partVal == null || partVal.isEmpty) return false;
  final p = partVal.trim().toLowerCase();
  final f = filterVal.trim().toLowerCase();
  if (p == f || p.contains(f)) return true;

  if (f == 'верх' || f == 'верхний' || f == 't' || f == 'u' || f == 'top' || f == 'upper') {
    return p == 't' || p == 'u' || p.contains('верх') || p.contains('top') || p.contains('upper');
  }
  if (f == 'низ' || f == 'нижний' || f == 'b' || f == 'l' || f == 'bottom' || f == 'lower') {
    return p == 'b' || p == 'l' || p.contains('низ') || p.contains('bottom') || p.contains('lower');
  }
  if (f == 'середина' || f == 'средний' || f == 'm' || f == 'mid' || f == 'middle') {
    return p.contains('середина') || p.contains('средн') || p == 'm' || p == 'mid';
  }
  if (f == 'верх / низ' || f == 'верх/низ') {
    return p.contains('верх') || p.contains('низ') || p == 't' || p == 'b';
  }
  return false;
}

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

          // Локальная фильтрация в кэше с поддержкой мультипоиска
          if (filter.search.isNotEmpty) {
            final fragments = filter.search
                .split(RegExp(r'(?:\s+(?:или|or)\s+|[,\n\r;|]+)', caseSensitive: false))
                .map((f) => f.trim().toLowerCase())
                .where((f) => f.isNotEmpty)
                .toList();

            filteredList = filteredList.where((p) {
              final name = p.name.toLowerCase();
              final brand = p.brand?.toLowerCase() ?? '';
              final model = p.model?.toLowerCase() ?? '';
              final oem = p.oemCode?.toLowerCase() ?? '';
              final number = p.number?.toLowerCase() ?? '';

              // Хотя бы один фрагмент должен совпасть (OR)
              return fragments.any((fragment) {
                final terms = fragment
                    .split(RegExp(r'\s+'))
                    .where((t) => t.isNotEmpty)
                    .toList();
                if (terms.isEmpty) return true;

                // Внутри фрагмента все термины должны совпасть (AND)
                for (final term in terms) {
                  final inText = name.contains(term) ||
                      brand.contains(term) ||
                      model.contains(term) ||
                      oem.contains(term) ||
                      number.contains(term);
                  if (inText) continue;

                  // Проверка совпадения со свойствами положения
                  if ((term.startsWith('перед') || term == 'front') &&
                      _matchesFrontRear(p.frontRear, 'перед')) {
                    continue;
                  }
                  if ((term.startsWith('зад') || term == 'rear') &&
                      _matchesFrontRear(p.frontRear, 'зад')) {
                    continue;
                  }
                  if ((term.startsWith('прав') || term == 'right') &&
                      _matchesLeftRight(p.leftRight, 'право')) {
                    continue;
                  }
                  if ((term.startsWith('лев') || term == 'left') &&
                      _matchesLeftRight(p.leftRight, 'лево')) {
                    continue;
                  }
                  if ((term.startsWith('верх') || term == 'top') &&
                      _matchesTopBottom(p.topBottom, 'верх')) {
                    continue;
                  }
                  if ((term.startsWith('низ') || term == 'bottom') &&
                      _matchesTopBottom(p.topBottom, 'низ')) {
                    continue;
                  }

                  return false;
                }
                return true;
              });
            }).toList();
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
          if (filter.carReleasePeriod.isNotEmpty) {
            final value = filter.carReleasePeriod.toLowerCase();
            filteredList = filteredList
                .where((part) => part.carReleasePeriod?.toLowerCase().contains(value) == true)
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
          if (filter.address.isNotEmpty) {
            final value = filter.address.toLowerCase();
            filteredList = filteredList
                .where((part) => part.address.toLowerCase().contains(value))
                .toList();
          }
          if (filter.minPrice.isNotEmpty) {
            final minP = double.tryParse(filter.minPrice);
            if (minP != null) {
              filteredList = filteredList.where((p) => p.price >= minP).toList();
            }
          }
          if (filter.maxPrice.isNotEmpty) {
            final maxP = double.tryParse(filter.maxPrice);
            if (maxP != null) {
              filteredList = filteredList.where((p) => p.price <= maxP).toList();
            }
          }
          if (filter.minQuantity.isNotEmpty) {
            final minQ = int.tryParse(filter.minQuantity);
            if (minQ != null) {
              filteredList = filteredList.where((p) => p.quantity >= minQ).toList();
            }
          }
          if (filter.maxQuantity.isNotEmpty) {
            final maxQ = int.tryParse(filter.maxQuantity);
            if (maxQ != null) {
              filteredList = filteredList.where((p) => p.quantity <= maxQ).toList();
            }
          }
          if (filter.frontRear.isNotEmpty) {
            filteredList = filteredList
                .where((part) => _matchesFrontRear(part.frontRear, filter.frontRear))
                .toList();
          }
          if (filter.leftRight.isNotEmpty) {
            filteredList = filteredList
                .where((part) => _matchesLeftRight(part.leftRight, filter.leftRight))
                .toList();
          }
          if (filter.topBottom.isNotEmpty) {
            filteredList = filteredList
                .where((part) => _matchesTopBottom(part.topBottom, filter.topBottom))
                .toList();
          }
          if (filter.manufacturerCode.isNotEmpty) {
            final value = filter.manufacturerCode.toLowerCase();
            filteredList = filteredList
                .where((part) => part.manufacturerCode?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.supplierCode.isNotEmpty) {
            final value = filter.supplierCode.toLowerCase();
            filteredList = filteredList
                .where((part) => part.supplierCode?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.transmissionModel.isNotEmpty) {
            final value = filter.transmissionModel.toLowerCase();
            filteredList = filteredList
                .where((part) => part.transmissionModel?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.wearPercentage.isNotEmpty) {
            final value = filter.wearPercentage.toLowerCase();
            filteredList = filteredList
                .where((part) => part.wearPercentage?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.season.isNotEmpty) {
            final value = filter.season.toLowerCase();
            filteredList = filteredList
                .where((part) => part.season?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.diameter.isNotEmpty) {
            final value = filter.diameter.toLowerCase();
            filteredList = filteredList
                .where((part) => part.diameter?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.width.isNotEmpty) {
            final value = filter.width.toLowerCase();
            filteredList = filteredList
                .where((part) => part.width?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.profile.isNotEmpty) {
            final value = filter.profile.toLowerCase();
            filteredList = filteredList
                .where((part) => part.profile?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.tireQuantity.isNotEmpty) {
            final value = filter.tireQuantity.toLowerCase();
            filteredList = filteredList
                .where((part) => part.tireQuantity?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.drilling.isNotEmpty) {
            final value = filter.drilling.toLowerCase();
            filteredList = filteredList
                .where((part) => part.drilling?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.offset.isNotEmpty) {
            final value = filter.offset.toLowerCase();
            filteredList = filteredList
                .where((part) => part.offset?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.centerHoleDiameter.isNotEmpty) {
            final value = filter.centerHoleDiameter.toLowerCase();
            filteredList = filteredList
                .where((part) => part.centerHoleDiameter?.toLowerCase().contains(value) == true)
                .toList();
          }
          if (filter.tireModel.isNotEmpty) {
            final value = filter.tireModel.toLowerCase();
            filteredList = filteredList
                .where((part) => part.tireModel?.toLowerCase().contains(value) == true)
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
