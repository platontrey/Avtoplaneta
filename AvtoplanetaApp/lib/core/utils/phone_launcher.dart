/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

/// Утилита для очистки телефонного номера и открытия в приложении "Телефон".
class PhoneLauncher {
  const PhoneLauncher._();

  /// Очищает номер телефона для передачи в схему tel:.
  /// Убирает скобки, дефисы, пробелы: "+7 (999) 123-45-67" -> "+79991234567".
  static String cleanPhone(String raw) {
    final trimmed = raw.trim();
    if (trimmed.isEmpty) return '';

    final hasPlus = trimmed.startsWith('+');
    final digits = trimmed.replaceAll(RegExp(r'\D'), '');
    if (digits.isEmpty) return '';

    if (hasPlus) {
      return '+$digits';
    }

    // Если 11 цифр и начинается с 8 (Россия)
    if (digits.length == 11 && digits.startsWith('8')) {
      return '+7${digits.substring(1)}';
    }

    // Если 11 цифр и начинается с 7
    if (digits.length == 11 && digits.startsWith('7')) {
      return '+$digits';
    }

    // Если 10 цифр (без кода страны)
    if (digits.length == 10 && digits.startsWith('9')) {
      return '+7$digits';
    }

    return digits;
  }

  /// Открывает системное приложение "Телефон" с набранным номером (схема tel:).
  static Future<bool> makeCall(String rawPhone, [BuildContext? context]) async {
    final phone = cleanPhone(rawPhone);
    if (phone.isEmpty) {
      if (context != null && context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Номер телефона не указан')),
        );
      }
      return false;
    }

    final uri = Uri(scheme: 'tel', path: phone);
    try {
      final launched = await launchUrl(uri);
      if (!launched && context != null && context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось открыть телефон для $phone')),
        );
      }
      return launched;
    } catch (e) {
      if (context != null && context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка при вызове: $e')),
        );
      }
      return false;
    }
  }
}
