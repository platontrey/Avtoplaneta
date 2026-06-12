import 'dart:convert';
import 'package:crypto/crypto.dart';

class QrSigner {
  static const String secretSeed = 'avtoplaneta_secret_qr_seed_2026';

  /// Generates signed string in format: ap:{partId}:{timestamp}:{hmac}
  static String generateQrData({required int partId, required DateTime? createdAt}) {
    final timestamp = createdAt != null ? createdAt.millisecondsSinceEpoch ~/ 1000 : 1771653694;
    final dataToSign = '$partId:$timestamp';
    
    final key = utf8.encode(secretSeed);
    final bytes = utf8.encode(dataToSign);
    
    final hmac = Hmac(sha256, key);
    final signature = hmac.convert(bytes).toString();
    
    return 'ap:$partId:$timestamp:$signature';
  }

  /// Verifies if the HMAC signature matches the partId and timestamp
  static bool verify(int partId, int timestamp, String hmacToVerify) {
    final dataToSign = '$partId:$timestamp';
    
    final key = utf8.encode(secretSeed);
    final bytes = utf8.encode(dataToSign);
    
    final hmac = Hmac(sha256, key);
    final expectedSignature = hmac.convert(bytes).toString();
    
    return expectedSignature == hmacToVerify;
  }
}
