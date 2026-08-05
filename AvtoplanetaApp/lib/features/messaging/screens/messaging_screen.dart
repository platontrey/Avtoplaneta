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
            (user['name'] ?? user['username'] ?? user['email'] ?? 'Пользователь')
                .toString(),
  };
});

class MessagingScreen extends ConsumerWidget {
  const MessagingScreen({super.key});

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
            tooltip: 'Обновить',
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.invalidate(conversationsProvider),
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
                separatorBuilder: (_, __) => const SizedBox(height: 10),
                itemBuilder: (_, i) {
                  final conv = convs[i] as Map<String, dynamic>;
                  final title =
                      _conversationTitle(conv, userNames, currentUserId);
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
                      backgroundColor: AppTheme.primaryColor.withValues(alpha: 0.16),
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
                    subtitle: Text(
                      conv['last_message'] as String? ?? '',
                      style: const TextStyle(color: AppTheme.mutedColor, fontSize: 12),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    onTap: () => _openChat(
                      context,
                      conv,
                      userNames,
                      currentUserId,
                    ),
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
    if (title.isNotEmpty) return title;

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
    if (names.isNotEmpty) return names.join(', ');
    return 'Чат №${conv['id']}';
  }

  void _openChat(
    BuildContext context,
    Map<String, dynamic> conv,
    Map<int, String> userNames,
    int? currentUserId,
  ) {
    Navigator.push(
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
  List<dynamic> _messages = [];
  bool _loading = true;
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
    try {
      final response = await apiClient.dio
          .get('/api/messaging/conversations/${widget.convId}/messages');
      final data = response.data;
      if (!mounted) return;
      setState(() {
        _messages = (data is Map ? data['messages'] ?? [] : []) as List<dynamic>;
        _loading = false;
      });
    } catch (_) {
      if (mounted && !silent) setState(() => _loading = false);
    }
  }

  Future<void> _send() async {
    final text = _msgCtrl.text.trim();
    if (text.isEmpty) return;
    _msgCtrl.clear();
    try {
      await apiClient.dio.post(
        '/api/messaging/conversations/${widget.convId}/messages',
        data: {'content': text},
      );
      await _loadMessages();
    } catch (_) {}
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
                          final msg = _messages[_messages.length - 1 - i]
                              as Map<String, dynamic>;
                          return _MessageBubble(
                            msg: msg,
                            userNames: widget.userNames,
                            isMine:
                                (msg['sender_id'] as num?)?.toInt() ==
                                    currentUserId,
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
                        contentPadding:
                            EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                      ),
                      onSubmitted: (_) => _send(),
                    ),
                  ),
                  const SizedBox(width: 8),
                  IconButton.filled(
                    tooltip: 'Отправить',
                    icon: const Icon(Icons.arrow_upward_rounded),
                    onPressed: _send,
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
  const _MessageBubble({
    required this.msg,
    required this.userNames,
    required this.isMine,
  });

  @override
  Widget build(BuildContext context) {
    final content = msg['content'] as String? ?? '';
    final senderId = (msg['sender_id'] as num?)?.toInt();
    final senderName = senderId == null
        ? 'Неизвестно'
        : userNames[senderId] ?? 'Пользователь №$senderId';

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment:
            isMine ? CrossAxisAlignment.end : CrossAxisAlignment.start,
        children: [
          if (!isMine)
            Padding(
              padding: const EdgeInsets.only(left: 4),
              child: Text(
                senderName,
                style: const TextStyle(color: AppTheme.mutedColor, fontSize: 11),
              ),
            ),
          const SizedBox(height: 2),
          FractionallySizedBox(
            widthFactor: 0.82,
            alignment: isMine ? Alignment.centerRight : Alignment.centerLeft,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
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
              child: Text(
                content,
                style: const TextStyle(color: Colors.white, height: 1.35),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
