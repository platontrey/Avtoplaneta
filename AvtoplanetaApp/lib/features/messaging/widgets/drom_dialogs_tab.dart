import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import '../../../app/theme.dart';
import '../models/messaging_models.dart';
import 'drom_auth_dialog.dart';

class DromDialogsTab extends StatelessWidget {
  final List<DromDialog> dialogs;
  final bool isLoading;
  final Future<void> Function() onRefresh;
  final Function(DromDialog dialog) onSelectDialog;

  const DromDialogsTab({
    super.key,
    required this.dialogs,
    required this.isLoading,
    required this.onRefresh,
    required this.onSelectDialog,
  });

  @override
  Widget build(BuildContext context) {
    if (isLoading && dialogs.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (dialogs.isEmpty) {
      return Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Container(
                width: 72,
                height: 72,
                decoration: BoxDecoration(
                  color: AppTheme.primaryColor.withValues(alpha: 0.12),
                  borderRadius: BorderRadius.circular(24),
                  border: Border.all(
                    color: AppTheme.primaryColor.withValues(alpha: 0.22),
                  ),
                ),
                child: const Icon(LucideIcons.key_round, size: 34, color: AppTheme.primaryColor),
              ),
              const SizedBox(height: 20),
              const Text(
                'Нет загруженных диалогов Drom',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white),
              ),
              const SizedBox(height: 8),
              const Text(
                'Проверьте авторизацию на Drom.ru или обновите список сообщений',
                textAlign: TextAlign.center,
                style: TextStyle(color: AppTheme.mutedColor, fontSize: 13, height: 1.4),
              ),
              const SizedBox(height: 24),
              Wrap(
                spacing: 12,
                runSpacing: 10,
                alignment: WrapAlignment.center,
                children: [
                  OutlinedButton.icon(
                    style: OutlinedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                    ),
                    icon: const Icon(LucideIcons.key_round, size: 16),
                    label: const Text('Аккаунт Drom'),
                    onPressed: () {
                      DromAuthDialog.show(context, onSuccess: onRefresh);
                    },
                  ),
                  FilledButton.icon(
                    style: FilledButton.styleFrom(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                    ),
                    icon: const Icon(LucideIcons.refresh_cw, size: 16),
                    label: const Text('Обновить'),
                    onPressed: onRefresh,
                  ),
                ],
              ),
            ],
          ),
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: onRefresh,
      color: AppTheme.primaryColor,
      backgroundColor: AppTheme.cardColor,
      child: ListView.separated(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        itemCount: dialogs.length,
        separatorBuilder: (_, _) => const SizedBox(height: 8),
        itemBuilder: (context, index) {
          final dialog = dialogs[index];

          return Material(
            color: AppTheme.cardColor,
            borderRadius: BorderRadius.circular(16),
            child: InkWell(
              onTap: () => onSelectDialog(dialog),
              borderRadius: BorderRadius.circular(16),
              child: Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(
                    color: dialog.isUnread
                        ? AppTheme.primaryColor.withValues(alpha: 0.5)
                        : AppTheme.borderColor,
                  ),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Part photo or placeholder
                    ClipRRect(
                      borderRadius: BorderRadius.circular(12),
                      child: Container(
                        width: 52,
                        height: 52,
                        color: AppTheme.surfaceColor,
                        child: dialog.partPhotoUrl != null && dialog.partPhotoUrl!.isNotEmpty
                            ? CachedNetworkImage(
                                imageUrl: dialog.partPhotoUrl!,
                                fit: BoxFit.cover,
                                placeholder: (_, _) => const Center(
                                  child: SizedBox(
                                    width: 16,
                                    height: 16,
                                    child: CircularProgressIndicator(strokeWidth: 2),
                                  ),
                                ),
                                errorWidget: (_, _, _) => const Icon(
                                  LucideIcons.package,
                                  size: 24,
                                  color: AppTheme.mutedColor,
                                ),
                              )
                            : const Icon(
                                LucideIcons.message_circle,
                                size: 24,
                                color: AppTheme.mutedColor,
                              ),
                      ),
                    ),
                    const SizedBox(width: 12),

                    // Dialog details
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Expanded(
                                child: Row(
                                  children: [
                                    Flexible(
                                      child: Text(
                                        dialog.interlocutor,
                                        style: const TextStyle(
                                          fontWeight: FontWeight.w700,
                                          fontSize: 14,
                                          color: Colors.white,
                                        ),
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                    ),
                                    if (dialog.userRating != null && dialog.userRating!.isNotEmpty) ...[
                                      const SizedBox(width: 6),
                                      Container(
                                        padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
                                        decoration: BoxDecoration(
                                          color: AppTheme.successColor.withValues(alpha: 0.18),
                                          borderRadius: BorderRadius.circular(6),
                                        ),
                                        child: Text(
                                          dialog.userRating!,
                                          style: const TextStyle(
                                            color: AppTheme.successColor,
                                            fontSize: 10,
                                            fontWeight: FontWeight.w700,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                              ),
                              Row(
                                children: [
                                  if (dialog.formattedTime.isNotEmpty)
                                    Text(
                                      dialog.formattedTime,
                                      style: const TextStyle(
                                        fontSize: 11,
                                        color: AppTheme.mutedColor,
                                      ),
                                    ),
                                  if (dialog.isUnread) ...[
                                    const SizedBox(width: 6),
                                    Container(
                                      width: 8,
                                      height: 8,
                                      decoration: const BoxDecoration(
                                        color: AppTheme.dangerColor,
                                        shape: BoxShape.circle,
                                      ),
                                    ),
                                  ],
                                ],
                              ),
                            ],
                          ),

                          // Part title & price
                          if (dialog.partTitle != null && dialog.partTitle!.isNotEmpty) ...[
                            const SizedBox(height: 4),
                            Row(
                              children: [
                                const Icon(LucideIcons.tag, size: 12, color: AppTheme.primaryColor),
                                const SizedBox(width: 4),
                                Expanded(
                                  child: Text(
                                    dialog.partTitle!,
                                    style: const TextStyle(
                                      fontSize: 12,
                                      fontWeight: FontWeight.w600,
                                      color: AppTheme.primaryColor,
                                    ),
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                ),
                                if (dialog.partPrice != null && dialog.partPrice!.isNotEmpty) ...[
                                  const SizedBox(width: 6),
                                  Text(
                                    dialog.partPrice!,
                                    style: const TextStyle(
                                      fontSize: 12,
                                      fontWeight: FontWeight.bold,
                                      color: Colors.white,
                                    ),
                                  ),
                                ],
                              ],
                            ),
                          ],

                          const SizedBox(height: 4),

                          // Last message
                          Text(
                            dialog.lastMessage ?? 'Нет сообщений',
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                              fontSize: 12,
                              color: dialog.isUnread ? Colors.white : AppTheme.mutedColor,
                              fontWeight: dialog.isUnread ? FontWeight.w600 : FontWeight.normal,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}
