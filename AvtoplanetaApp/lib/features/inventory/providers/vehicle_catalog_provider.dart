import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/storage/cache_storage.dart';
import '../data/vehicle_catalog.dart';

/// Справочник марок и моделей. Кэшируется так же, как каталог запчастей:
/// при недоступной сети форма продолжает работать на последней копии.
final vehicleCatalogProvider = FutureProvider<VehicleCatalog>((ref) async {
  try {
    final response = await apiClient.dio.get('/api/vehicle-catalog');
    final json = Map<String, dynamic>.from(response.data as Map);
    await CacheStorage.saveVehicleCatalog(json);
    return VehicleCatalog.fromJson(json);
  } catch (_) {
    final cached = CacheStorage.getVehicleCatalog();
    if (cached != null) {
      return VehicleCatalog.fromJson(cached);
    }
    rethrow;
  }
});
