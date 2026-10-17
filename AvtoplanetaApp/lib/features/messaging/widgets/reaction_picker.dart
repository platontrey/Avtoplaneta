import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import '../../../app/theme.dart';

class ReactionPickerSheet extends StatelessWidget {
  final bool isMine;
  final Function(String emoji) onSelectEmoji;
  final VoidCallback? onDelete;

  static const List<String> availableEmojis = ['👍', '❤️', '😂', '😮', '😢', '😡'];

  const ReactionPickerSheet({
    super.key,
    required this.isMine,
    required this.onSelectEmoji,
    this.onDelete,
  });

  static Future<void> show(
    BuildContext context, {
    required bool isMine,
    required Function(String emoji) onSelectEmoji,
    VoidCallback? onDelete,
  }) {
    return showModalBottomSheet(
      context: context,
      backgroundColor: AppTheme.surfaceColor,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => ReactionPickerSheet(
        isMine: isMine,
        onSelectEmoji: onSelectEmoji,
        onDelete: onDelete,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 20, vertical: 4),
              child: Text(
                'Реакция',
                style: TextStyle(
                  color: AppTheme.mutedColor,
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                children: availableEmojis.map((emoji) {
                  return InkWell(
                    onTap: () {
                      Navigator.pop(context);
                      onSelectEmoji(emoji);
                    },
                    borderRadius: BorderRadius.circular(16),
                    child: Container(
                      padding: const EdgeInsets.all(8),
                      child: Text(
                        emoji,
                        style: const TextStyle(fontSize: 28),
                      ),
                    ),
                  );
                }).toList(),
              ),
            ),
            if (isMine && onDelete != null) ...[
              const Divider(color: AppTheme.borderColor, height: 24),
              ListTile(
                leading: const Icon(LucideIcons.trash, color: AppTheme.dangerColor),
                title: const Text(
                  'Удалить сообщение',
                  style: TextStyle(
                    color: AppTheme.dangerColor,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                onTap: () {
                  Navigator.pop(context);
                  onDelete!();
                },
              ),
            ],
          ],
        ),
      ),
    );
  }
}
