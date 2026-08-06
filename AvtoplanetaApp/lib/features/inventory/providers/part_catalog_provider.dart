import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/storage/cache_storage.dart';
import '../data/part_catalog.dart';

final partCatalogProvider = FutureProvider<PartCatalog>((ref) async {
  try {
    final response = await apiClient.dio.get('/api/part-catalog');
    final json = Map<String, dynamic>.from(response.data as Map);
    await CacheStorage.savePartCatalog(json);
    return PartCatalog.fromJson(json);
  } catch (_) {
    final cached = CacheStorage.getPartCatalog();
    if (cached != null) {
      return PartCatalog.fromJson(cached);
    }
    rethrow;
  }
});
