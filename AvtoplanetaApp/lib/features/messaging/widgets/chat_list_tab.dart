import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../app/theme.dart';
import '../../../shared/widgets/app_states.dart';
import '../models/messaging_models.dart';
import '../providers/messaging_providers.dart';
import 'chat_info_dialog.dart';

class ChatListTab extends ConsumerWidget {
  final List<Conversation> conversations;
  final Map<int, ChatUser> users;
  final Map<int, UserStatus> userStatuses;
  final int? currentUserId;
  final bool isAdmin;
  final Function(Conversation conversation) onSelectConversation;
  final Future<void> Function() onRefresh;

  const ChatListTab({
    super.key,
    required this.conversations,
    required this.users,
    required this.userStatuses,
    required this.currentUserId,
    required this.isAdmin,
    required this.onSelectConversation,
    required this.onRefresh,
  });

  String _formatRelativeTime(DateTime dateTime) {
    final now = DateTime.now();
    final diff = now.difference(dateTime);

    if (diff.inMinutes < 1) return 'только что';
    if (diff.inMinutes < 60) return '${diff.inMinutes} мин';
    if (diff.inHours < 24) return '${diff.inHours} ч';
    if (diff.inDays < 7) return '${diff.inDays} д';
    return '${dateTime.day.toString().padLeft(2, '0')}.${dateTime.month.toString().padLeft(2, '0')}.${dateTime.year}';
  }

  String _getChatTitle(Conversation conv) {
    if (conv.title.isNotEmpty) return conv.title;

    final otherParticipantNames = conv.participants
        .where((id) => id != currentUserId)
        .map((id) => users[id]?.name)
        .whereType<String>()
        .where((name) => name.isNotEmpty)
        .toList();

    if (conv.participants.length == 2) {
      return otherParticipantNames.isNotEmpty
          ? otherParticipantNames.join(', ')
          : 'Чат #${conv.id}';
    } else if (conv.participants.length > 2) {
      final first = otherParticipantNames.isNotEmpty ? otherParticipantNames.first : null;
      return first != null ? '$first: Групповой чат' : 'Групповой чат';
    }
    return 'Чат #${conv.id}';
  }

  int _getOnlineCount(List<int> participantIds) {
    return participantIds.where((id) {
      final status = userStatuses[id];
      return status?.isOnline == true;
    }).length;
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (conversations.isEmpty) {
      return AppEmptyState(
        icon: LucideIcons.message_square_dashed,
        title: 'Нет активных чатов',
        message: 'Создайте новый диалог для общения с коллегами.',
        actionLabel: 'Обновить',
        onAction: onRefresh,
      );
    }

    return RefreshIndicator(
      onRefresh: onRefresh,
      color: AppTheme.primaryColor,
      backgroundColor: AppTheme.cardColor,
      child: ListView.separated(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        itemCount: conversations.length,
        separatorBuilder: (_, _) => const SizedBox(height: 8),
        itemBuilder: (context, index) {
          final conv = conversations[index];
          final title = _getChatTitle(conv);
          final onlineCount = _getOnlineCount(conv.participants);
          final timeStr = _formatRelativeTime(conv.lastMessageAt);
          final initials = title.isNotEmpty ? title[0].toUpperCase() : '?';

          return Material(
            color: AppTheme.cardColor,
            borderRadius: BorderRadius.circular(16),
            child: InkWell(
              onTap: () => onSelectConversation(conv),
              borderRadius: BorderRadius.circular(16),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: AppTheme.borderColor),
                ),
                child: Row(
                  children: [
                    // Avatar
                    CircleAvatar(
                      radius: 22,
                      backgroundColor: AppTheme.primaryColor.withValues(alpha: 0.16),
                      child: Text(
                        initials,
                        style: const TextStyle(
                          color: AppTheme.primaryColor,
                          fontWeight: FontWeight.w800,
                          fontSize: 16,
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),

                    // Title & Last message
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Expanded(
                                child: Text(
                                  title,
                                  style: const TextStyle(
                                    fontWeight: FontWeight.w700,
                                    fontSize: 14,
                                    color: Colors.white,
                                  ),
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                              const SizedBox(width: 8),
                              Text(
                                timeStr,
                                style: const TextStyle(
                                  fontSize: 11,
                                  color: AppTheme.mutedColor,
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 3),
                          Text(
                            conv.lastMessage ?? 'Нет сообщений',
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              fontSize: 12,
                              color: AppTheme.mutedColor,
                            ),
                          ),
                          const SizedBox(height: 6),
                          Row(
                            children: [
                              Text(
                                '${conv.participants.length} уч.',
                                style: const TextStyle(
                                  fontSize: 11,
                                  color: AppTheme.mutedColor,
                                ),
                              ),
                              if (onlineCount > 0) ...[
                                const SizedBox(width: 8),
                                Container(
                                  width: 6,
                                  height: 6,
                                  decoration: const BoxDecoration(
                                    color: AppTheme.successColor,
                                    shape: BoxShape.circle,
                                  ),
                                ),
                                const SizedBox(width: 4),
                                Text(
                                  '$onlineCount онлайн',
                                  style: const TextStyle(
                                    fontSize: 11,
                                    color: AppTheme.successColor,
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                              ],
                            ],
                          ),
                        ],
                      ),
                    ),

                    const SizedBox(width: 8),

                    // Unread badge & Actions
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        PopupMenuButton<String>(
                          icon: const Icon(LucideIcons.ellipsis_vertical, size: 18, color: AppTheme.mutedColor),
                          color: AppTheme.surfaceColor,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                            side: const BorderSide(color: AppTheme.borderColor),
                          ),
                          onSelected: (val) {
                            if (val == 'settings') {
                              ChatInfoDialog.show(
                                context,
                                conversation: conv,
                                users: users,
                                currentUserId: currentUserId,
                                isAdmin: isAdmin,
                                onUpdated: (updated) {
                                  ref.invalidate(conversationsProvider);
                                },
                                onDeleted: () {
                                  ref.invalidate(conversationsProvider);
                                },
                              );
                            }
                          },
                          itemBuilder: (_) => [
                            const PopupMenuItem(
                              value: 'settings',
                              child: Row(
                                children: [
                                  Icon(LucideIcons.settings_2, size: 16),
                                  SizedBox(width: 8),
                                  Text('Настройки чата'),
                                ],
                              ),
                            ),
                          ],
                        ),
                        if (conv.unreadCount > 0)
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
                            decoration: BoxDecoration(
                              color: AppTheme.primaryColor,
                              borderRadius: BorderRadius.circular(10),
                            ),
                            child: Text(
                              conv.unreadCount > 99 ? '99+' : '${conv.unreadCount}',
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 10,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                      ],
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
