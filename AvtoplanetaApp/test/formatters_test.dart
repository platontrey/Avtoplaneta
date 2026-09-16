/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import 'package:flutter_test/flutter_test.dart';
import 'package:flutter/services.dart';
import 'package:avtoplaneta_app/core/utils/formatters.dart';

void main() {
  group('CarReleasePeriodFormatter', () {
    const formatter = CarReleasePeriodFormatter();

    test('leaves up to 4 digits unchanged without dash', () {
      final res = formatter.formatEditUpdate(
        const TextEditingValue(text: '200'),
        const TextEditingValue(text: '2001', selection: TextSelection.collapsed(offset: 4)),
      );
      expect(res.text, '2001');
      expect(res.selection.baseOffset, 4);
    });

    test('automatically adds dash when 5th digit is entered', () {
      final res = formatter.formatEditUpdate(
        const TextEditingValue(text: '2001'),
        const TextEditingValue(text: '20015', selection: TextSelection.collapsed(offset: 5)),
      );
      expect(res.text, '2001-5');
      expect(res.selection.baseOffset, 6);
    });

    test('supports full period up to 8 digits', () {
      final res = formatter.formatEditUpdate(
        const TextEditingValue(text: '2001-200'),
        const TextEditingValue(text: '2001-2007', selection: TextSelection.collapsed(offset: 9)),
      );
      expect(res.text, '2001-2007');
      expect(res.selection.baseOffset, 9);
    });

    test('caps at 8 digits', () {
      final res = formatter.formatEditUpdate(
        const TextEditingValue(text: '2001-2007'),
        const TextEditingValue(text: '2001-20078', selection: TextSelection.collapsed(offset: 10)),
      );
      expect(res.text, '2001-2007');
    });

    test('allows explicit trailing dash or slash after 4 digits', () {
      final dashRes = formatter.formatEditUpdate(
        const TextEditingValue(text: '2001'),
        const TextEditingValue(text: '2001-', selection: TextSelection.collapsed(offset: 5)),
      );
      expect(dashRes.text, '2001-');

      final slashRes = formatter.formatEditUpdate(
        const TextEditingValue(text: '2001'),
        const TextEditingValue(text: '2001/', selection: TextSelection.collapsed(offset: 5)),
      );
      expect(slashRes.text, '2001-');
    });

    test('deleting 5th digit removes dash cleanly', () {
      final res = formatter.formatEditUpdate(
        const TextEditingValue(text: '2001-5'),
        const TextEditingValue(text: '2001-', selection: TextSelection.collapsed(offset: 5)),
      );
      expect(res.text, '2001');
    });

    test('deleting dash returns 4 digits', () {
      final res = formatter.formatEditUpdate(
        const TextEditingValue(text: '2001-'),
        const TextEditingValue(text: '2001', selection: TextSelection.collapsed(offset: 4)),
      );
      expect(res.text, '2001');
      expect(res.selection.baseOffset, 4);
    });
  });
}
