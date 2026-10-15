import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import '../../../core/models/customer.dart';

extension CustomerUI on Customer {
  Color get categoryColor {
    switch (category) {
      case 'vip':
        return const Color(0xFFD97706);
      case 'wholesale':
        return const Color(0xFF2563EB);
      case 'blacklist':
        return const Color(0xFFDC2626);
      case 'regular':
      default:
        return const Color(0xFF71717A);
    }
  }

  IconData get categoryIcon {
    switch (category) {
      case 'vip':
        return LucideIcons.crown;
      case 'wholesale':
        return LucideIcons.building;
      case 'blacklist':
        return LucideIcons.triangle_alert;
      case 'regular':
      default:
        return LucideIcons.user;
    }
  }
}
