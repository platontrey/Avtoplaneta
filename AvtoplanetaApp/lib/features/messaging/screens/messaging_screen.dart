import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
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

class MessagingScreen extends ConsumerWidget {
  const MessagingScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final convsAsync = ref.watch(conversationsProvider);

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
                  return ListTile(
                    leading: CircleAvatar(
                      backgroundColor: const Color(0xFF4F8EF7),
                      child: Text(
                        (conv['title'] as String? ?? '?')[0].toUpperCase(),
                        style: const TextStyle(color: Colors.white),
                      ),
                    ),
                    title: Text(
                      conv['title'] as String? ?? 'Диалог',
                      style: const TextStyle(color: Colors.white),
                    ),
                    subtitle: Text(
                      conv['last_message'] as String? ?? '',
                      style: const TextStyle(color: Colors.white54, fontSize: 12),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    onTap: () => _openChat(context, conv),
                  );
                },
              ),
      ),
    );
  }

  void _openChat(BuildContext context, Map<String, dynamic> conv) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => _ChatScreen(
          convId: conv['id'] as int,
          title: conv['title'] as String? ?? 'Диалог',
        ),
      ),
    );
  }
}

class _ChatScreen extends ConsumerStatefulWidget {
  final int convId;
  final String title;
  const _ChatScreen({required this.convId, required this.title});

  @override
  ConsumerState<_ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<_ChatScreen> {
  final _msgCtrl = TextEditingController();
  List<dynamic> _messages = [];
  bool _loading = true;
  WebSocketChannel? _channel;
  bool _isDisposed = false;

  @override
  void initState() {
    super.initState();
    _loadMessages();
    _connectWebSocket();
  }

  void _connectWebSocket() {
    if (_isDisposed) return;
    final user = ref.read(authProvider).valueOrNull;
    if (user == null) return;

    try {
      final wsUrl = Uri.parse('ws://192.168.1.63:8084/api/messaging/ws?userId=${user.id}');
      _channel = WebSocketChannel.connect(wsUrl);
      _channel!.stream.listen((message) {
        try {
          final payload = jsonDecode(message as String);
          if (payload['event'] == 'new_message') {
            final msgData = payload['data'];
            if (msgData['conversation_id'] == widget.convId) {
              setState(() {
                final id = msgData['id'];
                if (!_messages.any((m) => m['id'] == id)) {
                  _messages.add(msgData);
                }
              });
            }
          }
        } catch (_) {}
      }, onError: (err) {
        _reconnect();
      }, onDone: () {
        _reconnect();
      });
    } catch (_) {
      _reconnect();
    }
  }

  void _reconnect() {
    if (_isDisposed) return;
    Future.delayed(const Duration(seconds: 3), () {
      if (!_isDisposed) {
        _connectWebSocket();
      }
    });
  }

  @override
  void dispose() {
    _isDisposed = true;
    _channel?.sink.close();
    _msgCtrl.dispose();
    super.dispose();
  }

  Future<void> _loadMessages() async {
    try {
      final response = await apiClient.dio
          .get('/api/messaging/conversations/${widget.convId}/messages');
      final data = response.data;
      setState(() {
        _messages = (data is Map ? data['messages'] ?? [] : []) as List<dynamic>;
        _loading = false;
      });
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  Future<void> _send() async {
    final text = _msgCtrl.text.trim();
    if (text.isEmpty) return;
    _msgCtrl.clear();
    try {
      await apiClient.dio.post(
        '/api/messaging/conversations/${widget.convId}/messages',
        data: {'content': text, 'type': 'text'},
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
                      return _MessageBubble(msg: msg);
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
  const _MessageBubble({required this.msg});

  @override
  Widget build(BuildContext context) {
    final content = msg['content'] as String? ?? '';
    final senderName = msg['sender_name'] as String? ?? 'Неизвестно';

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
