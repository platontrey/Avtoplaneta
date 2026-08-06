import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../app/theme.dart';
import '../../../core/api/api_client.dart';
import '../../../shared/widgets/app_states.dart';
import '../../auth/providers/auth_provider.dart';

final conversationsProvider = FutureProvider<List<dynamic>>((ref) async {
  final response = await apiClient.dio.get('/api/messaging/conversations');
  final data = response.data;
  if (data is List) {
    return data;
  }
  if (data is Map && data['conversations'] != null) {
    return data['conversations'] as List<dynamic>;
  }
  return [];
});

final messagingUsersProvider = FutureProvider<Map<int, String>>((ref) async {
  final response = await apiClient.dio.get('/api/messaging/users');
  final data = response.data;
  final users = data is Map && data['users'] is List
      ? data['users'] as List
      : const [];
  return {
    for (final user in users.whereType<Map>())
      if (user['id'] is num)
        (user['id'] as num).toInt():
            (user['name'] ??
                    user['username'] ??
                    user['email'] ??
                    'Пользователь')
                .toString(),
  };
});

class MessagingScreen extends ConsumerWidget {
  const MessagingScreen({super.key});

  Future<void> _refresh(WidgetRef ref) async {
    ref.invalidate(conversationsProvider);
    ref.invalidate(messagingUsersProvider);
    await Future.wait([
      ref.read(conversationsProvider.future),
      ref.read(messagingUsersProvider.future),
    ]);
  }

  Future<void> _createConversation(
    BuildContext context,
    WidgetRef ref,
    Map<int, String> userNames,
    int? currentUserId,
  ) async {
    final availableUsers =
        userNames.entries.where((entry) => entry.key != currentUserId).toList()
          ..sort((a, b) => a.value.compareTo(b.value));
    if (availableUsers.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Нет доступных пользователей для диалога'),
        ),
      );
      return;
    }

    final titleController = TextEditingController();
    final selected = <int>{};
    final created = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('Новый диалог'),
          content: SizedBox(
            width: double.maxFinite,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: titleController,
                  decoration: const InputDecoration(
                    labelText: 'Название (необязательно)',
                  ),
                ),
                const SizedBox(height: 12),
                SizedBox(
                  height: 300,
                  child: ListView(
                    children: availableUsers
                        .map(
                          (user) => CheckboxListTile(
                            value: selected.contains(user.key),
                            title: Text(user.value),
                            dense: true,
                            controlAffinity: ListTileControlAffinity.leading,
                            onChanged: (checked) => setDialogState(() {
                              if (checked == true) {
                                selected.add(user.key);
                              } else {
                                selected.remove(user.key);
                              }
                            }),
                          ),
                        )
                        .toList(),
                  ),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(dialogContext, false),
              child: const Text('Отмена'),
            ),
            FilledButton(
              onPressed: selected.isEmpty
                  ? null
                  : () => Navigator.pop(dialogContext, true),
              child: const Text('Создать'),
            ),
          ],
        ),
      ),
    );
    final conversationTitle = titleController.text.trim();
    titleController.dispose();
    if (created != true || !context.mounted) {
      return;
    }

    try {
      await apiClient.dio.post(
        '/api/messaging/conversations',
        data: {'participants': selected.toList(), 'title': conversationTitle},
      );
      if (!context.mounted) {
        return;
      }
      ref.invalidate(conversationsProvider);
    } catch (error) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось создать диалог: $error')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final convsAsync = ref.watch(conversationsProvider);
    final userNames = ref.watch(messagingUsersProvider).valueOrNull ?? const {};
    final currentUserId = ref.watch(authProvider).valueOrNull?.id;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Сообщения'),
        actions: [
          IconButton(
            tooltip: 'Новый диалог',
            icon: const Icon(Icons.edit_square),
            onPressed: () =>
                _createConversation(context, ref, userNames, currentUserId),
          ),
          IconButton(
            tooltip: 'Обновить',
            icon: const Icon(Icons.refresh),
            onPressed: () => _refresh(ref),
          ),
        ],
      ),
      body: convsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => AppEmptyState(
          icon: Icons.cloud_off_rounded,
          title: 'Не удалось загрузить сообщения',
          message: 'Проверьте соединение и повторите попытку.',
          actionLabel: 'Повторить',
          onAction: () => ref.invalidate(conversationsProvider),
        ),
        data: (convs) => convs.isEmpty
            ? const AppEmptyState(
                icon: Icons.forum_outlined,
                title: 'Диалогов пока нет',
                message: 'Здесь появится переписка с вашей командой.',
              )
            : ListView.separated(
                padding: const EdgeInsets.fromLTRB(16, 10, 16, 20),
                itemCount: convs.length,
                separatorBuilder: (_, _) => const SizedBox(height: 10),
                itemBuilder: (_, i) {
                  final conv = Map<String, dynamic>.from(convs[i] as Map);
                  final title = _conversationTitle(
                    conv,
                    userNames,
                    currentUserId,
                  );
                  final unreadCount =
                      (conv['unread_count'] as num?)?.toInt() ?? 0;
                  return ListTile(
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 14,
                      vertical: 7,
                    ),
                    tileColor: AppTheme.cardColor,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(18),
                      side: const BorderSide(color: AppTheme.borderColor),
                    ),
                    leading: CircleAvatar(
                      radius: 23,
                      backgroundColor: AppTheme.primaryColor.withValues(
                        alpha: 0.16,
                      ),
                      child: Text(
                        title.isEmpty ? '?' : title[0].toUpperCase(),
                        style: const TextStyle(
                          color: AppTheme.primaryColor,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ),
                    title: Text(
                      title,
                      style: const TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    subtitle: Row(
                      children: [
                        Expanded(
                          child: Text(
                            conv['last_message']?.toString() ?? '',
                            style: const TextStyle(
                              color: AppTheme.mutedColor,
                              fontSize: 12,
                            ),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        if (unreadCount > 0)
                          Container(
                            margin: const EdgeInsets.only(left: 8),
                            padding: const EdgeInsets.symmetric(
                              horizontal: 7,
                              vertical: 2,
                            ),
                            decoration: const BoxDecoration(
                              color: AppTheme.primaryColor,
                              shape: BoxShape.circle,
                            ),
                            child: Text(
                              unreadCount > 99 ? '99+' : '$unreadCount',
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 10,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                      ],
                    ),
                    onTap: () async {
                      await _openChat(context, conv, userNames, currentUserId);
                      ref.invalidate(conversationsProvider);
                    },
                    trailing: const Icon(Icons.chevron_right_rounded, size: 20),
                  );
                },
              ),
      ),
    );
  }

  String _conversationTitle(
    Map<String, dynamic> conv,
    Map<int, String> userNames,
    int? currentUserId,
  ) {
    final title = conv['title']?.toString().trim() ?? '';
    if (title.isNotEmpty) {
      return title;
    }

    final participants = conv['participants'] is List
        ? conv['participants'] as List
        : const [];
    final names = participants
        .whereType<num>()
        .map((id) => id.toInt())
        .where((id) => id != currentUserId)
        .map((id) => userNames[id])
        .whereType<String>()
        .where((name) => name.isNotEmpty)
        .toList();
    if (names.isNotEmpty) {
      return names.join(', ');
    }
    return 'Чат №${conv['id']}';
  }

  Future<void> _openChat(
    BuildContext context,
    Map<String, dynamic> conv,
    Map<int, String> userNames,
    int? currentUserId,
  ) {
    return Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => _ChatScreen(
          convId: (conv['id'] as num).toInt(),
          title: _conversationTitle(conv, userNames, currentUserId),
          userNames: userNames,
        ),
      ),
    );
  }
}

class _ChatScreen extends ConsumerStatefulWidget {
  final int convId;
  final String title;
  final Map<int, String> userNames;
  const _ChatScreen({
    required this.convId,
    required this.title,
    required this.userNames,
  });

  @override
  ConsumerState<_ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<_ChatScreen> {
  final _msgCtrl = TextEditingController();
  List<Map<String, dynamic>> _messages = [];
  bool _loading = true;
  bool _fetchingMessages = false;
  bool _sending = false;
  String? _loadError;
  Timer? _refreshTimer;

  @override
  void initState() {
    super.initState();
    _loadMessages();
    // Сайт работает через тот же HTTP API. Периодическое обновление не зависит
    // от недоступного извне внутреннего порта messaging-service.
    _refreshTimer = Timer.periodic(
      const Duration(seconds: 5),
      (_) => _loadMessages(silent: true),
    );
  }

  @override
  void dispose() {
    _refreshTimer?.cancel();
    _msgCtrl.dispose();
    super.dispose();
  }

  Future<void> _loadMessages({bool silent = false}) async {
    if (_fetchingMessages) {
      return;
    }
    _fetchingMessages = true;
    try {
      final response = await apiClient.dio.get(
        '/api/messaging/conversations/${widget.convId}/messages',
      );
      final data = response.data;
      if (!mounted) {
        return;
      }
      final rawMessages = data is Map && data['messages'] is List
          ? data['messages'] as List
          : const [];
      final messages = rawMessages
          .whereType<Map>()
          .map((message) => Map<String, dynamic>.from(message))
          .toList();
      setState(() {
        _messages = messages;
        _loading = false;
        _loadError = null;
      });
      unawaited(_markIncomingMessagesRead(messages).catchError((_) {}));
    } catch (error) {
      if (mounted && !silent) {
        setState(() {
          _loading = false;
          _loadError = error.toString();
        });
      }
    } finally {
      _fetchingMessages = false;
    }
  }

  Future<void> _markIncomingMessagesRead(
    List<Map<String, dynamic>> messages,
  ) async {
    final currentUserId = ref.read(authProvider).valueOrNull?.id;
    if (currentUserId == null) {
      return;
    }
    final unread = messages.where((message) {
      final senderId = (message['sender_id'] as num?)?.toInt();
      final readBy = message['read_by'] is List
          ? message['read_by'] as List
          : const [];
      return senderId != currentUserId &&
          !readBy.whereType<num>().any((id) => id.toInt() == currentUserId);
    });
    await Future.wait(
      unread.map((message) async {
        final messageId = (message['id'] as num?)?.toInt();
        if (messageId != null) {
          await apiClient.dio.put('/api/messaging/messages/$messageId/read');
        }
      }),
    );
  }

  Future<void> _send() async {
    final text = _msgCtrl.text.trim();
    if (text.isEmpty || _sending) {
      return;
    }
    setState(() => _sending = true);
    try {
      await apiClient.dio.post(
        '/api/messaging/conversations/${widget.convId}/messages',
        data: {'content': text},
      );
      if (!mounted) {
        return;
      }
      _msgCtrl.clear();
      await _loadMessages();
      if (mounted) {
        ref.invalidate(conversationsProvider);
      }
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось отправить сообщение: $error')),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _sending = false);
      }
    }
  }

  Future<void> _showMessageActions(
    Map<String, dynamic> message,
    bool isMine,
  ) async {
    final messageId = (message['id'] as num?)?.toInt();
    if (messageId == null) {
      return;
    }
    final action = await showModalBottomSheet<String>(
      context: context,
      builder: (sheetContext) => SafeArea(
        child: Wrap(
          children: [
            const Padding(
              padding: EdgeInsets.fromLTRB(16, 12, 16, 4),
              child: Text('Реакция'),
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: ['👍', '❤️', '😂', '😮', '😢']
                  .map(
                    (emoji) => IconButton(
                      onPressed: () => Navigator.pop(sheetContext, emoji),
                      icon: Text(emoji, style: const TextStyle(fontSize: 24)),
                    ),
                  )
                  .toList(),
            ),
            if (isMine)
              ListTile(
                leading: const Icon(Icons.delete_outline, color: Colors.red),
                title: const Text('Удалить сообщение'),
                onTap: () => Navigator.pop(sheetContext, 'delete'),
              ),
          ],
        ),
      ),
    );
    if (action == null || !mounted) {
      return;
    }
    try {
      if (action == 'delete') {
        await apiClient.dio.delete('/api/messaging/messages/$messageId');
      } else {
        await apiClient.dio.post(
          '/api/messaging/messages/$messageId/reactions',
          data: {'emoji': action},
        );
      }
      await _loadMessages();
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось выполнить действие: $error')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final currentUserId = ref.watch(authProvider).valueOrNull?.id;

    return Scaffold(
      appBar: AppBar(title: Text(widget.title)),
      body: Column(
        children: [
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : _loadError != null
                ? AppEmptyState(
                    icon: Icons.cloud_off_rounded,
                    title: 'Не удалось загрузить сообщения',
                    message: _loadError!,
                    actionLabel: 'Повторить',
                    onAction: _loadMessages,
                  )
                : _messages.isEmpty
                ? const AppEmptyState(
                    icon: Icons.waving_hand_outlined,
                    title: 'Начните разговор',
                    message: 'Напишите первое сообщение в этом диалоге.',
                  )
                : ListView.builder(
                    reverse: true,
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 20),
                    itemCount: _messages.length,
                    itemBuilder: (_, i) {
                      final msg = _messages[_messages.length - 1 - i];
                      final isMine =
                          (msg['sender_id'] as num?)?.toInt() == currentUserId;
                      return _MessageBubble(
                        msg: msg,
                        userNames: widget.userNames,
                        isMine: isMine,
                        onLongPress: () => _showMessageActions(msg, isMine),
                      );
                    },
                  ),
          ),
          // Поле ввода
          Container(
            decoration: const BoxDecoration(
              color: AppTheme.surfaceColor,
              border: Border(top: BorderSide(color: AppTheme.borderColor)),
            ),
            padding: const EdgeInsets.fromLTRB(12, 10, 10, 10),
            child: SafeArea(
              top: false,
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _msgCtrl,
                      minLines: 1,
                      maxLines: 4,
                      textCapitalization: TextCapitalization.sentences,
                      decoration: const InputDecoration(
                        hintText: 'Написать сообщение…',
                        contentPadding: EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 12,
                        ),
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
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          )
                        : const Icon(Icons.arrow_upward_rounded),
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

class _MessageBubble extends StatelessWidget {
  final Map<String, dynamic> msg;
  final Map<int, String> userNames;
  final bool isMine;
  final VoidCallback onLongPress;
  const _MessageBubble({
    required this.msg,
    required this.userNames,
    required this.isMine,
    required this.onLongPress,
  });

  @override
  Widget build(BuildContext context) {
    final content = msg['content'] as String? ?? '';
    final senderId = (msg['sender_id'] as num?)?.toInt();
    final senderName = senderId == null
        ? 'Неизвестно'
        : userNames[senderId] ?? 'Пользователь №$senderId';

    final reactions = msg['reactions'] is List
        ? msg['reactions'] as List
        : const [];
    final createdAt = DateTime.tryParse(msg['created_at']?.toString() ?? '');
    final time = createdAt == null
        ? ''
        : '${createdAt.toLocal().hour.toString().padLeft(2, '0')}:${createdAt.toLocal().minute.toString().padLeft(2, '0')}';

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: isMine
            ? CrossAxisAlignment.end
            : CrossAxisAlignment.start,
        children: [
          if (!isMine)
            Padding(
              padding: const EdgeInsets.only(left: 4),
              child: Text(
                senderName,
                style: const TextStyle(
                  color: AppTheme.mutedColor,
                  fontSize: 11,
                ),
              ),
            ),
          const SizedBox(height: 2),
          GestureDetector(
            onLongPress: onLongPress,
            child: FractionallySizedBox(
              widthFactor: 0.82,
              alignment: isMine ? Alignment.centerRight : Alignment.centerLeft,
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  color: isMine ? AppTheme.primaryColor : AppTheme.cardColor,
                  borderRadius: BorderRadius.only(
                    topLeft: const Radius.circular(18),
                    topRight: const Radius.circular(18),
                    bottomLeft: Radius.circular(isMine ? 18 : 5),
                    bottomRight: Radius.circular(isMine ? 5 : 18),
                  ),
                  border: isMine
                      ? null
                      : Border.all(color: AppTheme.borderColor),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Align(
                      alignment: Alignment.centerLeft,
                      child: Text(
                        content,
                        style: const TextStyle(
                          color: Colors.white,
                          height: 1.35,
                        ),
                      ),
                    ),
                    if (reactions.isNotEmpty) ...[
                      const SizedBox(height: 5),
                      Align(
                        alignment: Alignment.centerLeft,
                        child: Wrap(
                          spacing: 4,
                          children: reactions
                              .whereType<Map>()
                              .map(
                                (reaction) =>
                                    Text(reaction['emoji']?.toString() ?? ''),
                              )
                              .toList(),
                        ),
                      ),
                    ],
                    if (time.isNotEmpty)
                      Text(
                        time,
                        style: const TextStyle(
                          color: Colors.white54,
                          fontSize: 9,
                        ),
                      ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
