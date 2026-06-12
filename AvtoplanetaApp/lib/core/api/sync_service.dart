import 'dart:async';
import 'dart:developer';
import '../api/api_client.dart';
import '../storage/cache_storage.dart';

class SyncService {
  static Timer? _syncTimer;
  static bool _isSyncing = false;

  /// Запуск периодического фонового кэширования
  static void startPeriodicSync() {
    _syncTimer?.cancel();
    // Запускаем синхронизацию сразу, затем каждые 5 минут
    _sync();
    _syncTimer = Timer.periodic(const Duration(minutes: 5), (_) => _sync());
  }

  /// Остановка синхронизации
  static void stopSync() {
    _syncTimer?.cancel();
    _syncTimer = null;
  }

  /// Процесс фонового скачивания каталога запчастей без картинок
  static Future<void> _sync() async {
    if (_isSyncing) return;
    _isSyncing = true;
    log('Background inventory sync started...');

    try {
      final List<dynamic> allParts = [];
      int page = 1;
      const limit = 200;
      bool hasMore = true;

      while (hasMore) {
        final response = await apiClient.dio.get(
          '/api/inventory',
          queryParameters: {
            'page': page,
            'limit': limit,
            // Передаем параметры, если необходимо оптимизировать выдачу на бэкенде
          },
        );

        final List<dynamic> list = response.data is List ? response.data as List : [];
        allParts.addAll(list);

        if (list.length < limit || allParts.length >= 5000) {
          hasMore = false;
        } else {
          page++;
        }
      }

      if (allParts.isNotEmpty) {
        await CacheStorage.savePartsCache(allParts);
        log('Background sync completed. Cached ${allParts.length} parts.');
      }
    } catch (e) {
      log('Background sync failed: $e');
    } finally {
      _isSyncing = false;
    }
  }
}
