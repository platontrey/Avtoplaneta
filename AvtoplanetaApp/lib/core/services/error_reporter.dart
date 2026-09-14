import 'dart:convert';
import 'dart:io';
import 'dart:math';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'package:path_provider/path_provider.dart';

/// Модель отдельной записи об ошибке приложения
class ClientLogEntry {
  final String id;
  final DateTime timestamp;
  final String type; // 'network', 'flutter', 'fatal', 'custom'
  final String message;
  final String? stackTrace;
  final String? endpoint;
  final int? statusCode;
  final String? method;
  final String? responseBody;
  final String? deviceInfo;
  final String? appVersion;
  final int? buildNumber;
  bool isSent;

  ClientLogEntry({
    required this.id,
    required this.timestamp,
    required this.type,
    required this.message,
    this.stackTrace,
    this.endpoint,
    this.statusCode,
    this.method,
    this.responseBody,
    this.deviceInfo,
    this.appVersion,
    this.buildNumber,
    this.isSent = false,
  });

  Map<String, dynamic> toJson() => {
        'id': id,
        'timestamp': timestamp.toIso8601String(),
        'type': type,
        'message': message,
        if (stackTrace != null) 'stack_trace': stackTrace,
        if (endpoint != null) 'endpoint': endpoint,
        if (statusCode != null) 'status_code': statusCode,
        if (method != null) 'method': method,
        if (responseBody != null) 'response_body': responseBody,
        if (deviceInfo != null) 'device_info': deviceInfo,
        if (appVersion != null) 'app_version': appVersion,
        if (buildNumber != null) 'build_number': buildNumber,
        'is_sent': isSent,
      };

  factory ClientLogEntry.fromJson(Map<String, dynamic> json) => ClientLogEntry(
        id: json['id'] as String? ?? '${DateTime.now().millisecondsSinceEpoch}',
        timestamp: json['timestamp'] != null
            ? DateTime.tryParse(json['timestamp'] as String) ?? DateTime.now()
            : DateTime.now(),
        type: json['type'] as String? ?? 'custom',
        message: json['message'] as String? ?? 'Неизвестная ошибка',
        stackTrace: json['stack_trace'] as String?,
        endpoint: json['endpoint'] as String?,
        statusCode: json['status_code'] as int?,
        method: json['method'] as String?,
        responseBody: json['response_body'] as String?,
        deviceInfo: json['device_info'] as String?,
        appVersion: json['app_version'] as String?,
        buildNumber: json['build_number'] as int?,
        isSent: json['is_sent'] as bool? ?? false,
      );
}

/// Синглтон-сервис сбора, локального хранения и отправки клиентских ошибок
class ErrorReporter {
  static final ErrorReporter _instance = ErrorReporter._internal();
  static ErrorReporter get instance => _instance;

  ErrorReporter._internal();

  final List<ClientLogEntry> _logs = [];
  final ValueNotifier<List<ClientLogEntry>> logsNotifier = ValueNotifier([]);

  String? _appVersion;
  int? _buildNumber;
  String? _deviceInfo;
  bool _initialized = false;
  File? _storageFile;
  final Dio _senderDio = Dio(BaseOptions(
    connectTimeout: const Duration(seconds: 10),
    receiveTimeout: const Duration(seconds: 15),
  ));

  static const int maxLogsCount = 100;
  static const String _defaultBaseUrl = 'https://backend-server.ru';

  /// Инициализация сервиса: загрузка метаданных устройства и локального кэша логов
  Future<void> init() async {
    if (_initialized) return;

    try {
      final info = await PackageInfo.fromPlatform();
      _appVersion = info.version;
      _buildNumber = int.tryParse(info.buildNumber) ?? 0;
    } catch (e) {
      _appVersion = '1.0.1';
      _buildNumber = 2;
    }

    try {
      if (!kIsWeb) {
        _deviceInfo = '${Platform.operatingSystem} ${Platform.operatingSystemVersion}';
      } else {
        _deviceInfo = 'Web';
      }
    } catch (_) {
      _deviceInfo = 'Unknown device';
    }

    try {
      final dir = await getApplicationDocumentsDirectory();
      _storageFile = File('${dir.path}/client_error_logs.json');
      if (await _storageFile!.exists()) {
        final content = await _storageFile!.readAsString();
        if (content.isNotEmpty) {
          final list = jsonDecode(content) as List<dynamic>;
          _logs.clear();
          for (final item in list) {
            try {
              _logs.add(ClientLogEntry.fromJson(item as Map<String, dynamic>));
            } catch (_) {}
          }
          logsNotifier.value = List.unmodifiable(_logs);
        }
      }
    } catch (e) {
      debugPrint('Не удалось загрузить кэш логов: $e');
    }

    _initialized = true;
  }

  List<ClientLogEntry> get logs => List.unmodifiable(_logs);

  String get appVersion => _appVersion ?? '1.0.1';
  int get buildNumber => _buildNumber ?? 2;
  String get deviceInfo => _deviceInfo ?? 'Android';

  /// Запись сетевой ошибки от Dio
  void recordNetworkError({
    required String method,
    required String endpoint,
    int? statusCode,
    dynamic error,
    dynamic responseData,
  }) {
    final sanitizedResponse = _sanitizeData(responseData);
    final errorMessage = _extractNetworkErrorMessage(error, statusCode, endpoint);

    _addLog(ClientLogEntry(
      id: _generateId(),
      timestamp: DateTime.now(),
      type: 'network',
      message: errorMessage,
      method: method,
      endpoint: endpoint,
      statusCode: statusCode,
      responseBody: sanitizedResponse,
      deviceInfo: _deviceInfo,
      appVersion: _appVersion,
      buildNumber: _buildNumber,
    ));
  }

  /// Запись ошибки рендеринга Flutter
  void recordFlutterError(FlutterErrorDetails details) {
    final message = details.exceptionAsString();
    final stack = details.stack?.toString();

    _addLog(ClientLogEntry(
      id: _generateId(),
      timestamp: DateTime.now(),
      type: 'flutter',
      message: message,
      stackTrace: stack,
      deviceInfo: _deviceInfo,
      appVersion: _appVersion,
      buildNumber: _buildNumber,
    ));
  }

  /// Запись общего или фатального сбоя Dart
  void recordError(
    dynamic error,
    StackTrace? stack, {
    String type = 'fatal',
    String? customMessage,
  }) {
    final message = customMessage ?? error.toString();

    _addLog(ClientLogEntry(
      id: _generateId(),
      timestamp: DateTime.now(),
      type: type,
      message: message,
      stackTrace: stack?.toString(),
      deviceInfo: _deviceInfo,
      appVersion: _appVersion,
      buildNumber: _buildNumber,
    ));
  }

  void _addLog(ClientLogEntry entry) {
    _logs.insert(0, entry);
    if (_logs.length > maxLogsCount) {
      _logs.removeRange(maxLogsCount, _logs.length);
    }
    logsNotifier.value = List.unmodifiable(_logs);
    _saveToStorage();
  }

  Future<void> _saveToStorage() async {
    if (_storageFile == null) return;
    try {
      final jsonList = _logs.map((e) => e.toJson()).toList();
      await _storageFile!.writeAsString(jsonEncode(jsonList));
    } catch (e) {
      debugPrint('Ошибка сохранения логов: $e');
    }
  }

  /// Отправляет неотправленные логи пачкой на бэкенд (POST /api/app/logs)
  Future<bool> flush({String baseUrl = _defaultBaseUrl}) async {
    final unsentLogs = _logs.where((e) => !e.isSent).take(50).toList();
    if (unsentLogs.isEmpty) return true;

    final targetUrl = baseUrl.endsWith('/')
        ? '${baseUrl}api/app/logs'
        : '$baseUrl/api/app/logs';

    try {
      final payload = {
        'client': 'AvtoplanetaApp',
        'version': appVersion,
        'build': buildNumber,
        'device': deviceInfo,
        'logs': unsentLogs.map((e) => e.toJson()).toList(),
      };

      final response = await _senderDio.post(
        targetUrl,
        data: payload,
        options: Options(headers: {'Content-Type': 'application/json'}),
      );

      if (response.statusCode == 200) {
        for (final item in unsentLogs) {
          item.isSent = true;
        }
        logsNotifier.value = List.unmodifiable(_logs);
        await _saveToStorage();
        return true;
      }
    } catch (e) {
      debugPrint('Не удалось отправить логи на сервер: $e');
    }
    return false;
  }

  /// Очищает все логи локально
  Future<void> clearLogs() async {
    _logs.clear();
    logsNotifier.value = List.unmodifiable(_logs);
    if (_storageFile != null && await _storageFile!.exists()) {
      try {
        await _storageFile!.delete();
      } catch (_) {}
    }
  }

  /// Формирует структурированный текстовый отчет об ошибках для отправки разработчикам
  String exportReportText() {
    final buffer = StringBuffer();
    buffer.writeln('📋 ОТЧЕТ ОБ ОШИБКАХ ПРИЛОЖЕНИЯ «АВТОПЛАНЕТА»');
    buffer.writeln('═══════════════════════════════════════════');
    buffer.writeln('📱 Устройство: $deviceInfo');
    buffer.writeln('📦 Версия: v$appVersion (сборка $buildNumber)');
    buffer.writeln('🕒 Время генерации: ${DateTime.now().toLocal()}');
    buffer.writeln('🔢 Всего ошибок в журнале: ${_logs.length}');
    buffer.writeln('═══════════════════════════════════════════\n');

    if (_logs.isEmpty) {
      buffer.writeln('✅ Ошибок не зафиксировано.');
      return buffer.toString();
    }

    for (int i = 0; i < _logs.length; i++) {
      final log = _logs[i];
      buffer.writeln('[$i] [${log.type.toUpperCase()}] ${log.timestamp.toLocal()}');
      if (log.method != null && log.endpoint != null) {
        buffer.writeln('  URL: ${log.method} ${log.endpoint} (Статус: ${log.statusCode ?? 'нет'})');
      }
      buffer.writeln('  Сообщение: ${log.message}');
      if (log.responseBody != null && log.responseBody!.isNotEmpty) {
        buffer.writeln('  Ответ сервера: ${log.responseBody}');
      }
      if (log.stackTrace != null && log.stackTrace!.isNotEmpty) {
        final lines = log.stackTrace!.split('\n').take(8).join('\n');
        buffer.writeln('  Стек вызовов (первые строки):\n$lines');
      }
      buffer.writeln('───────────────────────────────────────────');
    }

    return buffer.toString();
  }

  String _generateId() {
    final rand = Random().nextInt(0xFFFFFF).toRadixString(16).padLeft(6, '0');
    return '${DateTime.now().millisecondsSinceEpoch}_$rand';
  }

  String? _sanitizeData(dynamic data) {
    if (data == null) return null;
    try {
      if (data is Map<String, dynamic>) {
        final cleaned = <String, dynamic>{};
        for (final entry in data.entries) {
          final lower = entry.key.toLowerCase();
          if (lower.contains('password') ||
              lower.contains('token') ||
              lower.contains('secret') ||
              lower.contains('authorization')) {
            cleaned[entry.key] = '***MASKED***';
          } else {
            cleaned[entry.key] = entry.value;
          }
        }
        final str = jsonEncode(cleaned);
        return str.length > 1000 ? '${str.substring(0, 1000)}... [обрезано]' : str;
      }
      final str = data.toString();
      return str.length > 1000 ? '${str.substring(0, 1000)}... [обрезано]' : str;
    } catch (_) {
      return null;
    }
  }

  String _extractNetworkErrorMessage(dynamic error, int? statusCode, String endpoint) {
    if (error is DioException) {
      if (statusCode != null) {
        switch (statusCode) {
          case 401:
            return '401 Unauthorized: сессия недействительна или отсутствует доступ ($endpoint)';
          case 403:
            return '403 Forbidden: недостаточно прав для выполнения операции ($endpoint)';
          case 404:
            return '404 Not Found: ресурс не найден ($endpoint)';
          case 405:
            return '405 Method Not Allowed: недопустимый метод запроса ($endpoint)';
          case 500:
            return '500 Внутренняя ошибка сервера ($endpoint)';
          case 502:
            return '502 Bad Gateway: шлюз не смог связаться с бэкенд-сервисом ($endpoint)';
          case 503:
            return '503 Service Unavailable: сервис временно недоступен ($endpoint)';
          default:
            return 'HTTP $statusCode при запросе к $endpoint: ${error.message}';
        }
      }
      return 'Сетевой сбой Dio (${error.type.name}): ${error.message} ($endpoint)';
    }
    return error?.toString() ?? 'Ошибка запроса к $endpoint';
  }
}

/// Riverpod провайдер для удобного доступа к сервису логов
final errorReporterProvider = Provider<ErrorReporter>((ref) {
  return ErrorReporter.instance;
});
