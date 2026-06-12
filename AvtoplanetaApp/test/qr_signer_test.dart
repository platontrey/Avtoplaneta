import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/core/utils/qr_signer.dart';

void main() {
  group('QrSigner Tests', () {
    final testPartId = 12345;
    final testCreatedAt = DateTime.fromMillisecondsSinceEpoch(1771653694000);

    test('generate QR data format', () {
      final qrData = QrSigner.generateQrData(partId: testPartId, createdAt: testCreatedAt);
      expect(qrData.startsWith('ap:'), true);
      
      final parts = qrData.split(':');
      expect(parts.length, 4);
      expect(parts[1], '12345');
      expect(parts[2], '1771653694');
    });

    test('verification success', () {
      final qrData = QrSigner.generateQrData(partId: testPartId, createdAt: testCreatedAt);
      final parts = qrData.split(':');
      final id = int.parse(parts[1]);
      final ts = int.parse(parts[2]);
      final hmac = parts[3];

      final isValid = QrSigner.verify(id, ts, hmac);
      expect(isValid, true);
    });

    test('altered signature rejected', () {
      final qrData = QrSigner.generateQrData(partId: testPartId, createdAt: testCreatedAt);
      final parts = qrData.split(':');
      final id = int.parse(parts[1]);
      final ts = int.parse(parts[2]);
      final hmac = parts[3];

      final alteredHmac = hmac.replaceFirst(hmac[0], hmac[0] == 'a' ? 'b' : 'a');
      final isValidAltered = QrSigner.verify(id, ts, alteredHmac);
      expect(isValidAltered, false);
    });

    test('altered parameters rejected', () {
      final qrData = QrSigner.generateQrData(partId: testPartId, createdAt: testCreatedAt);
      final parts = qrData.split(':');
      final id = int.parse(parts[1]);
      final ts = int.parse(parts[2]);
      final hmac = parts[3];

      expect(QrSigner.verify(id + 1, ts, hmac), false);
      expect(QrSigner.verify(id, ts - 1, hmac), false);
    });

    test('determinism check', () {
      final qrData1 = QrSigner.generateQrData(partId: testPartId, createdAt: testCreatedAt);
      final qrData2 = QrSigner.generateQrData(partId: testPartId, createdAt: testCreatedAt);
      expect(qrData1, qrData2);
    });
  });
}
