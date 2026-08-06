import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_sign_in/google_sign_in.dart';
import '../../../core/api/api_client.dart';
import '../../../core/api/sync_service.dart';
import '../../../core/models/user.dart';
import '../../../core/storage/secure_storage.dart';

const _googleServerClientId = String.fromEnvironment(
  'GOOGLE_SERVER_CLIENT_ID',
  defaultValue: '',
);

// Провайдер текущего пользователя
final authProvider = StateNotifierProvider<AuthNotifier, AsyncValue<User?>>(
  (ref) => AuthNotifier(),
);

class AuthNotifier extends StateNotifier<AsyncValue<User?>> {
  AuthNotifier({User? initialUser, bool initialize = true})
    : super(
        initialUser == null
            ? const AsyncValue.loading()
            : AsyncValue.data(initialUser),
      ) {
    if (initialize) {
      _init();
    }
  }

  bool _isNetworkError(Object err) {
    if (err is DioException) {
      final type = err.type;
      if (type == DioExceptionType.connectionTimeout ||
          type == DioExceptionType.sendTimeout ||
          type == DioExceptionType.receiveTimeout ||
          type == DioExceptionType.connectionError) {
        return true;
      }
      if (err.response == null) {
        return true;
      }
      final status = err.response?.statusCode;
      if (status != null && status >= 500) {
        return true;
      }
    }
    return false;
  }

  Future<void> _loadCachedUser() async {
    try {
      final cachedJson = await SecureStorage.getUser();
      if (cachedJson != null) {
        final userMap = jsonDecode(cachedJson) as Map<String, dynamic>;
        state = AsyncValue.data(User.fromJson(userMap));
        SyncService.startPeriodicSync();
      } else {
        state = const AsyncValue.data(null);
      }
    } catch (_) {
      state = const AsyncValue.data(null);
    }
  }

  /// Проверяем сохранённый токен при старте
  Future<void> _init() async {
    try {
      final token = await SecureStorage.getToken();
      if (token == null || token.isEmpty) {
        state = const AsyncValue.data(null);
        return;
      }
      final response = await apiClient.dio.get('/auth/me');
      final user = User.fromJson(response.data as Map<String, dynamic>);
      await SecureStorage.saveUser(jsonEncode(user.toJson()));
      state = AsyncValue.data(user);
      SyncService.startPeriodicSync();
    } catch (e) {
      if (_isNetworkError(e)) {
        await _loadCachedUser();
        return;
      }
      // Токен протух или иная ошибка аутентификации — пробуем обновить
      try {
        await _tryRefresh();
      } catch (refreshErr) {
        if (_isNetworkError(refreshErr)) {
          await _loadCachedUser();
        } else {
          await SecureStorage.clearTokens();
          state = const AsyncValue.data(null);
        }
      }
    }
  }

  Future<void> login(String email, String password) async {
    state = const AsyncValue.loading();
    try {
      final response = await apiClient.dio.post(
        '/auth/login',
        data: {'email': email, 'password': password},
      );
      final loginResponse = LoginResponse.fromJson(
        response.data as Map<String, dynamic>,
      );

      if (loginResponse.token.isNotEmpty) {
        await SecureStorage.saveTokens(
          token: loginResponse.token,
          refreshToken: loginResponse.refreshToken,
        );
        await SecureStorage.saveUser(jsonEncode(loginResponse.user.toJson()));
        SyncService.startPeriodicSync();
      }
      state = AsyncValue.data(loginResponse.user);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<void> loginWithGoogle() async {
    state = const AsyncValue.loading();
    try {
      await GoogleSignIn.instance.initialize(
        serverClientId: _googleServerClientId.isEmpty
            ? null
            : _googleServerClientId,
      );
      final GoogleSignInAccount googleUser = await GoogleSignIn.instance
          .authenticate();
      final String? idToken = googleUser.authentication.idToken;
      if (idToken == null || idToken.isEmpty) {
        throw Exception('Не удалось получить Google ID Token');
      }

      final response = await apiClient.dio.post(
        '/auth/google/mobile',
        data: {'id_token': idToken},
      );
      final loginResponse = LoginResponse.fromJson(
        response.data as Map<String, dynamic>,
      );

      if (loginResponse.token.isNotEmpty) {
        await SecureStorage.saveTokens(
          token: loginResponse.token,
          refreshToken: loginResponse.refreshToken,
        );
        await SecureStorage.saveUser(jsonEncode(loginResponse.user.toJson()));
        SyncService.startPeriodicSync();
      }
      state = AsyncValue.data(loginResponse.user);
    } catch (e, st) {
      final errStr = e.toString();
      if (errStr.contains('canceled') || errStr.contains('Sign in failed')) {
        await _loadCachedUser();
        return;
      }
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<void> logout() async {
    SyncService.stopSync();
    try {
      await apiClient.dio.post('/auth/logout');
    } catch (_) {}
    try {
      await GoogleSignIn.instance.signOut();
    } catch (_) {}
    await SecureStorage.clearTokens();
    state = const AsyncValue.data(null);
  }

  Future<void> _tryRefresh() async {
    final refreshToken = await SecureStorage.getRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) {
      throw Exception('no refresh token');
    }

    final response = await apiClient.dio.post(
      '/auth/refresh',
      data: {'refresh_token': refreshToken},
    );
    final newToken = response.data['token'] as String;
    final newRefresh = response.data['refresh_token'] as String;
    await SecureStorage.saveTokens(token: newToken, refreshToken: newRefresh);

    final meResponse = await apiClient.dio.get('/auth/me');
    final user = User.fromJson(meResponse.data as Map<String, dynamic>);
    await SecureStorage.saveUser(jsonEncode(user.toJson()));
    state = AsyncValue.data(user);
    SyncService.startPeriodicSync();
  }
}
