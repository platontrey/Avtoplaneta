import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../storage/secure_storage.dart';

const _baseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'https://backend-server.ru',
);

class ApiClient {
  static final ApiClient _instance = ApiClient._internal();
  factory ApiClient() => _instance;

  late final Dio _dio;

  ApiClient._internal() {
    _dio = Dio(BaseOptions(
      baseUrl: _baseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 30),
      headers: {'Content-Type': 'application/json'},
    ));

    _dio.interceptors.add(_AuthInterceptor(_dio));

    if (kDebugMode) {
      _dio.interceptors.add(LogInterceptor(
        requestBody: true,
        responseBody: true,
        error: true,
      ));
    }
  }

  Dio get dio => _dio;

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
  bool _isRefreshing = false;

  _AuthInterceptor(this._dio);

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
    if (err.response?.statusCode == 401 && !_isRefreshing) {
      _isRefreshing = true;
      try {
        final refreshToken = await SecureStorage.getRefreshToken();
        if (refreshToken == null || refreshToken.isEmpty) {
          await SecureStorage.clearTokens();
          handler.next(err);
          return;
        }

        // Запрашиваем новый токен
        final response = await _dio.post(
          '/auth/refresh',
          data: {'refresh_token': refreshToken},
          options: Options(headers: {'Authorization': ''}), // без токена
        );

        final newToken = response.data['token'] as String;
        final newRefresh = response.data['refresh_token'] as String;
        await SecureStorage.saveTokens(
            token: newToken, refreshToken: newRefresh);

        // Повторяем оригинальный запрос с новым токеном
        final retryOptions = err.requestOptions;
        retryOptions.headers['Authorization'] = 'Bearer $newToken';
        final retryResponse = await _dio.fetch(retryOptions);
        handler.resolve(retryResponse);
      } catch (_) {
        await SecureStorage.clearTokens();
        handler.next(err);
      } finally {
        _isRefreshing = false;
      }
    } else {
      handler.next(err);
    }
  }
}

/// Сингтлон Dio для использования в провайдерах
final apiClient = ApiClient();
