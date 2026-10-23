import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../services/error_reporter.dart';
import '../storage/secure_storage.dart';

const _baseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'https://backend-server.ru',
);

class ApiClient {
  static final ApiClient _instance = ApiClient._internal();
  factory ApiClient() => _instance;

  late final Dio _dio;
  late final _AuthInterceptor _authInterceptor;
  void Function()? onUnauthorized;

  ApiClient._internal() {
    _dio = Dio(BaseOptions(
      baseUrl: _baseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 30),
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    ));

    _authInterceptor = _AuthInterceptor(_dio);
    _dio.interceptors.add(_ContentTypeValidatorInterceptor());
    _dio.interceptors.add(_authInterceptor);
    _dio.interceptors.add(_ErrorLoggingInterceptor());

    if (kDebugMode) {
      _dio.interceptors.add(LogInterceptor(
        requestBody: true,
        responseBody: true,
        error: true,
      ));
    }
  }

  Dio get dio => _dio;

  /// Дедуплицированное обновление access/refresh пары токенов.
  /// Возвращает новый access-токен при успехе или выбрасывает исходную ошибку.
  Future<String?> refreshAccessToken() => _authInterceptor.refreshToken();

  /// Проверяет, вызвана ли ошибка временным сбоем сети или недоступностью шлюза.
  static bool isTransientNetworkError(Object err) {
    if (err is SocketException || err is TimeoutException) {
      return true;
    }
    if (err is DioException) {
      final type = err.type;
      if (type == DioExceptionType.connectionTimeout ||
          type == DioExceptionType.sendTimeout ||
          type == DioExceptionType.receiveTimeout ||
          type == DioExceptionType.connectionError) {
        return true;
      }
      if (type == DioExceptionType.unknown &&
          (err.error is SocketException || err.response == null)) {
        return true;
      }
      final status = err.response?.statusCode;
      if (status == 502 || status == 503 || status == 504) {
        return true;
      }
    }
    return false;
  }

  /// Преобразует относительный путь из API в публичный URL.
  String resolveUrl(String path) {
    final uri = Uri.tryParse(path);
    if (uri != null && uri.hasScheme) return path;
    return Uri.parse(_dio.options.baseUrl).resolve(path).toString();
  }
}

/// Перехватчик для автоматического добавления JWT и обновления токена
class _AuthInterceptor extends Interceptor {
  final Dio _dio;
  Completer<String?>? _refreshCompleter;

  _AuthInterceptor(this._dio);

  bool _isAuthEndpoint(String path) {
    return path.contains('/auth/login') ||
        path.contains('/auth/refresh') ||
        path.contains('/auth/google') ||
        path.contains('/auth/logout');
  }

  /// Выполняет обновление токена с объединением параллельных запросов в единую очередь.
  Future<String?> refreshToken() async {
    if (_refreshCompleter != null) {
      return _refreshCompleter!.future;
    }

    final completer = Completer<String?>();
    _refreshCompleter = completer;

    try {
      final storedRefresh = await SecureStorage.getRefreshToken();
      if (storedRefresh == null || storedRefresh.isEmpty) {
        await SecureStorage.clearTokens();
        apiClient.onUnauthorized?.call();
        completer.complete(null);
        return null;
      }

      final refreshDio = Dio(BaseOptions(
        baseUrl: _dio.options.baseUrl,
        connectTimeout: const Duration(seconds: 12),
        receiveTimeout: const Duration(seconds: 12),
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      ));

      final response = await refreshDio.post(
        '/auth/refresh',
        data: {'refresh_token': storedRefresh},
      );

      final data = response.data;
      if (data is! Map) {
        throw DioException(
          requestOptions: response.requestOptions,
          response: response,
          type: DioExceptionType.badResponse,
          error: 'Некорректный формат ответа /auth/refresh',
        );
      }

      final newToken = data['token'] as String?;
      final newRefresh = data['refresh_token'] as String?;
      if (newToken == null ||
          newToken.isEmpty ||
          newRefresh == null ||
          newRefresh.isEmpty) {
        throw DioException(
          requestOptions: response.requestOptions,
          response: response,
          type: DioExceptionType.badResponse,
          error: 'Пустой токен в ответе /auth/refresh',
        );
      }

      await SecureStorage.saveTokens(
        token: newToken,
        refreshToken: newRefresh,
      );

      completer.complete(newToken);
      return newToken;
    } catch (e) {
      final statusCode = e is DioException ? e.response?.statusCode : null;
      final isExplicitTokenRejection = statusCode == 401 || statusCode == 403;

      // Разлогиниваем пользователя ТОЛЬКО если сервер явно отверг refresh_token (401/403).
      // При сетевых сбоях, таймаутах или 5xx сохраняем токены для следующих запросов.
      if (isExplicitTokenRejection) {
        await SecureStorage.clearTokens();
        apiClient.onUnauthorized?.call();
      }

      completer.complete(null);
      rethrow;
    } finally {
      _refreshCompleter = null;
    }
  }

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final token = await SecureStorage.getToken();
    if (token != null && token.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $token';
    }

    // messaging-service использует тот же идентификатор пользователя, что и сайт.
    if (options.path.startsWith('/api/messaging/')) {
      final cachedUser = await SecureStorage.getUser();
      if (cachedUser != null) {
        try {
          final user = jsonDecode(cachedUser) as Map<String, dynamic>;
          final userId = user['id'];
          if (userId != null) options.headers['X-User-ID'] = userId.toString();
        } catch (_) {
          // Повреждённый кэш не должен блокировать остальные API-запросы.
        }
      }
    }
    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final statusCode = err.response?.statusCode;
    final alreadyRetried = err.requestOptions.extra['_auth_retried'] == true;

    if (statusCode == 401 &&
        !alreadyRetried &&
        !_isAuthEndpoint(err.requestOptions.path)) {
      String? newToken;
      try {
        newToken = await refreshToken();
      } catch (_) {
        handler.next(err);
        return;
      }

      if (newToken == null || newToken.isEmpty) {
        handler.next(err);
        return;
      }

      // Повторяем оригинальный запрос в отдельном try/catch, чтобы ошибка
      // самого бизнес-запроса (400/404/500/таймаут) никогда не стирала токены.
      try {
        final retryOptions = err.requestOptions;
        retryOptions.extra['_auth_retried'] = true;
        retryOptions.headers['Authorization'] = 'Bearer $newToken';
        final retryResponse = await _dio.fetch(retryOptions);
        handler.resolve(retryResponse);
      } on DioException catch (retryErr) {
        handler.next(retryErr);
      } catch (_) {
        handler.next(err);
      }
    } else {
      handler.next(err);
    }
  }
}

/// Перехватчик для автоматической фиксации сетевых ошибок API в ErrorReporter
class _ErrorLoggingInterceptor extends Interceptor {
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    // Исключаем эндпоинт отправки самих логов, чтобы избежать рекурсии
    final path = err.requestOptions.path;
    if (!path.contains('/api/app/logs')) {
      ErrorReporter.instance.recordNetworkError(
        method: err.requestOptions.method,
        endpoint: path,
        statusCode: err.response?.statusCode,
        error: err,
        responseData: err.response?.data,
      );
    }
    handler.next(err);
  }
}

/// Перехватчик для защиты от ошибочных HTML-ответов сервера (например, при некорректной маршрутизации SPA или сбое прокси).
class _ContentTypeValidatorInterceptor extends Interceptor {
  @override
  void onResponse(Response response, ResponseInterceptorHandler handler) {
    final contentType =
        response.headers.value('content-type')?.toLowerCase() ?? '';
    final data = response.data;

    final isHtmlContent = contentType.contains('text/html');
    final isHtmlString = data is String &&
        (data.trimLeft().toLowerCase().startsWith('<!doctype html') ||
            data.trimLeft().toLowerCase().startsWith('<html'));

    if (isHtmlContent || isHtmlString) {
      final path = response.requestOptions.path;
      final error = DioException(
        requestOptions: response.requestOptions,
        response: response,
        type: DioExceptionType.badResponse,
        error:
            'Ошибка маршрутизации API: сервер вернул HTML-страницу вместо JSON данных для эндпоинта "$path". '
            'Проверьте конфигурацию маршрутизации бэкенда.',
      );
      handler.reject(error);
      return;
    }

    handler.next(response);
  }
}

/// Сингтлон Dio для использования в провайдерах
final apiClient = ApiClient();
