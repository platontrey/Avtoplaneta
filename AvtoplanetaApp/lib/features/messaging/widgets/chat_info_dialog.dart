import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import '../../../app/theme.dart';
import '../models/messaging_models.dart';
import '../providers/messaging_providers.dart';

class ChatInfoDialog extends StatefulWidget {
  final Conversation conversation;
  final Map<int, ChatUser> users;
  final int? currentUserId;
  final bool isAdmin;
  final Function(Conversation updated) onUpdated;
  final VoidCallback onDeleted;

  const ChatInfoDialog({
    super.key,
    required this.conversation,
    required this.users,
    required this.currentUserId,
    required this.isAdmin,
    required this.onUpdated,
    required this.onDeleted,
  });

  static Future<void> show(
    BuildContext context, {
    required Conversation conversation,
    required Map<int, ChatUser> users,
    required int? currentUserId,
    required bool isAdmin,
    required Function(Conversation updated) onUpdated,
    required VoidCallback onDeleted,
  }) {
    return showDialog(
      context: context,
      builder: (_) => ChatInfoDialog(
        conversation: conversation,
        users: users,
        currentUserId: currentUserId,
        isAdmin: isAdmin,
        onUpdated: onUpdated,
        onDeleted: onDeleted,
      ),
    );
  }

  @override
  State<ChatInfoDialog> createState() => _ChatInfoDialogState();
}

class _ChatInfoDialogState extends State<ChatInfoDialog> {
  late TextEditingController _titleController;
  bool _isEditingTitle = false;
  bool _isSavingTitle = false;
  bool _isDeleting = false;
  int? _removingParticipantId;
  late Conversation _currentConversation;

  @override
  void initState() {
    super.initState();
    _currentConversation = widget.conversation;
    _titleController = TextEditingController(text: _currentConversation.title);
  }

  @override
  void dispose() {
    _titleController.dispose();
    super.dispose();
  }

  Future<void> _saveTitle() async {
    final newTitle = _titleController.text.trim();
    if (newTitle.isEmpty) return;

    setState(() => _isSavingTitle = true);
    try {
      final updated = await MessagingRepository.updateConversation(
        _currentConversation.id,
        newTitle,
      );
      if (mounted) {
        setState(() {
          _currentConversation = updated;
          _isEditingTitle = false;
        });
        widget.onUpdated(updated);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Название чата обновлено')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка обновления: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isSavingTitle = false);
    }
  }

  Future<void> _removeParticipant(int participantId) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить участника?'),
        content: const Text('Этот пользователь больше не сможет просматривать новые сообщения.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Отмена'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: AppTheme.dangerColor),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Удалить'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    setState(() => _removingParticipantId = participantId);
    try {
      await MessagingRepository.removeParticipant(_currentConversation.id, participantId);
      final newParticipants = _currentConversation.participants.where((id) => id != participantId).toList();
      final updated = _currentConversation.copyWith(participants: newParticipants);
      if (mounted) {
        setState(() {
          _currentConversation = updated;
        });
        widget.onUpdated(updated);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Участник удален из чата')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось удалить: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _removingParticipantId = null);
    }
  }

  Future<void> _deleteChat() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить этот чат?'),
        content: const Text('История переписки будет удалена без возможности восстановления.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Отмена'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: AppTheme.dangerColor),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Удалить чат'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    setState(() => _isDeleting = true);
    try {
      await MessagingRepository.deleteConversation(_currentConversation.id);
      if (mounted) {
        Navigator.pop(context); // Close dialog
        widget.onDeleted();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка удаления: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isDeleting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      backgroundColor: AppTheme.surfaceColor,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(20),
        side: const BorderSide(color: AppTheme.borderColor),
      ),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 420),
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text(
                    'Настройки чата',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(LucideIcons.x, size: 20),
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              ),
              const SizedBox(height: 16),

              // Chat Title Section
              const Text(
                'Название чата',
                style: TextStyle(
                  color: AppTheme.mutedColor,
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 6),
              if (_isEditingTitle)
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: _titleController,
                        autofocus: true,
                        decoration: const InputDecoration(
                          hintText: 'Введите название',
                          contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    IconButton(
                      icon: _isSavingTitle
                          ? const SizedBox(
                              width: 16,
                              height: 16,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(LucideIcons.check, color: AppTheme.successColor),
                      onPressed: _isSavingTitle ? null : _saveTitle,
                    ),
                    IconButton(
                      icon: const Icon(LucideIcons.x, color: AppTheme.mutedColor),
                      onPressed: () {
                        setState(() {
                          _titleController.text = _currentConversation.title;
                          _isEditingTitle = false;
                        });
                      },
                    ),
                  ],
                )
              else
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: Text(
                        _currentConversation.title.isNotEmpty
                            ? _currentConversation.title
                            : 'Без названия',
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: Colors.white,
                        ),
                      ),
                    ),
                    IconButton(
                      icon: const Icon(LucideIcons.pen_line, size: 18, color: AppTheme.primaryColor),
                      onPressed: () => setState(() => _isEditingTitle = true),
                    ),
                  ],
                ),

              const SizedBox(height: 16),
              const Divider(color: AppTheme.borderColor),
              const SizedBox(height: 12),

              // Participants Section
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Участники (${_currentConversation.participants.length})',
                    style: const TextStyle(
                      color: AppTheme.mutedColor,
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              ConstrainedBox(
                constraints: const BoxConstraints(maxHeight: 220),
                child: ListView.separated(
                  shrinkWrap: true,
                  itemCount: _currentConversation.participants.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 6),
                  itemBuilder: (_, index) {
                    final pId = _currentConversation.participants[index];
                    final user = widget.users[pId];
                    final isCurrentUser = widget.currentUserId == pId;
                    final canRemove = (widget.isAdmin || isCurrentUser) &&
                        _currentConversation.participants.length > 1;

                    final userName = user?.name ?? 'Пользователь №$pId';
                    final role = user?.role ?? 'operator';
                    final initials = user?.initials ?? (userName.isNotEmpty ? userName[0].toUpperCase() : '?');

                    return Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                      decoration: BoxDecoration(
                        color: AppTheme.cardColor,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: AppTheme.borderColor),
                      ),
                      child: Row(
                        children: [
                          CircleAvatar(
                            radius: 16,
                            backgroundColor: AppTheme.primaryColor.withValues(alpha: 0.18),
                            child: Text(
                              initials,
                              style: const TextStyle(
                                fontSize: 12,
                                fontWeight: FontWeight.bold,
                                color: AppTheme.primaryColor,
                              ),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  children: [
                                    Expanded(
                                      child: Text(
                                        userName,
                                        style: const TextStyle(
                                          fontSize: 13,
                                          fontWeight: FontWeight.w600,
                                          color: Colors.white,
                                        ),
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                    ),
                                    if (isCurrentUser)
                                      const Text(
                                        ' (вы)',
                                        style: TextStyle(
                                          fontSize: 11,
                                          color: AppTheme.primaryColor,
                                          fontWeight: FontWeight.w700,
                                        ),
                                      ),
                                  ],
                                ),
                                Text(
                                  role,
                                  style: const TextStyle(
                                    fontSize: 11,
                                    color: AppTheme.mutedColor,
                                  ),
                                ),
                              ],
                            ),
                          ),
                          if (canRemove && !isCurrentUser)
                            IconButton(
                              icon: _removingParticipantId == pId
                                  ? const SizedBox(
                                      width: 14,
                                      height: 14,
                                      child: CircularProgressIndicator(strokeWidth: 2),
                                    )
                                  : const Icon(LucideIcons.user_minus, size: 16, color: AppTheme.dangerColor),
                              onPressed: _removingParticipantId == null ? () => _removeParticipant(pId) : null,
                            ),
                        ],
                      ),
                    );
                  },
                ),
              ),

              const SizedBox(height: 20),
              const Divider(color: AppTheme.borderColor),
              const SizedBox(height: 12),

              // Delete Chat Action
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  style: OutlinedButton.styleFrom(
                    foregroundColor: AppTheme.dangerColor,
                    side: const BorderSide(color: AppTheme.dangerColor),
                    padding: const EdgeInsets.symmetric(vertical: 12),
                  ),
                  icon: _isDeleting
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2, color: AppTheme.dangerColor),
                        )
                      : const Icon(LucideIcons.trash, size: 18),
                  label: const Text('Удалить чат'),
                  onPressed: _isDeleting ? null : _deleteChat,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
