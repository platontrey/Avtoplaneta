import 'dart:io';
import 'package:dio/dio.dart';
import 'package:dio/io.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:open_filex/open_filex.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'package:path_provider/path_provider.dart';

import '../api/api_client.dart';

/// Информация о версии приложения (с бэкенда или GitHub Releases)
class AppUpdateInfo {
  final String version;
  final int buildNumber;
  final String downloadUrl;
  final bool forceUpdate;
  final int minSupportedBuild;
  final String changelog;
  final int fileSizeBytes;
  final DateTime? publishedAt;

  const AppUpdateInfo({
    required this.version,
    required this.buildNumber,
    required this.downloadUrl,
    required this.forceUpdate,
    required this.minSupportedBuild,
    required this.changelog,
    this.fileSizeBytes = 0,
    this.publishedAt,
  });

  /// Парсинг ответа от собственного бэкенда (/api/app/version)
  factory AppUpdateInfo.fromJson(Map<String, dynamic> json) {
    return AppUpdateInfo(
      version: json['version'] as String? ?? '',
      buildNumber: json['build_number'] as int? ?? 0,
      downloadUrl: json['download_url'] as String? ?? '',
      forceUpdate: json['force_update'] as bool? ?? false,
      minSupportedBuild: json['min_supported_build'] as int? ?? 0,
      changelog: json['changelog'] as String? ?? '',
      fileSizeBytes: json['file_size_bytes'] as int? ?? 0,
      publishedAt: json['published_at'] != null
          ? DateTime.tryParse(json['published_at'] as String)
          : null,
    );
  }

  /// Парсинг ответа GitHub Releases API (https://api.github.com/repos/.../releases/...)
  factory AppUpdateInfo.fromGitHubRelease(Map<String, dynamic> json) {
    final tagName = (json['tag_name'] as String? ?? '').trim();
    final name = json['name'] as String? ?? '';
    final body = json['body'] as String? ?? '';

    // Находим ассет APK
    final assets = (json['assets'] as List<dynamic>?) ?? [];
    String downloadUrl = '';
    int size = 0;

    for (final asset in assets) {
      if (asset is Map<String, dynamic>) {
        final assetName = (asset['name'] as String? ?? '').toLowerCase();
        if (assetName.endsWith('.apk')) {
          downloadUrl = asset['browser_download_url'] as String? ?? '';
          size = asset['size'] as int? ?? 0;
          break;
        }
      }
    }

    // Извлекаем версию (например "v1.0.2+5" -> version="1.0.2", build=5)
    String version = tagName.replaceFirst(RegExp(r'^v', caseSensitive: false), '');
    int buildNumber = 0;

    if (version.contains('+')) {
      final parts = version.split('+');
      version = parts[0];
      buildNumber = int.tryParse(parts[1]) ?? 0;
    }

    // Если в теле релиза есть указание сборки (например, run_number или build: 12)
    final buildMatch = RegExp(
      r'(?:build|сборка|run_number)[\s:=#]+(\d+)',
      caseSensitive: false,
    ).firstMatch(body);
    if (buildMatch != null) {
      buildNumber = int.tryParse(buildMatch.group(1)!) ?? buildNumber;
    }

    final isForced = body.contains('[force_update]') ||
        body.toLowerCase().contains('обязательное обновление');

    return AppUpdateInfo(
      version: version.isEmpty ? (name.isNotEmpty ? name : 'Новая версия') : version,
      buildNumber: buildNumber,
      downloadUrl: downloadUrl,
      forceUpdate: isForced,
      minSupportedBuild: 0,
      changelog: body.isNotEmpty ? body : (name.isNotEmpty ? name : 'Обновление приложения'),
      fileSizeBytes: size,
      publishedAt: json['published_at'] != null
          ? DateTime.tryParse(json['published_at'] as String)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'version': version,
      'build_number': buildNumber,
      'download_url': downloadUrl,
      'force_update': forceUpdate,
      'min_supported_build': minSupportedBuild,
      'changelog': changelog,
      'file_size_bytes': fileSizeBytes,
      'published_at': publishedAt?.toIso8601String(),
    };
  }
}

/// Результат проверки обновлений
class UpdateCheckResult {
  final bool hasUpdate;
  final bool isForced;
  final AppUpdateInfo? updateInfo;
  final String currentVersion;
  final int currentBuildNumber;

  const UpdateCheckResult({
    required this.hasUpdate,
    required this.isForced,
    this.updateInfo,
    required this.currentVersion,
    required this.currentBuildNumber,
  });
}

class UpdateService {
  final String _baseUrl;
  final Dio _dio;
  final Dio? _downloadDio;
  final String githubRepo;
  final bool checkGitHubReleases;
  final bool checkPlatform;

  UpdateService({
    String? baseUrl,
    Dio? dio,
    Dio? downloadDio,
    this.githubRepo = 'platontrey/Avtoplaneta',
    this.checkGitHubReleases = true,
    this.checkPlatform = true,
  })  : _baseUrl = baseUrl ?? apiClient.dio.options.baseUrl,
        _downloadDio = downloadDio,
        _dio = dio ??
            Dio(BaseOptions(
              baseUrl: baseUrl ?? apiClient.dio.options.baseUrl,
              connectTimeout: const Duration(seconds: 10),
              receiveTimeout: const Duration(seconds: 15),
              headers: {'Content-Type': 'application/json'},
            ));

  /// Получает информацию о текущей установленной версии приложения
  Future<PackageInfo> getCurrentPackageInfo() async {
    return await PackageInfo.fromPlatform();
  }

  /// Утилита сравнения семантических версий и номеров сборок
  static bool isVersionNewer({
    required String serverVersion,
    required int serverBuild,
    required String currentVersion,
    required int currentBuild,
  }) {
    // 1. Если сервер отдает номер сборки > 0 и у клиента есть номер сборки > 0
    if (serverBuild > 0 && currentBuild > 0) {
      if (serverBuild > currentBuild) return true;
      if (serverBuild < currentBuild) return false;
    }

    // 2. Сравнение по семантическим версиям (например "1.0.2" > "1.0.1")
    return compareSemver(serverVersion, currentVersion) > 0;
  }

  /// Сравнивает две строки версий: возвращает 1 если v1 > v2, -1 если v1 < v2, 0 если равны
  static int compareSemver(String v1, String v2) {
    final clean1 = v1
        .replaceFirst(RegExp(r'^v', caseSensitive: false), '')
        .split('+')
        .first
        .split('-')
        .first
        .trim();
    final clean2 = v2
        .replaceFirst(RegExp(r'^v', caseSensitive: false), '')
        .split('+')
        .first
        .split('-')
        .first
        .trim();

    if (clean1.isEmpty && clean2.isEmpty) return 0;
    if (clean1.isEmpty) return -1;
    if (clean2.isEmpty) return 1;

    final parts1 = clean1.split('.').map((e) => int.tryParse(e) ?? 0).toList();
    final parts2 = clean2.split('.').map((e) => int.tryParse(e) ?? 0).toList();

    final maxLen = parts1.length > parts2.length ? parts1.length : parts2.length;
    for (int i = 0; i < maxLen; i++) {
      final p1 = i < parts1.length ? parts1[i] : 0;
      final p2 = i < parts2.length ? parts2[i] : 0;
      if (p1 > p2) return 1;
      if (p1 < p2) return -1;
    }
    return 0;
  }

  /// Проверяет наличие обновлений (с бэкенда или с GitHub Releases)
  Future<UpdateCheckResult> checkForUpdates({
    String? overrideCurrentVersion,
    int? overrideCurrentBuild,
    bool? overrideCheckPlatform,
  }) async {
    String currentVersion = overrideCurrentVersion ?? '1.0.0';
    int currentBuild = overrideCurrentBuild ?? 0;

    if (overrideCurrentVersion == null || overrideCurrentBuild == null) {
      try {
        final packageInfo = await getCurrentPackageInfo();
        currentVersion = packageInfo.version;
        currentBuild = int.tryParse(packageInfo.buildNumber) ?? 0;
      } catch (e) {
        debugPrint('Не удалось получить локальную версию PackageInfo: $e');
      }
    }

    final shouldCheckPlatform = overrideCheckPlatform ?? checkPlatform;
    if (shouldCheckPlatform && !kIsWeb && !Platform.isAndroid) {
      return UpdateCheckResult(
        hasUpdate: false,
        isForced: false,
        currentVersion: currentVersion,
        currentBuildNumber: currentBuild,
      );
    }

    // 1. Сначала пробуем получить конфигурацию с собственного бэкенда
    try {
      final response = await _dio.get(
        '/api/app/version',
        options: Options(responseType: ResponseType.json),
      );

      if (response.statusCode == 200 && response.data != null) {
        final data = response.data;
        final jsonMap = data is Map
            ? Map<String, dynamic>.from(data)
            : <String, dynamic>{};
        final serverInfo = AppUpdateInfo.fromJson(jsonMap);

        final hasNewer = isVersionNewer(
          serverVersion: serverInfo.version,
          serverBuild: serverInfo.buildNumber,
          currentVersion: currentVersion,
          currentBuild: currentBuild,
        );

        final isMandatory = serverInfo.forceUpdate ||
            (serverInfo.minSupportedBuild > 0 && serverInfo.minSupportedBuild > currentBuild);

        if (hasNewer && serverInfo.downloadUrl.isNotEmpty) {
          return UpdateCheckResult(
            hasUpdate: true,
            isForced: isMandatory,
            updateInfo: serverInfo,
            currentVersion: currentVersion,
            currentBuildNumber: currentBuild,
          );
        }
      }
    } catch (e) {
      debugPrint('Проверка через бэкенд не удалась: $e');
    }

    // 2. Если бэкенд недоступен или включен fallback на GitHub Releases
    if (checkGitHubReleases && githubRepo.isNotEmpty) {
      final githubResult = await checkGitHubUpdate(
        currentVersion: currentVersion,
        currentBuild: currentBuild,
      );
      if (githubResult.hasUpdate) {
        return githubResult;
      }
    }

    return UpdateCheckResult(
      hasUpdate: false,
      isForced: false,
      currentVersion: currentVersion,
      currentBuildNumber: currentBuild,
    );
  }

  /// Проверка обновлений напрямую через GitHub Releases API
  Future<UpdateCheckResult> checkGitHubUpdate({
    required String currentVersion,
    required int currentBuild,
  }) async {
    final client = Dio(BaseOptions(
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 15),
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'User-Agent': 'AvtoplanetaApp-Updater',
      },
    ));

    final urls = [
      'https://api.github.com/repos/$githubRepo/releases/tags/mobile-latest',
      'https://api.github.com/repos/$githubRepo/releases/latest',
    ];

    for (final url in urls) {
      try {
        final response = await client.get(url);
        if (response.statusCode == 200 && response.data is Map<String, dynamic>) {
          final info = AppUpdateInfo.fromGitHubRelease(response.data as Map<String, dynamic>);
          if (info.downloadUrl.isNotEmpty) {
            final hasNewer = isVersionNewer(
              serverVersion: info.version,
              serverBuild: info.buildNumber,
              currentVersion: currentVersion,
              currentBuild: currentBuild,
            );

            if (hasNewer) {
              return UpdateCheckResult(
                hasUpdate: true,
                isForced: info.forceUpdate,
                updateInfo: info,
                currentVersion: currentVersion,
                currentBuildNumber: currentBuild,
              );
            }
          }
        }
      } catch (e) {
        debugPrint('Ошибка запроса GitHub Release ($url): $e');
      }
    }

    return UpdateCheckResult(
      hasUpdate: false,
      isForced: false,
      currentVersion: currentVersion,
      currentBuildNumber: currentBuild,
    );
  }

  /// Скачивает APK и запускает системный инсталлятор
  Future<OpenResult> downloadAndInstall({
    required AppUpdateInfo info,
    required void Function(double progress, int receivedBytes, int totalBytes) onProgress,
    CancelToken? cancelToken,
  }) async {
    final tempDir = await getTemporaryDirectory();
    final targetFile = File('${tempDir.path}/avtoplaneta-update.apk');

    String finalUrl = info.downloadUrl;
    if (!finalUrl.startsWith('http://') && !finalUrl.startsWith('https://')) {
      final uri = Uri.tryParse(finalUrl);
      if (uri != null && uri.hasScheme) {
        finalUrl = uri.toString();
      } else {
        finalUrl = Uri.parse(_baseUrl).resolve(finalUrl).toString();
      }
    }

    final downloadDio = _downloadDio ?? Dio(BaseOptions(
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(minutes: 10),
      followRedirects: true,
      maxRedirects: 5,
    ));

    if (_downloadDio == null && !kIsWeb) {
      downloadDio.httpClientAdapter = IOHttpClientAdapter(
        createHttpClient: () {
          final client = HttpClient();
          client.idleTimeout = const Duration(seconds: 60);
          client.connectionTimeout = const Duration(seconds: 30);
          client.autoUncompress = false;
          return client;
        },
      );
    }

    int retries = 0;
    const maxRetries = 3;
    bool downloadSuccess = false;

    while (retries < maxRetries && !downloadSuccess) {
      if (cancelToken?.isCancelled ?? false) {
        throw DioException(
          requestOptions: RequestOptions(path: finalUrl),
          type: DioExceptionType.cancel,
          error: 'Загрузка отменена пользователем',
        );
      }

      try {
        final currentBytes = targetFile.existsSync() ? targetFile.lengthSync() : 0;
        final headers = <String, dynamic>{};
        bool isResuming = false;

        // Если файл частично скачан (более 2 МБ) и это первая попытка докачки, пробуем Range
        if (currentBytes > 2 * 1024 * 1024 && retries == 1) {
          headers['Range'] = 'bytes=$currentBytes-';
          isResuming = true;
          debugPrint('Попытка докачки APK с байта $currentBytes (Range)');
        } else if (retries == 0 || retries > 1) {
          // При первой попытке или после неудачной докачки начинаем начисто
          if (targetFile.existsSync()) {
            try {
              await targetFile.delete();
            } catch (_) {}
          }
        }

        final response = await downloadDio.download(
          finalUrl,
          targetFile.path,
          fileAccessMode: isResuming ? FileAccessMode.append : FileAccessMode.write,
          options: Options(headers: headers),
          cancelToken: cancelToken,
          deleteOnError: false,
          onReceiveProgress: (received, total) {
            final effectiveReceived = isResuming ? currentBytes + received : received;
            final effectiveTotal = isResuming && total > 0 ? currentBytes + total : total;
            if (effectiveTotal > 0) {
              final progress = (effectiveReceived / effectiveTotal).clamp(0.0, 1.0);
              onProgress(progress, effectiveReceived, effectiveTotal);
            } else {
              onProgress(0, effectiveReceived, effectiveTotal);
            }
          },
        );

        if (response.statusCode == 200 || response.statusCode == 206) {
          downloadSuccess = true;
        }
      } on DioException catch (e) {
        if (CancelToken.isCancel(e)) {
          rethrow;
        }
        retries++;
        debugPrint('Сбой загрузки APK (попытка $retries из $maxRetries): $e');
        if (retries >= maxRetries) {
          rethrow;
        }
        await Future.delayed(Duration(seconds: retries));
      } catch (e) {
        retries++;
        debugPrint('Неожиданный сбой загрузки APK (попытка $retries из $maxRetries): $e');
        if (retries >= maxRetries) {
          rethrow;
        }
        await Future.delayed(Duration(seconds: retries));
      }
    }

    final openResult = await OpenFilex.open(
      targetFile.path,
      type: 'application/vnd.android.package-archive',
    );

    return openResult;
  }
}

/// Провайдер сервиса автообновлений (использует независимый чистый Dio без AuthInterceptor)
final updateServiceProvider = Provider<UpdateService>((ref) {
  return UpdateService();
});
