import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../app/theme.dart';
import '../../../shared/widgets/app_states.dart';
import '../../auth/providers/auth_provider.dart';
import '../models/messaging_models.dart';
import '../providers/messaging_providers.dart';
import '../widgets/chat_list_tab.dart';
import '../widgets/chat_settings_tab.dart';
import '../widgets/create_chat_dialog.dart';
import '../widgets/drom_auth_dialog.dart';
import '../widgets/drom_dialogs_tab.dart';
import 'chat_screen.dart';
import 'drom_chat_screen.dart';

class MessagingScreen extends ConsumerStatefulWidget {
  const MessagingScreen({super.key});

  @override
  ConsumerState<MessagingScreen> createState() => _MessagingScreenState();
}

class _MessagingScreenState extends ConsumerState<MessagingScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  Timer? _dromPollTimer;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _tabController.addListener(_handleTabChange);
    _setupDromPolling();
  }

  @override
  void dispose() {
    _dromPollTimer?.cancel();
    _tabController.removeListener(_handleTabChange);
    _tabController.dispose();
    super.dispose();
  }

  void _handleTabChange() {
    if (mounted) setState(() {});
  }

  void _setupDromPolling() {
    _dromPollTimer?.cancel();
    final settings = ref.read(chatSettingsProvider);
    if (!settings.enableAutoRefresh) return;

    final intervalSeconds = settings.refreshInterval.clamp(10, 300);
    _dromPollTimer = Timer.periodic(Duration(seconds: intervalSeconds), (_) {
      if (_tabController.index == 1) {
        ref.invalidate(dromDialogsProvider);
      }
    });
  }

  Future<void> _refreshAll() async {
    ref.invalidate(conversationsProvider);
    ref.invalidate(messagingUsersProvider);
    ref.invalidate(userStatusesProvider);
    ref.invalidate(dromDialogsProvider);
    await Future.wait([
      ref.read(conversationsProvider.future).catchError((_) => <Conversation>[]),
      ref.read(messagingUsersProvider.future).catchError((_) => <int, ChatUser>{}),
      ref.read(userStatusesProvider.future).catchError((_) => <int, UserStatus>{}),
      ref.read(dromDialogsProvider.future).catchError((_) => <DromDialog>[]),
    ]);
  }

  void _openCreateChat(Map<int, ChatUser> users, int? currentUserId) {
    CreateChatDialog.show(
      context,
      users: users,
      currentUserId: currentUserId,
      onSuccess: (newConv) {
        ref.invalidate(conversationsProvider);
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => ChatScreen(
              convId: newConv.id,
              title: newConv.title.isNotEmpty ? newConv.title : 'Новый диалог',
              initialParticipants: newConv.participants,
            ),
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    // Listen for settings change to update polling timer
    ref.listen(chatSettingsProvider, (prev, next) {
      if (prev?.enableAutoRefresh != next.enableAutoRefresh ||
          prev?.refreshInterval != next.refreshInterval) {
        _setupDromPolling();
      }
    });

    final currentUserId = ref.watch(authProvider).valueOrNull?.id;
    final currentUserRole = ref.watch(authProvider).valueOrNull?.role ?? 'operator';
    final isAdmin = currentUserRole == 'admin';

    final convsAsync = ref.watch(conversationsProvider);
    final usersAsync = ref.watch(messagingUsersProvider);
    final userStatusesAsync = ref.watch(userStatusesProvider);
    final dromDialogsAsync = ref.watch(dromDialogsProvider);

    final users = usersAsync.valueOrNull ?? const {};
    final userStatuses = userStatusesAsync.valueOrNull ?? const {};
    final dromDialogs = dromDialogsAsync.valueOrNull ?? const [];

    final hasUnreadDrom = dromDialogs.any((d) => d.isUnread);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Сообщения'),
        actions: [
          if (_tabController.index == 0) ...[
            IconButton(
              tooltip: 'Новый диалог',
              icon: const Icon(LucideIcons.square_pen),
              onPressed: () => _openCreateChat(users, currentUserId),
            ),
            IconButton(
              tooltip: 'Обновить чаты',
              icon: const Icon(LucideIcons.refresh_cw),
              onPressed: () {
                ref.invalidate(conversationsProvider);
                ref.invalidate(userStatusesProvider);
              },
            ),
          ] else if (_tabController.index == 1) ...[
            IconButton(
              tooltip: 'Аккаунт Drom.ru',
              icon: const Icon(LucideIcons.key_round),
              onPressed: () {
                DromAuthDialog.show(
                  context,
                  onSuccess: () => ref.invalidate(dromDialogsProvider),
                );
              },
            ),
            IconButton(
              tooltip: 'Обновить Drom диалоги',
              icon: const Icon(LucideIcons.refresh_cw),
              onPressed: () => ref.invalidate(dromDialogsProvider),
            ),
          ] else ...[
            IconButton(
              tooltip: 'Обновить все',
              icon: const Icon(LucideIcons.refresh_cw),
              onPressed: _refreshAll,
            ),
          ],
        ],
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: AppTheme.primaryColor,
          indicatorWeight: 3,
          labelColor: Colors.white,
          unselectedLabelColor: AppTheme.mutedColor,
          labelStyle: const TextStyle(fontWeight: FontWeight.w700, fontSize: 13),
          tabs: [
            const Tab(text: 'Сообщения'),
            Tab(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Text('Drom сообщения'),
                  if (hasUnreadDrom) ...[
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
            ),
            const Tab(
              icon: Icon(LucideIcons.settings, size: 18),
            ),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          // Tab 1: Internal Employee Chats
          convsAsync.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (err, _) => AppEmptyState(
              icon: LucideIcons.cloud_off,
              title: 'Не удалось загрузить сообщения',
              message: 'Проверьте соединение с сервером и повторите попытку.',
              actionLabel: 'Повторить',
              onAction: () => ref.invalidate(conversationsProvider),
            ),
            data: (convs) => ChatListTab(
              conversations: convs,
              users: users,
              userStatuses: userStatuses,
              currentUserId: currentUserId,
              isAdmin: isAdmin,
              onRefresh: () async {
                ref.invalidate(conversationsProvider);
                ref.invalidate(userStatusesProvider);
                await ref.read(conversationsProvider.future);
              },
              onSelectConversation: (conv) {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => ChatScreen(
                      convId: conv.id,
                      title: conv.title.isNotEmpty ? conv.title : 'Диалог #${conv.id}',
                      initialParticipants: conv.participants,
                    ),
                  ),
                ).then((_) => ref.invalidate(conversationsProvider));
              },
            ),
          ),

          // Tab 2: Drom Customer Messages
          dromDialogsAsync.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (err, _) => AppEmptyState(
              icon: LucideIcons.cloud_off,
              title: 'Не удалось загрузить Drom сообщения',
              message: err.toString(),
              actionLabel: 'Повторить',
              onAction: () => ref.invalidate(dromDialogsProvider),
            ),
            data: (dialogs) => DromDialogsTab(
              dialogs: dialogs,
              isLoading: dromDialogsAsync.isLoading,
              onRefresh: () async {
                ref.invalidate(dromDialogsProvider);
                await ref.read(dromDialogsProvider.future);
              },
              onSelectDialog: (dialog) {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => DromChatScreen(initialDialog: dialog),
                  ),
                ).then((_) => ref.invalidate(dromDialogsProvider));
              },
            ),
          ),

          // Tab 3: Chat Settings
          const ChatSettingsTab(),
        ],
      ),
    );
  }
}
