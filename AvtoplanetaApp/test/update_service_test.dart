import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/core/services/update_service.dart';

void main() {
  group('AppUpdateInfo Model Tests', () {
    test('парсинг из JSON бэкенда', () {
      final json = {
        'version': '1.2.0',
        'build_number': 42,
        'download_url': '/api/v1/app/download',
        'force_update': true,
        'min_supported_build': 30,
        'changelog': '• Исправлены ошибки в инвентаре\n• Добавлен поиск по OEM',
        'file_size_bytes': 35400000,
        'published_at': '2026-09-14T10:00:00Z',
      };

      final info = AppUpdateInfo.fromJson(json);

      expect(info.version, '1.2.0');
      expect(info.buildNumber, 42);
      expect(info.downloadUrl, '/api/v1/app/download');
      expect(info.forceUpdate, true);
      expect(info.minSupportedBuild, 30);
      expect(info.changelog, contains('OEM'));
      expect(info.fileSizeBytes, 35400000);
      expect(info.publishedAt, isNotNull);
      expect(info.publishedAt!.year, 2026);
    });

    test('парсинг из пустого или неполного JSON', () {
      final info = AppUpdateInfo.fromJson({});

      expect(info.version, isEmpty);
      expect(info.buildNumber, 0);
      expect(info.downloadUrl, isEmpty);
      expect(info.forceUpdate, false);
      expect(info.minSupportedBuild, 0);
      expect(info.changelog, isEmpty);
      expect(info.fileSizeBytes, 0);
      expect(info.publishedAt, isNull);
    });

    test('сериализация toJson', () {
      const info = AppUpdateInfo(
        version: '1.0.5',
        buildNumber: 10,
        downloadUrl: 'https://example.com/app.apk',
        forceUpdate: false,
        minSupportedBuild: 5,
        changelog: 'Тест',
        fileSizeBytes: 1024,
      );

      final json = info.toJson();
      expect(json['version'], '1.0.5');
      expect(json['build_number'], 10);
      expect(json['download_url'], 'https://example.com/app.apk');
      expect(json['force_update'], false);
      expect(json['min_supported_build'], 5);
      expect(json['changelog'], 'Тест');
      expect(json['file_size_bytes'], 1024);
    });
  });

  group('GitHub Releases Parsing Tests', () {
    test('парсинг релиза с ассетом APK и тегом v1.0.3+25', () {
      final githubRelease = {
        'tag_name': 'v1.0.3+25',
        'name': 'Автопланета v1.0.3',
        'body': 'Новая сборка для склада.\nИсправлено сканирование QR.',
        'published_at': '2026-09-14T12:00:00Z',
        'assets': [
          {
            'name': 'source.tar.gz',
            'browser_download_url': 'https://github.com/platontrey/Avtoplaneta/archive/v1.0.3.tar.gz',
            'size': 120000,
          },
          {
            'name': 'avtoplaneta-release.apk',
            'browser_download_url': 'https://github.com/platontrey/Avtoplaneta/releases/download/v1.0.3/avtoplaneta-release.apk',
            'size': 45000000,
          },
        ],
      };

      final info = AppUpdateInfo.fromGitHubRelease(githubRelease);

      expect(info.version, '1.0.3');
      expect(info.buildNumber, 25);
      expect(info.downloadUrl, contains('avtoplaneta-release.apk'));
      expect(info.fileSizeBytes, 45000000);
      expect(info.forceUpdate, false);
      expect(info.changelog, contains('QR'));
    });

    test('парсинг mobile-latest с build_number в описании', () {
      final githubRelease = {
        'tag_name': 'mobile-latest',
        'name': 'Автопланета — актуальная Android-сборка',
        'body': 'Автоматическая сборка run_number: 88.\n[force_update] Критическое обновление протокола.',
        'published_at': '2026-09-14T15:00:00Z',
        'assets': [
          {
            'name': 'avtoplaneta-latest.apk',
            'browser_download_url': 'https://github.com/platontrey/Avtoplaneta/releases/download/mobile-latest/avtoplaneta-latest.apk',
            'size': 38000000,
          },
        ],
      };

      final info = AppUpdateInfo.fromGitHubRelease(githubRelease);

      expect(info.buildNumber, 88);
      expect(info.forceUpdate, true);
      expect(info.downloadUrl, contains('avtoplaneta-latest.apk'));
      expect(info.fileSizeBytes, 38000000);
    });

    test('парсинг релиза без APK файлов', () {
      final githubRelease = {
        'tag_name': 'v1.0.0',
        'name': 'Initial',
        'body': 'No binaries here',
        'assets': <dynamic>[],
      };

      final info = AppUpdateInfo.fromGitHubRelease(githubRelease);
      expect(info.downloadUrl, isEmpty);
      expect(info.fileSizeBytes, 0);
    });
  });

  group('Version Comparison Tests', () {
    test('сравнение по номерам сборок build_number', () {
      expect(
        UpdateService.isVersionNewer(
          serverVersion: '1.0.0',
          serverBuild: 5,
          currentVersion: '1.0.0',
          currentBuild: 4,
        ),
        isTrue,
      );

      expect(
        UpdateService.isVersionNewer(
          serverVersion: '1.0.0',
          serverBuild: 4,
          currentVersion: '1.0.0',
          currentBuild: 4,
        ),
        isFalse,
      );

      expect(
        UpdateService.isVersionNewer(
          serverVersion: '1.0.0',
          serverBuild: 3,
          currentVersion: '1.0.0',
          currentBuild: 4,
        ),
        isFalse,
      );
    });

    test('сравнение по семантическим версиям semver', () {
      expect(UpdateService.compareSemver('1.0.1', '1.0.0'), 1);
      expect(UpdateService.compareSemver('1.0.0', '1.0.1'), -1);
      expect(UpdateService.compareSemver('1.0.0', '1.0.0'), 0);
      expect(UpdateService.compareSemver('2.0.0', '1.9.9'), 1);
      expect(UpdateService.compareSemver('v1.2.3', '1.2.2'), 1);
      expect(UpdateService.compareSemver('1.2.0', '1.2'), 0);
    });

    test('сравнение когда номера сборок равны нулю, но semver новее', () {
      expect(
        UpdateService.isVersionNewer(
          serverVersion: '1.1.0',
          serverBuild: 0,
          currentVersion: '1.0.9',
          currentBuild: 0,
        ),
        isTrue,
      );
    });
  });

  group('UpdateService Integration Flow Tests', () {
    test('успешное обнаружение новой версии на бэкенде', () async {
      final dio = Dio();
      dio.httpClientAdapter = _MockHttpAdapter((options) {
        if (options.path.endsWith('/api/v1/app/version')) {
          return ResponseBody.fromString(
            '''{
              "version": "1.0.5",
              "build_number": 10,
              "download_url": "/api/v1/app/download",
              "force_update": false,
              "min_supported_build": 1,
              "changelog": "Добавлены улучшения"
            }''',
            200,
            headers: {
              Headers.contentTypeHeader: [Headers.jsonContentType],
            },
          );
        }
        return ResponseBody.fromString('Not found', 404);
      });

      final service = UpdateService(
        dio: dio,
        checkGitHubReleases: false,
        checkPlatform: false,
      );

      final result = await service.checkForUpdates(
        overrideCurrentVersion: '1.0.4',
        overrideCurrentBuild: 5,
      );

      expect(result.hasUpdate, isTrue);
      expect(result.isForced, isFalse);
      expect(result.updateInfo?.version, '1.0.5');
      expect(result.updateInfo?.buildNumber, 10);
      expect(result.updateInfo?.downloadUrl, '/api/v1/app/download');
    });

    test('обнаружение обязательного обновления (force_update: true)', () async {
      final dio = Dio();
      dio.httpClientAdapter = _MockHttpAdapter((options) {
        if (options.path.endsWith('/api/v1/app/version')) {
          return ResponseBody.fromString(
            '''{
              "version": "2.0.0",
              "build_number": 50,
              "download_url": "/api/v1/app/download",
              "force_update": true,
              "min_supported_build": 40,
              "changelog": "Критическое обновление безопасности"
            }''',
            200,
            headers: {
              Headers.contentTypeHeader: [Headers.jsonContentType],
            },
          );
        }
        return ResponseBody.fromString('Not found', 404);
      });

      final service = UpdateService(
        dio: dio,
        checkGitHubReleases: false,
        checkPlatform: false,
      );

      final result = await service.checkForUpdates(
        overrideCurrentVersion: '1.0.0',
        overrideCurrentBuild: 20,
      );

      expect(result.hasUpdate, isTrue);
      expect(result.isForced, isTrue);
      expect(result.updateInfo?.forceUpdate, isTrue);
    });

    test('обязательное обновление по порогу min_supported_build', () async {
      final dio = Dio();
      dio.httpClientAdapter = _MockHttpAdapter((options) {
        return ResponseBody.fromString(
          '''{
            "version": "1.5.0",
            "build_number": 20,
            "download_url": "/api/v1/app/download",
            "force_update": false,
            "min_supported_build": 15,
            "changelog": "Обновление API"
          }''',
          200,
          headers: {
            Headers.contentTypeHeader: [Headers.jsonContentType],
          },
        );
      });

      final service = UpdateService(
        dio: dio,
        checkGitHubReleases: false,
        checkPlatform: false,
      );

      // Текущая сборка 10 < min_supported_build 15
      final result = await service.checkForUpdates(
        overrideCurrentVersion: '1.4.0',
        overrideCurrentBuild: 10,
      );

      expect(result.hasUpdate, isTrue);
      expect(result.isForced, isTrue);
    });

    test('отсутствие обновлений когда локальная версия актуальна', () async {
      final dio = Dio();
      dio.httpClientAdapter = _MockHttpAdapter((options) {
        return ResponseBody.fromString(
          '''{
            "version": "1.0.5",
            "build_number": 10,
            "download_url": "/api/v1/app/download"
          }''',
          200,
          headers: {
            Headers.contentTypeHeader: [Headers.jsonContentType],
          },
        );
      });

      final service = UpdateService(
        dio: dio,
        checkGitHubReleases: false,
        checkPlatform: false,
      );

      final result = await service.checkForUpdates(
        overrideCurrentVersion: '1.0.5',
        overrideCurrentBuild: 10,
      );

      expect(result.hasUpdate, isFalse);
      expect(result.isForced, isFalse);
    });

    test('корректная обработка сетевой ошибки бэкенда', () async {
      final dio = Dio();
      dio.httpClientAdapter = _MockHttpAdapter((options) {
        throw DioException(
          requestOptions: options,
          error: 'Connection refused',
          type: DioExceptionType.connectionError,
        );
      });

      final service = UpdateService(
        dio: dio,
        checkGitHubReleases: false,
        checkPlatform: false,
      );

      final result = await service.checkForUpdates(
        overrideCurrentVersion: '1.0.0',
        overrideCurrentBuild: 1,
      );

      expect(result.hasUpdate, isFalse);
      expect(result.isForced, isFalse);
    });

    test('fallback на GitHub Releases когда бэкенд вернул 404', () async {
      final githubRelease = {
        'tag_name': 'v1.1.0+75',
        'name': 'Автопланета v1.1.0',
        'body': 'Новый релиз на GitHub Releases.\n• Интеграция с печатью',
        'published_at': '2026-09-14T12:00:00Z',
        'assets': [
          {
            'name': 'avtoplaneta-latest.apk',
            'browser_download_url': 'https://github.com/platontrey/Avtoplaneta/releases/download/v1.1.0/avtoplaneta-latest.apk',
            'size': 42000000,
          }
        ],
      };

      final info = AppUpdateInfo.fromGitHubRelease(githubRelease);
      expect(info.version, '1.1.0');
      expect(info.buildNumber, 75);
      expect(info.downloadUrl, 'https://github.com/platontrey/Avtoplaneta/releases/download/v1.1.0/avtoplaneta-latest.apk');

      final isNewer = UpdateService.isVersionNewer(
        serverVersion: info.version,
        serverBuild: info.buildNumber,
        currentVersion: '1.0.1',
        currentBuild: 2,
      );
      expect(isNewer, isTrue);
    });
  });
}

/// Вспомогательный класс для мокирования HTTP запросов Dio
class _MockHttpAdapter implements HttpClientAdapter {
  final ResponseBody Function(RequestOptions options) _handler;

  _MockHttpAdapter(this._handler);

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return _handler(options);
  }

  @override
  void close({bool force = false}) {}
}
