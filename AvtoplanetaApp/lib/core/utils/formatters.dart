/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import 'package:flutter/services.dart';

/// Форматирует период выпуска автомобиля (например, 2001-2007).
/// При вводе 4 цифр и последующей 5-й автоматически подставляет дефис.
class CarReleasePeriodFormatter extends TextInputFormatter {
  const CarReleasePeriodFormatter();

  @override
  TextEditingValue formatEditUpdate(
    TextEditingValue oldValue,
    TextEditingValue newValue,
  ) {
    final text = newValue.text;
    if (text.isEmpty) {
      return newValue;
    }

    final isDeleting = oldValue.text.length > text.length;

    // Handle backspace when deleting the dash or char after dash
    if (isDeleting && oldValue.text.endsWith('-') && text == oldValue.text.substring(0, oldValue.text.length - 1)) {
      final digits = text.replaceAll(RegExp(r'\D'), '');
      return TextEditingValue(
        text: digits,
        selection: TextSelection.collapsed(offset: digits.length),
      );
    }

    final digits = text.replaceAll(RegExp(r'\D'), '');
    final capped = digits.length > 8 ? digits.substring(0, 8) : digits;

    String formatted;
    if (capped.length > 4) {
      formatted = '${capped.substring(0, 4)}-${capped.substring(4)}';
    } else if (!isDeleting && capped.length == 4 && (text.endsWith('-') || text.endsWith('/'))) {
      formatted = '$capped-';
    } else {
      formatted = capped;
    }

    int cursorPosition = formatted.length;
    final int digitsBeforeCursor = newValue.selection.end >= 0 && newValue.selection.end <= text.length
        ? text.substring(0, newValue.selection.end).replaceAll(RegExp(r'\D'), '').length
        : digits.length;

    if (digitsBeforeCursor <= 4) {
      final bool hadTrailingDash = newValue.selection.end >= 0 &&
          newValue.selection.end <= text.length &&
          text.substring(0, newValue.selection.end).endsWith('-');
      if (digitsBeforeCursor == 4 && (hadTrailingDash || formatted.length > 4)) {
        cursorPosition = hadTrailingDash ? 5 : 4;
      } else {
        cursorPosition = digitsBeforeCursor;
      }
    } else {
      cursorPosition = (digitsBeforeCursor + 1).clamp(0, formatted.length);
    }

    return TextEditingValue(
      text: formatted,
      selection: TextSelection.collapsed(offset: cursorPosition),
    );
  }
}
