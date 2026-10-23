import 'dart:async';
import 'dart:developer';
import '../api/api_client.dart';
import '../storage/cache_storage.dart';

class SyncService {
  static Timer? _initialTimer;
  static Timer? _syncTimer;
  static bool _isSyncing = false;

  /// Запуск периодического фонового кэширования
  static void startPeriodicSync() {
    _initialTimer?.cancel();
    _syncTimer?.cancel();

    // Откладываем первый запуск на 20 секунд, чтобы не конкурировать с основной загрузкой экрана склада
    _initialTimer = Timer(const Duration(seconds: 20), () {
      final lastCached = CacheStorage.getCachedAt();
      if (lastCached == null ||
          DateTime.now().difference(lastCached) > const Duration(minutes: 10)) {
        _sync();
      }
    });

    _syncTimer = Timer.periodic(const Duration(minutes: 15), (_) => _sync());
  }

  /// Остановка синхронизации
  static void stopSync() {
    _initialTimer?.cancel();
    _initialTimer = null;
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
          '/api/v1/inventory',
          queryParameters: {
            'page': page,
            'limit': limit,
          },
        );

        final data = response.data;
        final List<dynamic> list = data is List
            ? data
            : data is Map && data['parts'] is List
                ? data['parts'] as List<dynamic>
                : [];
        allParts.addAll(list);

        if (list.length < limit || allParts.length >= 5000) {
          hasMore = false;
        } else {
          page++;
          // Пауза между страницами, чтобы не забивать мобильный канал и не вызывать rate-limit
          await Future<void>.delayed(const Duration(milliseconds: 150));
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
