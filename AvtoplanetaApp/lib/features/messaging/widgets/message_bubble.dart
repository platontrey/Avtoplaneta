import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../../../app/theme.dart';
import '../models/messaging_models.dart';

class MessageBubble extends StatelessWidget {
  final Message message;
  final Map<int, ChatUser> users;
  final int? currentUserId;
  final VoidCallback onLongPress;
  final Function(String emoji) onAddReaction;
  final Function(int reactionId) onRemoveReaction;

  const MessageBubble({
    super.key,
    required this.message,
    required this.users,
    required this.currentUserId,
    required this.onLongPress,
    required this.onAddReaction,
    required this.onRemoveReaction,
  });

  bool get isMine => currentUserId != null && message.senderId == currentUserId;

  @override
  Widget build(BuildContext context) {
    final sender = users[message.senderId];
    final senderName = sender?.name ?? 'Пользователь №${message.senderId}';
    final timeStr = DateFormat('HH:mm').format(message.createdAt.toLocal());

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4, horizontal: 16),
      child: Column(
        crossAxisAlignment: isMine ? CrossAxisAlignment.end : CrossAxisAlignment.start,
        children: [
          if (!isMine)
            Padding(
              padding: const EdgeInsets.only(left: 6, bottom: 4),
              child: Text(
                senderName,
                style: const TextStyle(
                  color: AppTheme.mutedColor,
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          GestureDetector(
            onLongPress: onLongPress,
            child: Container(
              constraints: BoxConstraints(
                maxWidth: MediaQuery.of(context).size.width * 0.78,
              ),
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                color: isMine ? AppTheme.primaryColor : AppTheme.cardColor,
                borderRadius: BorderRadius.only(
                  topLeft: const Radius.circular(18),
                  topRight: const Radius.circular(18),
                  bottomLeft: Radius.circular(isMine ? 18 : 4),
                  bottomRight: Radius.circular(isMine ? 4 : 18),
                ),
                border: isMine ? null : Border.all(color: AppTheme.borderColor),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    message.content,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 14,
                      height: 1.35,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Align(
                    alignment: Alignment.bottomRight,
                    child: Text(
                      timeStr,
                      style: TextStyle(
                        color: isMine ? Colors.white.withValues(alpha: 0.7) : AppTheme.mutedColor,
                        fontSize: 10,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),

          // Reactions
          if (message.reactions.isNotEmpty) ...[
            const SizedBox(height: 4),
            Wrap(
              spacing: 6,
              runSpacing: 4,
              alignment: isMine ? WrapAlignment.end : WrapAlignment.start,
              children: _buildReactionChips(),
            ),
          ],
        ],
      ),
    );
  }

  List<Widget> _buildReactionChips() {
    // Group reactions by emoji
    final grouped = <String, List<Reaction>>{};
    for (final reaction in message.reactions) {
      grouped.putIfAbsent(reaction.emoji, () => []).add(reaction);
    }

    return grouped.entries.map((entry) {
      final emoji = entry.key;
      final reactionsList = entry.value;
      final count = reactionsList.length;
      final myReaction = currentUserId != null
          ? reactionsList.where((r) => r.userId == currentUserId).firstOrNull
          : null;
      final isMineReaction = myReaction != null;

      return InkWell(
        onTap: () {
          if (isMineReaction) {
            onRemoveReaction(myReaction.id);
          } else {
            onAddReaction(emoji);
          }
        },
        borderRadius: BorderRadius.circular(14),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
          decoration: BoxDecoration(
            color: isMineReaction
                ? AppTheme.primaryColor.withValues(alpha: 0.22)
                : AppTheme.surfaceColor,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(
              color: isMineReaction
                  ? AppTheme.primaryColor.withValues(alpha: 0.6)
                  : AppTheme.borderColor,
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(emoji, style: const TextStyle(fontSize: 13)),
              const SizedBox(width: 4),
              Text(
                count > 1 ? '$count' : '',
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                  color: isMineReaction ? AppTheme.primaryColor : AppTheme.mutedColor,
                ),
              ),
              if (isMineReaction) ...[
                const SizedBox(width: 2),
                const Text(
                  'вы',
                  style: TextStyle(
                    fontSize: 9,
                    fontWeight: FontWeight.w700,
                    color: AppTheme.primaryColor,
                  ),
                ),
              ],
            ],
          ),
        ),
      );
    }).toList();
  }
}
