import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SecureStorage {
  static const _storage = FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
  );

  static const _keyToken = 'jwt_token';
  static const _keyRefreshToken = 'jwt_refresh_token';
  static const _keyUser = 'cached_user';
  static const _keySavedLogin = 'saved_login';
  static const _keySavedPassword = 'saved_password';
  static const _keyHandledUpdateBuild = 'handled_update_build';

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

  static Future<void> saveCredentials({
    required String login,
    required String password,
  }) async {
    await Future.wait([
      _storage.write(key: _keySavedLogin, value: login),
      _storage.write(key: _keySavedPassword, value: password),
    ]);
  }

  static Future<String?> getSavedLogin() => _storage.read(key: _keySavedLogin);

  static Future<String?> getSavedPassword() =>
      _storage.read(key: _keySavedPassword);

  static Future<void> clearSavedCredentials() async {
    await Future.wait([
      _storage.delete(key: _keySavedLogin),
      _storage.delete(key: _keySavedPassword),
    ]);
  }

  static Future<void> saveHandledUpdateBuild(int buildNumber) =>
      _storage.write(
        key: _keyHandledUpdateBuild,
        value: buildNumber.toString(),
      );

  static Future<int> getHandledUpdateBuild() async {
    final val = await _storage.read(key: _keyHandledUpdateBuild);
    return int.tryParse(val ?? '') ?? 0;
  }

  static Future<void> clearTokens() async {
    await Future.wait([
      _storage.delete(key: _keyToken),
      _storage.delete(key: _keyRefreshToken),
      _storage.delete(key: _keyUser),
    ]);
  }
}
