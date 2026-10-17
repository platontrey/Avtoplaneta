import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../../../app/theme.dart';

class DateDivider extends StatelessWidget {
  final DateTime date;

  const DateDivider({super.key, required this.date});

  String _formatDate(DateTime d) {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final messageDay = DateTime(d.year, d.month, d.day);
    final yesterday = today.subtract(const Duration(days: 1));

    if (messageDay == today) {
      return 'Сегодня';
    } else if (messageDay == yesterday) {
      return 'Вчера';
    } else if (now.year == d.year) {
      try {
        return DateFormat('d MMMM', 'ru').format(d);
      } catch (_) {
        return DateFormat('d MMMM').format(d);
      }
    } else {
      try {
        return DateFormat('d MMMM yyyy', 'ru').format(d);
      } catch (_) {
        return DateFormat('dd.MM.yyyy').format(d);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 12),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        decoration: BoxDecoration(
          color: AppTheme.surfaceColor,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: AppTheme.borderColor.withValues(alpha: 0.5)),
        ),
        child: Text(
          _formatDate(date),
          style: const TextStyle(
            color: AppTheme.mutedColor,
            fontSize: 11,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }
}
