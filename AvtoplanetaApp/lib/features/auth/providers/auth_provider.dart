import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/user.dart';
import '../../../core/storage/secure_storage.dart';

// Провайдер текущего пользователя
final authProvider = StateNotifierProvider<AuthNotifier, AsyncValue<User?>>(
  (ref) => AuthNotifier(),
);

class AuthNotifier extends StateNotifier<AsyncValue<User?>> {
  AuthNotifier() : super(const AsyncValue.loading()) {
    _init();
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
      state = AsyncValue.data(user);
    } catch (_) {
      // Токен протух — пробуем обновить
      try {
        await _tryRefresh();
      } catch (_) {
        await SecureStorage.clearTokens();
        state = const AsyncValue.data(null);
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
      final loginResponse =
          LoginResponse.fromJson(response.data as Map<String, dynamic>);

      if (loginResponse.token.isNotEmpty) {
        await SecureStorage.saveTokens(
          token: loginResponse.token,
          refreshToken: loginResponse.refreshToken,
        );
      }
      state = AsyncValue.data(loginResponse.user);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      rethrow;
    }
  }

  Future<void> logout() async {
    try {
      await apiClient.dio.post('/auth/logout');
    } catch (_) {}
    await SecureStorage.clearTokens();
    state = const AsyncValue.data(null);
  }

  Future<void> _tryRefresh() async {
    final refreshToken = await SecureStorage.getRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) throw Exception('no refresh token');

    final response = await apiClient.dio.post(
      '/auth/refresh',
      data: {'refresh_token': refreshToken},
    );
    final newToken = response.data['token'] as String;
    final newRefresh = response.data['refresh_token'] as String;
    await SecureStorage.saveTokens(token: newToken, refreshToken: newRefresh);

    final meResponse = await apiClient.dio.get('/auth/me');
    final user = User.fromJson(meResponse.data as Map<String, dynamic>);
    state = AsyncValue.data(user);
  }
}
