import '../lib/core/utils/qr_signer.dart';

void main() {
  print('Running QrSigner self-contained unit tests...');
  
  final testPartId = 12345;
  final testCreatedAt = DateTime.fromMillisecondsSinceEpoch(1771653694000);

  // Test 1: format
  final qrData = QrSigner.generateQrData(partId: testPartId, createdAt: testCreatedAt);
  if (!qrData.startsWith('ap:')) {
    throw Exception('Test 1 failed: does not start with ap:');
  }
  final parts = qrData.split(':');
  if (parts.length != 4) {
    throw Exception('Test 1 failed: split length is not 4');
  }
  if (parts[1] != '12345') {
    throw Exception('Test 1 failed: wrong partId');
  }
  if (parts[2] != '1771653694') {
    throw Exception('Test 1 failed: wrong timestamp');
  }
  print('✓ Test 1 passed: correct format');

  // Test 2: verification success
  final id = int.parse(parts[1]);
  final ts = int.parse(parts[2]);
  final hmac = parts[3];
  final isValid = QrSigner.verify(id, ts, hmac);
  if (!isValid) {
    throw Exception('Test 2 failed: valid signature not verified');
  }
  print('✓ Test 2 passed: verification success');

  // Test 3: altered signature
  final alteredHmac = hmac.replaceFirst(hmac[0], hmac[0] == 'a' ? 'b' : 'a');
  final isValidAltered = QrSigner.verify(id, ts, alteredHmac);
  if (isValidAltered) {
    throw Exception('Test 3 failed: altered signature verified');
  }
  print('✓ Test 3 passed: altered signature rejected');

  // Test 4: altered parameters
  if (QrSigner.verify(id + 1, ts, hmac)) {
    throw Exception('Test 4 failed: altered id verified');
  }
  if (QrSigner.verify(id, ts - 1, hmac)) {
    throw Exception('Test 4 failed: altered timestamp verified');
  }
  print('✓ Test 4 passed: altered parameters rejected');

  // Test 5: determinism
  final qrData2 = QrSigner.generateQrData(partId: testPartId, createdAt: testCreatedAt);
  if (qrData != qrData2) {
    throw Exception('Test 5 failed: non-deterministic output');
  }
  print('✓ Test 5 passed: deterministic output');

  print('\nAll QrSigner unit tests passed successfully!');
}
