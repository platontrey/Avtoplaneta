import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../app/theme.dart';
import '../../../shared/widgets/app_states.dart';
import '../../auth/providers/auth_provider.dart';
import '../models/messaging_models.dart';
import '../providers/messaging_providers.dart';
import '../widgets/chat_info_dialog.dart';
import '../widgets/date_divider.dart';
import '../widgets/message_bubble.dart';
import '../widgets/reaction_picker.dart';

class ChatScreen extends ConsumerStatefulWidget {
  final int convId;
  final String title;
  final List<int> initialParticipants;

  const ChatScreen({
    super.key,
    required this.convId,
    required this.title,
    this.initialParticipants = const [],
  });

  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  final _msgController = TextEditingController();
  final _scrollController = ScrollController();
  List<Message> _messages = [];
  bool _loading = true;
  bool _sending = false;
  String? _error;
  Timer? _pollingTimer;
  late String _chatTitle;
  late List<int> _participants;

  @override
  void initState() {
    super.initState();
    _chatTitle = widget.title;
    _participants = List<int>.from(widget.initialParticipants);
    _loadMessages();

    // 5-second polling interval matching web app
    _pollingTimer = Timer.periodic(const Duration(seconds: 5), (_) {
      _loadMessages(silent: true);
    });
  }

  @override
  void dispose() {
    _pollingTimer?.cancel();
    _msgController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  Future<void> _loadMessages({bool silent = false}) async {
    try {
      final messages = await MessagingRepository.getMessages(widget.convId);
      if (!mounted) return;

      setState(() {
        _messages = messages;
        _loading = false;
        _error = null;
      });

      _markUnreadAsRead(messages);
    } catch (e) {
      if (mounted && !silent) {
        setState(() {
          _loading = false;
          _error = e.toString();
        });
      }
    }
  }

  void _markUnreadAsRead(List<Message> messages) {
    final currentUserId = ref.read(authProvider).valueOrNull?.id;
    if (currentUserId == null) return;

    for (final msg in messages) {
      if (msg.senderId != currentUserId && !msg.readBy.contains(currentUserId)) {
        MessagingRepository.markMessageRead(msg.id).catchError((_) {});
      }
    }
  }

  void _scrollToBottom() {
    if (_scrollController.hasClients) {
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent + 80,
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeOut,
      );
    }
  }

  Future<void> _send() async {
    final text = _msgController.text.trim();
    if (text.isEmpty || _sending) return;

    setState(() => _sending = true);
    try {
      final msg = await MessagingRepository.sendMessage(widget.convId, text);
      if (mounted) {
        _msgController.clear();
        setState(() {
          _messages.add(msg);
        });
        _scrollToBottom();
        ref.invalidate(conversationsProvider);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось отправить: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _sending = false);
    }
  }

  Future<void> _addReaction(int messageId, String emoji) async {
    try {
      await MessagingRepository.addReaction(messageId, emoji);
      await _loadMessages(silent: true);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка реакции: $e')),
        );
      }
    }
  }

  Future<void> _removeReaction(int messageId, int reactionId) async {
    try {
      await MessagingRepository.removeReaction(messageId, reactionId);
      await _loadMessages(silent: true);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка удаления реакции: $e')),
        );
      }
    }
  }

  Future<void> _deleteMessage(int messageId) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить сообщение?'),
        content: const Text('Сообщение будет удалено для всех участников диалога.'),
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

    try {
      await MessagingRepository.deleteMessage(messageId);
      setState(() {
        _messages.removeWhere((m) => m.id == messageId);
      });
      ref.invalidate(conversationsProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось удалить: $e')),
        );
      }
    }
  }

  void _openChatInfo(Map<int, ChatUser> users, int? currentUserId, bool isAdmin) {
    ChatInfoDialog.show(
      context,
      conversation: Conversation(
        id: widget.convId,
        participants: _participants,
        title: _chatTitle,
        createdAt: DateTime.now(),
        lastMessageAt: DateTime.now(),
      ),
      users: users,
      currentUserId: currentUserId,
      isAdmin: isAdmin,
      onUpdated: (updated) {
        setState(() {
          _chatTitle = updated.title.isNotEmpty ? updated.title : _chatTitle;
          _participants = updated.participants;
        });
        ref.invalidate(conversationsProvider);
      },
      onDeleted: () {
        ref.invalidate(conversationsProvider);
        Navigator.pop(context);
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final currentUserId = ref.watch(authProvider).valueOrNull?.id;
    final currentUserRole = ref.watch(authProvider).valueOrNull?.role ?? 'operator';
    final isAdmin = currentUserRole == 'admin';
    final users = ref.watch(messagingUsersProvider).valueOrNull ?? const {};
    final userStatuses = ref.watch(userStatusesProvider).valueOrNull ?? const {};

    final onlineCount = _participants.where((id) {
      return userStatuses[id]?.isOnline == true;
    }).length;

    return Scaffold(
      appBar: AppBar(
        titleSpacing: 0,
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              _chatTitle,
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700),
              overflow: TextOverflow.ellipsis,
            ),
            Row(
              children: [
                Text(
                  '${_participants.length} уч.',
                  style: const TextStyle(fontSize: 11, color: AppTheme.mutedColor),
                ),
                if (onlineCount > 0) ...[
                  const SizedBox(width: 6),
                  Container(
                    width: 5,
                    height: 5,
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
        actions: [
          IconButton(
            tooltip: 'Настройки чата',
            icon: const Icon(LucideIcons.info),
            onPressed: () => _openChatInfo(users, currentUserId, isAdmin),
          ),
        ],
      ),
      body: Column(
        children: [
          // Message List
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : _error != null
                    ? AppEmptyState(
                        icon: LucideIcons.cloud_off,
                        title: 'Не удалось загрузить сообщения',
                        message: _error!,
                        actionLabel: 'Повторить',
                        onAction: () => _loadMessages(),
                      )
                    : _messages.isEmpty
                        ? const AppEmptyState(
                            icon: LucideIcons.hand,
                            title: 'Начните разговор',
                            message: 'Напишите первое сообщение в этом диалоге.',
                          )
                        : ListView.builder(
                            controller: _scrollController,
                            padding: const EdgeInsets.symmetric(vertical: 12),
                            itemCount: _messages.length,
                            itemBuilder: (context, index) {
                              final message = _messages[index];
                              final isMine = message.senderId == currentUserId;

                              final prevMessage = index > 0 ? _messages[index - 1] : null;
                              final showDateDivider = prevMessage == null ||
                                  prevMessage.createdAt.toLocal().day !=
                                      message.createdAt.toLocal().day ||
                                  prevMessage.createdAt.toLocal().month !=
                                      message.createdAt.toLocal().month ||
                                  prevMessage.createdAt.toLocal().year !=
                                      message.createdAt.toLocal().year;

                              return Column(
                                children: [
                                  if (showDateDivider)
                                    DateDivider(date: message.createdAt.toLocal()),
                                  MessageBubble(
                                    message: message,
                                    users: users,
                                    currentUserId: currentUserId,
                                    onLongPress: () {
                                      ReactionPickerSheet.show(
                                        context,
                                        isMine: isMine,
                                        onSelectEmoji: (emoji) => _addReaction(message.id, emoji),
                                        onDelete: isMine ? () => _deleteMessage(message.id) : null,
                                      );
                                    },
                                    onAddReaction: (emoji) => _addReaction(message.id, emoji),
                                    onRemoveReaction: (reactionId) =>
                                        _removeReaction(message.id, reactionId),
                                  ),
                                ],
                              );
                            },
                          ),
          ),

          // Message Input Field
          Container(
            padding: const EdgeInsets.fromLTRB(12, 8, 12, 12),
            decoration: const BoxDecoration(
              color: AppTheme.surfaceColor,
              border: Border(top: BorderSide(color: AppTheme.borderColor)),
            ),
            child: SafeArea(
              top: false,
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _msgController,
                      minLines: 1,
                      maxLines: 4,
                      textCapitalization: TextCapitalization.sentences,
                      decoration: const InputDecoration(
                        hintText: 'Введите сообщение…',
                        contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                      ),
                      onSubmitted: (_) => _send(),
                    ),
                  ),
                  const SizedBox(width: 8),
                  IconButton.filled(
                    tooltip: 'Отправить',
                    icon: _sending
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                          )
                        : const Icon(LucideIcons.arrow_up),
                    onPressed: _sending ? null : _send,
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
