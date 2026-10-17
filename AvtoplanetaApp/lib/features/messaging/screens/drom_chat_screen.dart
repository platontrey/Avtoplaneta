import 'dart:async';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../app/theme.dart';
import '../../../shared/widgets/app_states.dart';
import '../models/messaging_models.dart';
import '../providers/messaging_providers.dart';

class DromChatScreen extends ConsumerStatefulWidget {
  final DromDialog initialDialog;

  const DromChatScreen({super.key, required this.initialDialog});

  @override
  ConsumerState<DromChatScreen> createState() => _DromChatScreenState();
}

class _DromChatScreenState extends ConsumerState<DromChatScreen> {
  final _msgController = TextEditingController();
  final _scrollController = ScrollController();
  late DromDialog _dialog;
  List<DromMessage> _messages = [];
  bool _loading = true;
  bool _sending = false;
  String? _error;
  Timer? _pollingTimer;

  @override
  void initState() {
    super.initState();
    _dialog = widget.initialDialog;
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
      final res = await MessagingRepository.getDromMessages(_dialog.dialogId);
      if (!mounted) return;

      final messages = (res['messages'] as List<DromMessage>?) ?? [];
      final updatedDialog = res['dialog'] as DromDialog?;

      setState(() {
        _messages = messages;
        if (updatedDialog != null) {
          _dialog = updatedDialog;
        }
        _loading = false;
        _error = null;
      });
    } catch (e) {
      if (mounted && !silent) {
        setState(() {
          _loading = false;
          _error = e.toString();
        });
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
      final msg = await MessagingRepository.sendDromMessage(_dialog.dialogId, text);
      if (mounted) {
        _msgController.clear();
        setState(() {
          _messages.add(msg);
        });
        _scrollToBottom();
        ref.invalidate(dromDialogsProvider);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось отправить сообщение: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _sending = false);
    }
  }

  Future<void> _deleteMessage(int messageId) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить сообщение?'),
        content: const Text('Сообщение будет скрыто из диалога.'),
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

    if (confirmed == true) {
      setState(() {
        _messages.removeWhere((m) => m.id == messageId);
      });
    }
  }

  Future<void> _openPartUrl(String url) async {
    final uri = Uri.tryParse(url);
    if (uri != null && await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось открыть ссылку: $url')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        titleSpacing: 0,
        title: Row(
          children: [
            Flexible(
              child: Text(
                _dialog.interlocutor,
                style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700),
                overflow: TextOverflow.ellipsis,
              ),
            ),
            if (_dialog.userRating != null && _dialog.userRating!.isNotEmpty) ...[
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: AppTheme.successColor.withValues(alpha: 0.18),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  _dialog.userRating!,
                  style: const TextStyle(
                    color: AppTheme.successColor,
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ],
            if (_dialog.userStatus != null && _dialog.userStatus!.isNotEmpty) ...[
              const SizedBox(width: 6),
              Flexible(
                child: Text(
                  '• ${_dialog.userStatus}',
                  style: const TextStyle(fontSize: 11, color: AppTheme.mutedColor),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ],
        ),
      ),
      body: Column(
        children: [
          // Pinned Part Info Header (matches web app)
          if ((_dialog.partTitle != null && _dialog.partTitle!.isNotEmpty) ||
              (_dialog.partUrl != null && _dialog.partUrl!.isNotEmpty)) ...[
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
              decoration: const BoxDecoration(
                color: AppTheme.cardColor,
                border: Border(bottom: BorderSide(color: AppTheme.borderColor)),
              ),
              child: Row(
                children: [
                  if (_dialog.partPhotoUrl != null && _dialog.partPhotoUrl!.isNotEmpty) ...[
                    ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: CachedNetworkImage(
                        imageUrl: _dialog.partPhotoUrl!,
                        width: 40,
                        height: 40,
                        fit: BoxFit.cover,
                        errorWidget: (_, _, _) => const Icon(LucideIcons.package, size: 20),
                      ),
                    ),
                    const SizedBox(width: 10),
                  ],
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _dialog.partTitle ?? 'Деталь на Дроме',
                          style: const TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.w700,
                            color: Colors.white,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        if (_dialog.partPrice != null && _dialog.partPrice!.isNotEmpty)
                          Text(
                            _dialog.partPrice!,
                            style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: AppTheme.primaryColor,
                            ),
                          ),
                      ],
                    ),
                  ),
                  if (_dialog.partUrl != null && _dialog.partUrl!.isNotEmpty)
                    IconButton.filledTonal(
                      tooltip: 'Открыть на Drom.ru',
                      icon: const Icon(LucideIcons.external_link, size: 16),
                      onPressed: () => _openPartUrl(_dialog.partUrl!),
                    ),
                ],
              ),
            ),
          ],

          // Messages List
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
                            icon: LucideIcons.message_circle,
                            title: 'Диалог пуст',
                            message: 'Покупатель пока не написал сообщение или оно ожидает синхронизации.',
                          )
                        : ListView.builder(
                            controller: _scrollController,
                            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                            itemCount: _messages.length,
                            itemBuilder: (context, index) {
                              final msg = _messages[index];
                              final isOutgoing = msg.isOutgoing;

                              return Padding(
                                padding: const EdgeInsets.symmetric(vertical: 4),
                                child: Align(
                                  alignment: isOutgoing ? Alignment.centerRight : Alignment.centerLeft,
                                  child: Container(
                                    constraints: BoxConstraints(
                                      maxWidth: MediaQuery.of(context).size.width * 0.78,
                                    ),
                                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                                    decoration: BoxDecoration(
                                      color: isOutgoing ? AppTheme.primaryColor : AppTheme.cardColor,
                                      borderRadius: BorderRadius.only(
                                        topLeft: const Radius.circular(18),
                                        topRight: const Radius.circular(18),
                                        bottomLeft: Radius.circular(isOutgoing ? 18 : 4),
                                        bottomRight: Radius.circular(isOutgoing ? 4 : 18),
                                      ),
                                      border: isOutgoing ? null : Border.all(color: AppTheme.borderColor),
                                    ),
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Row(
                                          mainAxisSize: MainAxisSize.min,
                                          children: [
                                            Expanded(
                                              child: Text(
                                                msg.text,
                                                style: const TextStyle(
                                                  color: Colors.white,
                                                  fontSize: 14,
                                                  height: 1.35,
                                                ),
                                              ),
                                            ),
                                            if (isOutgoing) ...[
                                              const SizedBox(width: 4),
                                              InkWell(
                                                onTap: () => _deleteMessage(msg.id),
                                                child: Icon(
                                                  LucideIcons.x,
                                                  size: 14,
                                                  color: Colors.white.withValues(alpha: 0.6),
                                                ),
                                              ),
                                            ],
                                          ],
                                        ),
                                        const SizedBox(height: 4),
                                        Align(
                                          alignment: Alignment.bottomRight,
                                          child: Text(
                                            msg.time,
                                            style: TextStyle(
                                              color: isOutgoing
                                                  ? Colors.white.withValues(alpha: 0.7)
                                                  : AppTheme.mutedColor,
                                              fontSize: 10,
                                            ),
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
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
                        hintText: 'Ответить покупателю…',
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
