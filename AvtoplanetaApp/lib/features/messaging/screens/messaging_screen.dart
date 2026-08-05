import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
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
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.invalidate(conversationsProvider),
          ),
        ],
      ),
      body: convsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (convs) => convs.isEmpty
            ? const Center(
                child: Text('Нет диалогов',
                    style: TextStyle(color: Colors.white54)))
            : ListView.builder(
                itemCount: convs.length,
                itemBuilder: (_, i) {
                  final conv = convs[i] as Map<String, dynamic>;
                  final title =
                      _conversationTitle(conv, userNames, currentUserId);
                  return ListTile(
                    leading: CircleAvatar(
                      backgroundColor: const Color(0xFF4F8EF7),
                      child: Text(
                        title.isEmpty ? '?' : title[0].toUpperCase(),
                        style: const TextStyle(color: Colors.white),
                      ),
                    ),
                    title: Text(
                      title,
                      style: const TextStyle(color: Colors.white),
                    ),
                    subtitle: Text(
                      conv['last_message'] as String? ?? '',
                      style: const TextStyle(color: Colors.white54, fontSize: 12),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    onTap: () => _openChat(
                      context,
                      conv,
                      userNames,
                      currentUserId,
                    ),
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
    return Scaffold(
      appBar: AppBar(title: Text(widget.title)),
      body: Column(
        children: [
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : ListView.builder(
                    reverse: true,
                    padding: const EdgeInsets.all(12),
                    itemCount: _messages.length,
                    itemBuilder: (_, i) {
                      final msg = _messages[_messages.length - 1 - i]
                          as Map<String, dynamic>;
                      return _MessageBubble(
                        msg: msg,
                        userNames: widget.userNames,
                      );
                    },
                  ),
          ),
          // Поле ввода
          Container(
            color: const Color(0xFF16213E),
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _msgCtrl,
                    style: const TextStyle(color: Colors.white),
                    decoration: const InputDecoration(
                      hintText: 'Сообщение...',
                      hintStyle: TextStyle(color: Colors.white38),
                      contentPadding:
                          EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                    ),
                    onSubmitted: (_) => _send(),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.send, color: Color(0xFF4F8EF7)),
                  onPressed: _send,
                ),
              ],
            ),
          ),
          SizedBox(height: MediaQuery.of(context).viewInsets.bottom),
        ],
      ),
    );
  }
}

class _MessageBubble extends StatelessWidget {
  final Map<String, dynamic> msg;
  final Map<int, String> userNames;
  const _MessageBubble({required this.msg, required this.userNames});

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
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(senderName,
              style: const TextStyle(color: Colors.white38, fontSize: 11)),
          const SizedBox(height: 2),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: const Color(0xFF0F3460),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Text(content, style: const TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }
}
