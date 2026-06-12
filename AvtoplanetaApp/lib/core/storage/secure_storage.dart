import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SecureStorage {
  static const _storage = FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
  );

  static const _keyToken = 'jwt_token';
  static const _keyRefreshToken = 'jwt_refresh_token';
  static const _keyUser = 'cached_user';

  static Future<void> saveTokens({
    required String token,
    required String refreshToken,
  }) async {
    await Future.wait([
      _storage.write(key: _keyToken, value: token),
      _storage.write(key: _keyRefreshToken, value: refreshToken),
    ]);
  }

  static Future<String?> getToken() => _storage.read(key: _keyToken);

  static Future<String?> getRefreshToken() =>
      _storage.read(key: _keyRefreshToken);

  static Future<void> saveUser(String userJson) =>
      _storage.write(key: _keyUser, value: userJson);

  static Future<String?> getUser() => _storage.read(key: _keyUser);

  static Future<void> clearUser() => _storage.delete(key: _keyUser);

  static Future<void> clearTokens() async {
    await Future.wait([
      _storage.delete(key: _keyToken),
      _storage.delete(key: _keyRefreshToken),
      _storage.delete(key: _keyUser),
    ]);
  }
}
