import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../../app/theme.dart';
import '../../../core/api/api_client.dart';
import '../../../core/api/api_endpoints.dart';
import '../../../core/models/user.dart';
import '../../../core/services/update_service.dart';
import '../../inventory/providers/inventory_provider.dart';
import '../../updater/widgets/update_dialog.dart';

// ─── Модели данных ─────────────────────────────────────────────────────────

class SupplierBatch {
  final String code;
  final String label;
  final int zeroCount;

  const SupplierBatch({
    required this.code,
    required this.label,
    this.zeroCount = 0,
  });

  factory SupplierBatch.fromJson(Map<String, dynamic> json) {
    final code = (json['code'] ?? '').toString();
    final label = (json['label'] ?? code).toString();
    final zeroCount = (json['zero_count'] as num?)?.toInt() ?? 0;
    return SupplierBatch(code: code, label: label, zeroCount: zeroCount);
  }
}

class ServerStatusInfo {
  final String serverStatus;
  final String goVersion;
  final String os;
  final String arch;
  final String dbStatus;
  final int totalParts;
  final int totalUsers;
  final String timestamp;

  const ServerStatusInfo({
    required this.serverStatus,
    required this.goVersion,
    required this.os,
    required this.arch,
    required this.dbStatus,
    required this.totalParts,
    required this.totalUsers,
    required this.timestamp,
  });

  factory ServerStatusInfo.fromJson(Map<String, dynamic> json) {
    final server = json['server'] as Map<String, dynamic>? ?? {};
    final db = json['database'] as Map<String, dynamic>? ?? {};
    return ServerStatusInfo(
      serverStatus: (server['status'] ?? 'unknown').toString(),
      goVersion: (server['go_version'] ?? 'unknown').toString(),
      os: (server['os'] ?? 'unknown').toString(),
      arch: (server['arch'] ?? 'unknown').toString(),
      dbStatus: (db['status'] ?? 'unknown').toString(),
      totalParts: (db['total_parts'] as num?)?.toInt() ?? 0,
      totalUsers: (db['total_users'] as num?)?.toInt() ?? 0,
      timestamp: (json['timestamp'] ?? '').toString(),
    );
  }
}

class ServerLogEntry {
  final String timestamp;
  final String level;
  final String message;

  const ServerLogEntry({
    required this.timestamp,
    required this.level,
    required this.message,
  });

  factory ServerLogEntry.fromJson(Map<String, dynamic> json) {
    return ServerLogEntry(
      timestamp: (json['timestamp'] ?? '').toString(),
      level: (json['level'] ?? 'INFO').toString(),
      message: (json['message'] ?? '').toString(),
    );
  }
}

class UserActivityLogEntry {
  final int id;
  final int userId;
  final String userName;
  final String userEmail;
  final String action;
  final String resourceType;
  final int? resourceId;
  final String details;
  final String createdAt;

  const UserActivityLogEntry({
    required this.id,
    required this.userId,
    required this.userName,
    required this.userEmail,
    required this.action,
    required this.resourceType,
    this.resourceId,
    required this.details,
    required this.createdAt,
  });

  factory UserActivityLogEntry.fromJson(Map<String, dynamic> json) {
    return UserActivityLogEntry(
      id: (json['id'] as num?)?.toInt() ?? 0,
      userId: (json['user_id'] as num?)?.toInt() ?? 0,
      userName: (json['user_name'] ?? 'Пользователь').toString(),
      userEmail: (json['user_email'] ?? '').toString(),
      action: (json['action'] ?? '').toString(),
      resourceType: (json['resource_type'] ?? '').toString(),
      resourceId: (json['resource_id'] as num?)?.toInt(),
      details: (json['details'] ?? '').toString(),
      createdAt: (json['created_at'] ?? '').toString(),
    );
  }
}

class ActivityFilterState {
  final bool usefulOnly;
  final int? selectedUserId;
  final String? selectedAction;

  const ActivityFilterState({
    this.usefulOnly = true,
    this.selectedUserId,
    this.selectedAction,
  });

  ActivityFilterState copyWith({
    bool? usefulOnly,
    int? Function()? selectedUserId,
    String? Function()? selectedAction,
  }) {
    return ActivityFilterState(
      usefulOnly: usefulOnly ?? this.usefulOnly,
      selectedUserId: selectedUserId != null ? selectedUserId() : this.selectedUserId,
      selectedAction: selectedAction != null ? selectedAction() : this.selectedAction,
    );
  }
}

// ─── Провайдеры ────────────────────────────────────────────────────────────

final usersListProvider = FutureProvider<List<User>>((ref) async {
  final response = await apiClient.dio.get(ApiEndpoints.adminUsers);
  final data = response.data as Map<String, dynamic>;
  final list = data['users'] as List<dynamic>? ?? [];
  return list.map((e) => User.fromJson(e as Map<String, dynamic>)).toList();
});

final supplierBatchesProvider = FutureProvider<List<SupplierBatch>>((ref) async {
  final response = await apiClient.dio.get(ApiEndpoints.adminSupplierCodes);
  final data = response.data as Map<String, dynamic>? ?? {};
  final rawBatches = data['batches'];
  if (rawBatches is List && rawBatches.isNotEmpty) {
    return rawBatches
        .whereType<Map>()
        .map((e) => SupplierBatch.fromJson(Map<String, dynamic>.from(e)))
        .where((b) => b.code.isNotEmpty)
        .toList();
  }
  final rawCodes = (data['supplier_codes'] ?? data['codes']) as List<dynamic>? ?? [];
  return rawCodes
      .map((e) => e.toString())
      .where((c) => c.isNotEmpty)
      .map((c) => SupplierBatch(code: c, label: c))
      .toList();
});

final serverStatusProvider = FutureProvider<ServerStatusInfo>((ref) async {
  final response = await apiClient.dio.get(ApiEndpoints.adminStatus);
  final data = response.data as Map<String, dynamic>? ?? {};
  return ServerStatusInfo.fromJson(data);
});

final serverLogsProvider = FutureProvider<List<ServerLogEntry>>((ref) async {
  final response = await apiClient.dio.get(ApiEndpoints.adminLogs);
  final data = response.data as Map<String, dynamic>? ?? {};
  final list = data['logs'] as List<dynamic>? ?? [];
  return list
      .whereType<Map>()
      .map((e) => ServerLogEntry.fromJson(Map<String, dynamic>.from(e)))
      .toList();
});

final activityFilterProvider = StateProvider<ActivityFilterState>((ref) {
  return const ActivityFilterState();
});

final userActivityLogsProvider = FutureProvider<List<UserActivityLogEntry>>((ref) async {
  final filters = ref.watch(activityFilterProvider);
  final queryParams = <String, dynamic>{
    'limit': 100,
    if (filters.usefulOnly) 'useful_only': 'true',
    if (filters.selectedUserId != null) 'user_id': filters.selectedUserId,
    if (filters.selectedAction != null && filters.selectedAction != 'all')
      'action': filters.selectedAction,
  };

  final response = await apiClient.dio.get(
    ApiEndpoints.adminUserActivityLogs,
    queryParameters: queryParams,
  );
  final data = response.data as Map<String, dynamic>? ?? {};
  final list = data['logs'] as List<dynamic>? ?? [];
  return list
      .whereType<Map>()
      .map((e) => UserActivityLogEntry.fromJson(Map<String, dynamic>.from(e)))
      .toList();
});

// ─── Главный экран администрирования ────────────────────────────────────────

class AdminScreen extends ConsumerStatefulWidget {
  const AdminScreen({super.key});

  @override
  ConsumerState<AdminScreen> createState() => _AdminScreenState();
}

class _AdminScreenState extends ConsumerState<AdminScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  void _refreshAll() {
    ref.invalidate(usersListProvider);
    ref.invalidate(supplierBatchesProvider);
    ref.invalidate(serverStatusProvider);
    ref.invalidate(serverLogsProvider);
    ref.invalidate(userActivityLogsProvider);
  }

  Future<void> _checkAppUpdates(BuildContext context) async {
    final scaffoldMessenger = ScaffoldMessenger.of(context);
    scaffoldMessenger.showSnackBar(
      const SnackBar(
        content: Text('Проверка наличия обновлений...'),
        duration: Duration(seconds: 1),
      ),
    );

    final updateService = ref.read(updateServiceProvider);
    final result = await updateService.checkForUpdates();

    if (!context.mounted) return;

    if (result.hasUpdate && result.updateInfo != null) {
      UpdateDialog.show(
        context: context,
        info: result.updateInfo!,
        updateService: updateService,
      );
    } else {
      scaffoldMessenger.showSnackBar(
        SnackBar(
          content: Text(
            'У вас установлена последняя версия (v${result.currentVersion}+${result.currentBuildNumber})',
          ),
          backgroundColor: Colors.green.shade800,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (!didPop) {
          if (context.canPop()) {
            context.pop();
          } else {
            context.go('/inventory');
          }
        }
      },
      child: Scaffold(
        appBar: AppBar(
          leading: IconButton(
            icon: const Icon(LucideIcons.arrow_left),
            tooltip: 'Назад',
            onPressed: () {
              if (context.canPop()) {
                context.pop();
              } else {
                context.go('/inventory');
              }
            },
          ),
          title: const Text('Панель администратора'),
          actions: [
            IconButton(
              icon: const Icon(LucideIcons.arrow_down_to_line),
              tooltip: 'Проверить обновления',
              onPressed: () => _checkAppUpdates(context),
            ),
            IconButton(
              icon: const Icon(LucideIcons.refresh_cw),
              tooltip: 'Обновить всё',
              onPressed: _refreshAll,
            ),
          ],
          bottom: TabBar(
            controller: _tabController,
            isScrollable: false,
            indicatorColor: AppTheme.primaryColor,
            labelColor: AppTheme.primaryColor,
            unselectedLabelColor: Colors.white70,
            tabs: const [
              Tab(icon: Icon(LucideIcons.users, size: 20), text: 'Люди'),
              Tab(icon: Icon(LucideIcons.package, size: 20), text: 'Запчасти'),
              Tab(icon: Icon(LucideIcons.activity, size: 20), text: 'Активность'),
              Tab(icon: Icon(LucideIcons.server, size: 20), text: 'Система'),
            ],
          ),
        ),
        body: TabBarView(
          controller: _tabController,
          children: const [
            _UsersTab(),
            _PartsTab(),
            _ActivityTab(),
            _SystemTab(),
          ],
        ),
      ),
    );
  }
}

// ─── Вкладка 1: Пользователи ────────────────────────────────────────────────

class _UsersTab extends ConsumerWidget {
  const _UsersTab();

  void _showCreateUser(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (_) => _CreateUserDialog(
        onCreated: () => ref.invalidate(usersListProvider),
      ),
    );
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final usersAsync = ref.watch(usersListProvider);

    return Scaffold(
      body: usersAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(LucideIcons.circle_alert, size: 64, color: Colors.orange.shade400),
                const SizedBox(height: 16),
                const Text(
                  'Не удалось загрузить пользователей',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),
                Text(
                  e.toString(),
                  style: const TextStyle(color: Colors.white70, fontSize: 13),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 20),
                FilledButton.icon(
                  onPressed: () => ref.invalidate(usersListProvider),
                  icon: const Icon(LucideIcons.refresh_cw),
                  label: const Text('Повторить'),
                ),
              ],
            ),
          ),
        ),
        data: (users) => RefreshIndicator(
          onRefresh: () async => ref.invalidate(usersListProvider),
          child: ListView(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
            children: [
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Всего сотрудников: ${users.length}',
                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                    ),
                    FilledButton.tonalIcon(
                      onPressed: () => _showCreateUser(context, ref),
                      icon: const Icon(LucideIcons.user_plus, size: 16),
                      label: const Text('Добавить'),
                    ),
                  ],
                ),
              ),
              ...users.map(
                (u) => _UserTile(
                  user: u,
                  onChanged: () => ref.invalidate(usersListProvider),
                ),
              ),
              const SizedBox(height: 80),
            ],
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        tooltip: 'Добавить пользователя',
        onPressed: () => _showCreateUser(context, ref),
        child: const Icon(LucideIcons.user_plus),
      ),
    );
  }
}

class _UserTile extends ConsumerWidget {
  final User user;
  final VoidCallback onChanged;

  const _UserTile({
    required this.user,
    required this.onChanged,
  });

  Color _roleColor(String role) {
    switch (role) {
      case 'admin':
        return const Color(0xFFE53935);
      case 'manager':
        return const Color(0xFFFDD835);
      default:
        return const Color(0xFF43A047);
    }
  }

  void _showEdit(BuildContext context) {
    showDialog(
      context: context,
      builder: (_) => _EditUserDialog(user: user, onUpdated: onChanged),
    );
  }

  void _confirmDelete(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить пользователя?'),
        content: Text('Учетная запись «${user.name}» (${user.email}) будет безвозвратно удалена.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Отмена'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: Colors.red.shade700),
            onPressed: () async {
              Navigator.pop(ctx);
              try {
                await apiClient.dio.delete(ApiEndpoints.adminUserById(user.id));
                onChanged();
                if (context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      content: Text('Пользователь ${user.name} удален'),
                      backgroundColor: Colors.green.shade800,
                    ),
                  );
                }
              } catch (e) {
                if (context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      content: Text('Ошибка удаления: $e'),
                      backgroundColor: Colors.red.shade800,
                    ),
                  );
                }
              }
            },
            child: const Text('Удалить'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final roleColor = _roleColor(user.role);

    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 4),
      color: AppTheme.surfaceColor,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: Colors.white10),
      ),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                CircleAvatar(
                  backgroundColor: roleColor.withValues(alpha: 0.2),
                  child: Text(
                    user.initials.isNotEmpty
                        ? user.initials
                        : user.name.isNotEmpty
                            ? user.name[0]
                            : '?',
                    style: TextStyle(
                      color: roleColor,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        user.name,
                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                      ),
                      Text(
                        user.email,
                        style: const TextStyle(color: Colors.white70, fontSize: 13),
                      ),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: roleColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(6),
                    border: Border.all(color: roleColor.withValues(alpha: 0.4)),
                  ),
                  child: Text(
                    user.role.toUpperCase(),
                    style: TextStyle(color: roleColor, fontSize: 11, fontWeight: FontWeight.bold),
                  ),
                ),
              ],
            ),
            if (user.initials.isNotEmpty || user.inn.isNotEmpty) ...[
              const SizedBox(height: 8),
              Wrap(
                spacing: 12,
                children: [
                  if (user.initials.isNotEmpty)
                    Text('Инициалы: ${user.initials}', style: const TextStyle(color: Colors.white60, fontSize: 12)),
                  if (user.inn.isNotEmpty)
                    Text('ИНН: ${user.inn}', style: const TextStyle(color: Colors.white60, fontSize: 12)),
                ],
              ),
            ],
            const Divider(height: 16, color: Colors.white10),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                OutlinedButton.icon(
                  onPressed: () => _showEdit(context),
                  icon: const Icon(LucideIcons.pencil, size: 16),
                  label: const Text('Изменить'),
                ),
                const SizedBox(width: 8),
                IconButton(
                  tooltip: 'Удалить',
                  icon: const Icon(LucideIcons.trash, color: Colors.redAccent, size: 18),
                  onPressed: () => _confirmDelete(context),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Диалог создания пользователя ──────────────────────────────────────────

class _CreateUserDialog extends StatefulWidget {
  final VoidCallback onCreated;
  const _CreateUserDialog({required this.onCreated});

  @override
  State<_CreateUserDialog> createState() => _CreateUserDialogState();
}

class _CreateUserDialogState extends State<_CreateUserDialog> {
  final _nameCtrl = TextEditingController();
  final _emailCtrl = TextEditingController();
  final _passCtrl = TextEditingController();
  final _initialsCtrl = TextEditingController();
  final _innCtrl = TextEditingController();
  String _role = 'operator';
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _nameCtrl.dispose();
    _emailCtrl.dispose();
    _passCtrl.dispose();
    _initialsCtrl.dispose();
    _innCtrl.dispose();
    super.dispose();
  }

  Future<void> _create() async {
    final name = _nameCtrl.text.trim();
    final email = _emailCtrl.text.trim();
    final pass = _passCtrl.text;
    if (name.isEmpty || email.isEmpty || pass.isEmpty) {
      setState(() => _error = 'Заполните обязательные поля (Имя, Email, Пароль)');
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      await apiClient.dio.post(
        ApiEndpoints.adminUsers,
        data: {
          'name': name,
          'email': email,
          'password': pass,
          'initials': _initialsCtrl.text.trim(),
          'inn': _innCtrl.text.trim(),
          'role': _role,
        },
      );
      widget.onCreated();
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Пользователь $name успешно создан'),
            backgroundColor: Colors.green.shade800,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = 'Ошибка создания: $e';
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Создать пользователя'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (_error != null)
              Container(
                margin: const EdgeInsets.only(bottom: 12),
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.red.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(_error!, style: const TextStyle(color: Colors.redAccent, fontSize: 12)),
              ),
            TextField(
              controller: _nameCtrl,
              decoration: const InputDecoration(labelText: 'Имя *'),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _emailCtrl,
              decoration: const InputDecoration(labelText: 'Email *'),
              keyboardType: TextInputType.emailAddress,
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _passCtrl,
              decoration: const InputDecoration(labelText: 'Пароль *'),
              obscureText: true,
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _initialsCtrl,
              decoration: const InputDecoration(labelText: 'Инициалы'),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _innCtrl,
              decoration: const InputDecoration(labelText: 'ИНН'),
              keyboardType: TextInputType.number,
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              initialValue: _role,
              items: const [
                DropdownMenuItem(value: 'operator', child: Text('Оператор')),
                DropdownMenuItem(value: 'manager', child: Text('Менеджер')),
                DropdownMenuItem(value: 'admin', child: Text('Администратор')),
              ],
              onChanged: (v) => setState(() => _role = v ?? 'operator'),
              decoration: const InputDecoration(labelText: 'Роль *'),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Отмена'),
        ),
        FilledButton(
          onPressed: _loading ? null : _create,
          child: Text(_loading ? 'Создание...' : 'Создать'),
        ),
      ],
    );
  }
}

// ─── Диалог редактирования пользователя ────────────────────────────────────

class _EditUserDialog extends StatefulWidget {
  final User user;
  final VoidCallback onUpdated;
  const _EditUserDialog({required this.user, required this.onUpdated});

  @override
  State<_EditUserDialog> createState() => _EditUserDialogState();
}

class _EditUserDialogState extends State<_EditUserDialog> {
  late final TextEditingController _nameCtrl;
  late final TextEditingController _emailCtrl;
  late final TextEditingController _initialsCtrl;
  late final TextEditingController _innCtrl;
  late String _role;
  bool _loading = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _nameCtrl = TextEditingController(text: widget.user.name);
    _emailCtrl = TextEditingController(text: widget.user.email);
    _initialsCtrl = TextEditingController(text: widget.user.initials);
    _innCtrl = TextEditingController(text: widget.user.inn);
    _role = widget.user.role;
  }

  @override
  void dispose() {
    _nameCtrl.dispose();
    _emailCtrl.dispose();
    _initialsCtrl.dispose();
    _innCtrl.dispose();
    super.dispose();
  }

  Future<void> _update() async {
    final name = _nameCtrl.text.trim();
    final email = _emailCtrl.text.trim();
    if (name.isEmpty || email.isEmpty) {
      setState(() => _error = 'Имя и Email обязательны');
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      await apiClient.dio.put(
        ApiEndpoints.adminUserById(widget.user.id),
        data: {
          'name': name,
          'email': email,
          'initials': _initialsCtrl.text.trim(),
          'inn': _innCtrl.text.trim(),
          'role': _role,
        },
      );
      widget.onUpdated();
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Данные пользователя $name обновлены'),
            backgroundColor: Colors.green.shade800,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = 'Ошибка сохранения: $e';
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text('Редактировать: ${widget.user.name}'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (_error != null)
              Container(
                margin: const EdgeInsets.only(bottom: 12),
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.red.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(_error!, style: const TextStyle(color: Colors.redAccent, fontSize: 12)),
              ),
            TextField(
              controller: _nameCtrl,
              decoration: const InputDecoration(labelText: 'Имя *'),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _emailCtrl,
              decoration: const InputDecoration(labelText: 'Email *'),
              keyboardType: TextInputType.emailAddress,
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _initialsCtrl,
              decoration: const InputDecoration(labelText: 'Инициалы'),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _innCtrl,
              decoration: const InputDecoration(labelText: 'ИНН'),
              keyboardType: TextInputType.number,
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              initialValue: _role,
              items: const [
                DropdownMenuItem(value: 'operator', child: Text('Оператор')),
                DropdownMenuItem(value: 'manager', child: Text('Менеджер')),
                DropdownMenuItem(value: 'admin', child: Text('Администратор')),
              ],
              onChanged: (v) => setState(() => _role = v ?? 'operator'),
              decoration: const InputDecoration(labelText: 'Роль *'),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Отмена'),
        ),
        FilledButton(
          onPressed: _loading ? null : _update,
          child: Text(_loading ? 'Сохранение...' : 'Сохранить'),
        ),
      ],
    );
  }
}

// ─── Вкладка 2: Запчасти (Управление и очистка нулей) ──────────────────────

class _PartsTab extends ConsumerWidget {
  const _PartsTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return RefreshIndicator(
      onRefresh: () async => ref.invalidate(supplierBatchesProvider),
      child: ListView(
        padding: const EdgeInsets.all(12),
        children: const [
          _PartsManagementCard(),
        ],
      ),
    );
  }
}

class _PartsManagementCard extends ConsumerStatefulWidget {
  const _PartsManagementCard();

  @override
  ConsumerState<_PartsManagementCard> createState() => _PartsManagementCardState();
}

class _PartsManagementCardState extends ConsumerState<_PartsManagementCard> {
  String? _selectedCode;
  bool _deleting = false;

  Future<void> _confirmAndDelete(List<SupplierBatch> batches) async {
    final code = _selectedCode;
    if (code == null || code.isEmpty) return;

    final batch = batches.where((b) => b.code == code).firstOrNull;
    final label = batch?.label ?? code;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить шаблонные запчасти?'),
        content: Text(
          'Все запчасти с количеством 0 для дефектной ведомости «$label» будут безвозвратно удалены.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Отмена'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: Colors.red.shade700),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Удалить'),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    setState(() => _deleting = true);
    try {
      final response = await apiClient.dio.post(
        '/api/v1/admin/parts/zero-quantity',
        data: {'supplier_code': code},
      );
      final data = response.data as Map<String, dynamic>? ?? {};
      final deletedCount = data['deleted_count'] ?? 0;

      if (!mounted) return;
      setState(() => _selectedCode = null);
      ref.invalidate(supplierBatchesProvider);
      ref.invalidate(inventoryProvider);

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Удалено $deletedCount запчастей с количеством 0'),
          backgroundColor: Colors.green.shade800,
        ),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Ошибка удаления: $e'),
          backgroundColor: Colors.red.shade800,
        ),
      );
    } finally {
      if (mounted) {
        setState(() => _deleting = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final batchesAsync = ref.watch(supplierBatchesProvider);

    return Card(
      color: AppTheme.surfaceColor,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: Colors.amber.withValues(alpha: 0.3)),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(LucideIcons.package, color: Colors.amber.shade400, size: 22),
                const SizedBox(width: 8),
                const Expanded(
                  child: Text(
                    'Управление дефектными ведомостями',
                    style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                  ),
                ),
                IconButton(
                  icon: const Icon(LucideIcons.refresh_cw, size: 18),
                  tooltip: 'Обновить ведомости',
                  onPressed: _deleting
                      ? null
                      : () => ref.invalidate(supplierBatchesProvider),
                ),
              ],
            ),
            const SizedBox(height: 6),
            const Text(
              'Удаление незаполненных шаблонных запчастей (quantity = 0) выбранной дефектной ведомости.',
              style: TextStyle(color: Colors.white70, fontSize: 13),
            ),
            const SizedBox(height: 16),
            batchesAsync.when(
              loading: () => const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (err, _) => Padding(
                padding: const EdgeInsets.symmetric(vertical: 8),
                child: Text(
                  'Не удалось загрузить список ведомостей: $err',
                  style: TextStyle(color: Colors.red.shade300, fontSize: 13),
                ),
              ),
              data: (batches) {
                final validSelected = batches.any((b) => b.code == _selectedCode)
                    ? _selectedCode
                    : null;

                if (batches.isEmpty) {
                  return Container(
                    padding: const EdgeInsets.all(14),
                    decoration: BoxDecoration(
                      color: Colors.white.withValues(alpha: 0.05),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Row(
                      children: [
                        Icon(LucideIcons.circle_alert, size: 18, color: Colors.white54),
                        SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            'Нет дефектных ведомостей с нулевыми запчастями',
                            style: TextStyle(color: Colors.white70, fontSize: 13),
                          ),
                        ),
                      ],
                    ),
                  );
                }

                return Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    _BatchSearchPicker(
                      batches: batches,
                      selectedCode: validSelected,
                      enabled: !_deleting,
                      onSelected: (code) => setState(() => _selectedCode = code),
                    ),
                    const SizedBox(height: 16),
                    FilledButton.icon(
                      style: FilledButton.styleFrom(
                        backgroundColor: Colors.red.shade700,
                        padding: const EdgeInsets.symmetric(vertical: 12),
                      ),
                      onPressed: (_deleting || validSelected == null)
                          ? null
                          : () => _confirmAndDelete(batches),
                      icon: _deleting
                          ? const SizedBox(
                              width: 16,
                              height: 16,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                color: Colors.white,
                              ),
                            )
                          : const Icon(LucideIcons.trash, size: 18),
                      label: Text(
                        _deleting
                            ? 'Удаление...'
                            : 'Удалить запчасти с quantity = 0',
                      ),
                    ),
                  ],
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

class _BatchSearchPicker extends StatelessWidget {
  final List<SupplierBatch> batches;
  final String? selectedCode;
  final ValueChanged<String?> onSelected;
  final bool enabled;

  const _BatchSearchPicker({
    required this.batches,
    required this.selectedCode,
    required this.onSelected,
    this.enabled = true,
  });

  Future<void> _openSearchSheet(BuildContext context) async {
    if (!enabled || batches.isEmpty) return;

    final chosen = await showModalBottomSheet<String>(
      context: context,
      isScrollControlled: true,
      backgroundColor: AppTheme.surfaceColor,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (_) => _BatchOptionsSheet(
        batches: batches,
        selectedCode: selectedCode,
      ),
    );

    if (chosen != null) {
      onSelected(chosen.isEmpty ? null : chosen);
    }
  }

  @override
  Widget build(BuildContext context) {
    final current = batches.where((b) => b.code == selectedCode).firstOrNull;

    return InkWell(
      onTap: enabled ? () => _openSearchSheet(context) : null,
      borderRadius: BorderRadius.circular(8),
      child: InputDecorator(
        decoration: InputDecoration(
          labelText: 'Дефектная ведомость (${batches.length} доступно)',
          contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
          suffixIcon: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (selectedCode != null && enabled)
                IconButton(
                  icon: const Icon(LucideIcons.x, size: 18),
                  tooltip: 'Очистить выбор',
                  onPressed: () => onSelected(null),
                ),
              const Padding(
                padding: EdgeInsets.only(right: 12),
                child: Icon(LucideIcons.chevron_down, size: 18),
              ),
            ],
          ),
        ),
        child: Text(
          current?.label ?? 'Выберите дефектную ведомость...',
          style: TextStyle(
            color: current != null ? Colors.white : Colors.white54,
            fontSize: 13,
          ),
          overflow: TextOverflow.ellipsis,
        ),
      ),
    );
  }
}

class _BatchOptionsSheet extends StatefulWidget {
  final List<SupplierBatch> batches;
  final String? selectedCode;

  const _BatchOptionsSheet({
    required this.batches,
    required this.selectedCode,
  });

  @override
  State<_BatchOptionsSheet> createState() => _BatchOptionsSheetState();
}

class _BatchOptionsSheetState extends State<_BatchOptionsSheet> {
  final _searchCtrl = TextEditingController();
  String _query = '';

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  List<SupplierBatch> get _filtered {
    final q = _query.trim().toLowerCase();
    if (q.isEmpty) return widget.batches;
    return widget.batches.where((b) {
      return b.label.toLowerCase().contains(q) || b.code.toLowerCase().contains(q);
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    final filtered = _filtered;

    return Padding(
      padding: EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom),
      child: SizedBox(
        height: MediaQuery.of(context).size.height * 0.75,
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              child: Row(
                children: [
                  const Icon(LucideIcons.package, size: 20, color: AppTheme.primaryColor),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'Выбор дефектной ведомости (${widget.batches.length})',
                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(LucideIcons.x, size: 20),
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: TextField(
                controller: _searchCtrl,
                autofocus: true,
                style: const TextStyle(fontSize: 14),
                decoration: InputDecoration(
                  hintText: 'Поиск по марке, модели, VIN или коду...',
                  prefixIcon: const Icon(LucideIcons.search, size: 18),
                  suffixIcon: _query.isNotEmpty
                      ? IconButton(
                          icon: const Icon(LucideIcons.x, size: 16),
                          onPressed: () {
                            _searchCtrl.clear();
                            setState(() => _query = '');
                          },
                        )
                      : null,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                ),
                onChanged: (val) => setState(() => _query = val),
              ),
            ),
            const SizedBox(height: 8),
            Expanded(
              child: filtered.isEmpty
                  ? Center(
                      child: Text(
                        _query.isEmpty ? 'Нет ведомостей' : 'Ничего не найдено по запросу «$_query»',
                        style: const TextStyle(color: Colors.white54, fontSize: 13),
                      ),
                    )
                  : ListView.builder(
                      itemCount: filtered.length,
                      itemBuilder: (ctx, idx) {
                        final b = filtered[idx];
                        final isSelected = b.code == widget.selectedCode;
                        return ListTile(
                          leading: CircleAvatar(
                            radius: 18,
                            backgroundColor: isSelected
                                ? AppTheme.primaryColor.withValues(alpha: 0.2)
                                : Colors.white.withValues(alpha: 0.05),
                            child: Icon(
                              LucideIcons.package,
                              size: 16,
                              color: isSelected ? AppTheme.primaryColor : Colors.white70,
                            ),
                          ),
                          title: Text(
                            b.label,
                            style: TextStyle(
                              fontSize: 13,
                              fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                              color: isSelected ? AppTheme.primaryColor : Colors.white,
                            ),
                          ),
                          subtitle: Text(
                            'Код: ${b.code}',
                            style: const TextStyle(fontSize: 11, color: Colors.white54),
                          ),
                          trailing: isSelected
                              ? const Icon(LucideIcons.check, color: AppTheme.primaryColor, size: 18)
                              : null,
                          onTap: () => Navigator.pop(context, b.code),
                        );
                      },
                    ),
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Вкладка 3: Логи активности пользователей ──────────────────────────────

class _ActivityTab extends ConsumerWidget {
  const _ActivityTab();

  Color _actionColor(String action) {
    if (action == 'login') return Colors.greenAccent;
    if (action == 'logout') return Colors.blueGrey;
    if (action.contains('create')) return Colors.blueAccent;
    if (action.contains('update')) return Colors.amberAccent;
    if (action.contains('delete')) return Colors.redAccent;
    return Colors.purpleAccent;
  }

  String _formatDate(String isoString) {
    final dt = DateTime.tryParse(isoString);
    if (dt == null) return isoString;
    return DateFormat('dd.MM.yyyy HH:mm:ss').format(dt.toLocal());
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final filters = ref.watch(activityFilterProvider);
    final activityAsync = ref.watch(userActivityLogsProvider);
    final usersAsync = ref.watch(usersListProvider);

    return Scaffold(
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(userActivityLogsProvider),
        child: ListView(
          padding: const EdgeInsets.all(10),
          children: [
            // Фильтры
            Card(
              color: AppTheme.surfaceColor,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
                side: const BorderSide(color: Colors.white10),
              ),
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Icon(LucideIcons.activity, size: 18, color: AppTheme.primaryColor),
                        const SizedBox(width: 8),
                        const Text('Фильтры активности', style: TextStyle(fontWeight: FontWeight.bold)),
                        const Spacer(),
                        TextButton(
                          onPressed: () {
                            ref.read(activityFilterProvider.notifier).state =
                                const ActivityFilterState();
                          },
                          child: const Text('Сброс'),
                        ),
                      ],
                    ),
                    SwitchListTile(
                      contentPadding: EdgeInsets.zero,
                      title: const Text('Только полезные действия', style: TextStyle(fontSize: 13)),
                      subtitle: const Text('Скрывать навигационный шум', style: TextStyle(fontSize: 11, color: Colors.white54)),
                      value: filters.usefulOnly,
                      onChanged: (val) {
                        ref.read(activityFilterProvider.notifier).state =
                            filters.copyWith(usefulOnly: val);
                      },
                    ),
                    const SizedBox(height: 6),
                    usersAsync.maybeWhen(
                      data: (users) => DropdownButtonFormField<int?>(
                        initialValue: filters.selectedUserId,
                        isExpanded: true,
                        decoration: const InputDecoration(
                          labelText: 'Пользователь',
                          contentPadding: EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                        ),
                        items: [
                          const DropdownMenuItem<int?>(
                            value: null,
                            child: Text('Все пользователи', style: TextStyle(fontSize: 13)),
                          ),
                          ...users.map(
                            (u) => DropdownMenuItem<int?>(
                              value: u.id,
                              child: Text('${u.name} (${u.email})', style: const TextStyle(fontSize: 13)),
                            ),
                          ),
                        ],
                        onChanged: (val) {
                          ref.read(activityFilterProvider.notifier).state =
                              filters.copyWith(selectedUserId: () => val);
                        },
                      ),
                      orElse: () => const SizedBox.shrink(),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 8),
            // Список логов
            activityAsync.when(
              loading: () => const Padding(
                padding: EdgeInsets.symmetric(vertical: 40),
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (e, _) => Center(
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Text('Ошибка загрузки логов активности: $e', style: TextStyle(color: Colors.red.shade300)),
                ),
              ),
              data: (logs) {
                if (logs.isEmpty) {
                  return const Padding(
                    padding: EdgeInsets.symmetric(vertical: 40),
                    child: Center(
                      child: Text('Логи активности отсутствуют', style: TextStyle(color: Colors.white54)),
                    ),
                  );
                }

                return Column(
                  children: logs.map((log) {
                    final actionColor = _actionColor(log.action);
                    return Card(
                      margin: const EdgeInsets.symmetric(vertical: 4),
                      color: AppTheme.surfaceColor,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                        side: const BorderSide(color: Colors.white10),
                      ),
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                  decoration: BoxDecoration(
                                    color: actionColor.withValues(alpha: 0.15),
                                    borderRadius: BorderRadius.circular(4),
                                    border: Border.all(color: actionColor.withValues(alpha: 0.4)),
                                  ),
                                  child: Text(
                                    log.action.replaceAll('_', ' ').toUpperCase(),
                                    style: TextStyle(color: actionColor, fontSize: 10, fontWeight: FontWeight.bold),
                                  ),
                                ),
                                const SizedBox(width: 8),
                                Expanded(
                                  child: Text(
                                    log.userName,
                                    style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                ),
                                Text(
                                  _formatDate(log.createdAt),
                                  style: const TextStyle(color: Colors.white54, fontSize: 11),
                                ),
                              ],
                            ),
                            const SizedBox(height: 6),
                            Text(
                              log.details,
                              style: const TextStyle(fontSize: 13),
                            ),
                            if (log.resourceType.isNotEmpty) ...[
                              const SizedBox(height: 4),
                              Text(
                                'Тип: ${log.resourceType}${log.resourceId != null ? ' (ID: ${log.resourceId})' : ''}',
                                style: const TextStyle(color: Colors.white54, fontSize: 11),
                              ),
                            ],
                          ],
                        ),
                      ),
                    );
                  }).toList(),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Вкладка 4: Система (Статус сервера и логи бэкенда) ───────────────────

class _SystemTab extends ConsumerWidget {
  const _SystemTab();

  Color _levelColor(String level) {
    switch (level.toUpperCase()) {
      case 'ERROR':
        return Colors.redAccent;
      case 'WARN':
      case 'WARNING':
        return Colors.amberAccent;
      default:
        return Colors.blueAccent;
    }
  }

  String _formatDate(String isoString) {
    final dt = DateTime.tryParse(isoString);
    if (dt == null) return isoString;
    return DateFormat('dd.MM.yyyy HH:mm:ss').format(dt.toLocal());
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final statusAsync = ref.watch(serverStatusProvider);
    final logsAsync = ref.watch(serverLogsProvider);

    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(serverStatusProvider);
        ref.invalidate(serverLogsProvider);
      },
      child: ListView(
        padding: const EdgeInsets.all(12),
        children: [
          // Карточка: Статус Сервера
          Card(
            color: AppTheme.surfaceColor,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
              side: const BorderSide(color: Colors.white10),
            ),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Row(
                    children: [
                      Icon(LucideIcons.server, color: AppTheme.primaryColor, size: 20),
                      SizedBox(width: 8),
                      Text('Состояние системы', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                    ],
                  ),
                  const SizedBox(height: 12),
                  statusAsync.when(
                    loading: () => const Center(child: CircularProgressIndicator()),
                    error: (e, _) => Text('Ошибка загрузки статуса: $e', style: TextStyle(color: Colors.red.shade300, fontSize: 12)),
                    data: (info) => Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _infoRow('Сервер:', info.serverStatus.toUpperCase(), Colors.greenAccent),
                        _infoRow('Go Version:', info.goVersion, Colors.white70),
                        _infoRow('ОС / Архитектура:', '${info.os} / ${info.arch}', Colors.white70),
                        const Divider(height: 16, color: Colors.white10),
                        _infoRow('База данных:', info.dbStatus.toUpperCase(), Colors.greenAccent),
                        _infoRow('Всего запчастей в БД:', '${info.totalParts}', AppTheme.primaryColor),
                        _infoRow('Всего пользователей:', '${info.totalUsers}', Colors.white70),
                        if (info.timestamp.isNotEmpty) ...[
                          const SizedBox(height: 6),
                          Text('Обновлено: ${_formatDate(info.timestamp)}', style: const TextStyle(color: Colors.white54, fontSize: 11)),
                        ],
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),

          // Карточка: Клиентские инструменты
          Card(
            color: AppTheme.surfaceColor,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
              side: const BorderSide(color: Colors.white10),
            ),
            child: ListTile(
              leading: const Icon(LucideIcons.file_text, color: AppTheme.primaryColor),
              title: const Text('Журнал ошибок приложения', style: TextStyle(fontWeight: FontWeight.w600)),
              subtitle: const Text('Сетевые сбои клиента и экспорт отчетов', style: TextStyle(fontSize: 12)),
              trailing: const Icon(LucideIcons.chevron_right, color: Colors.white54),
              onTap: () => context.go('/admin/logs'),
            ),
          ),
          const SizedBox(height: 12),

          // Карточка: Логи сервера
          Card(
            color: AppTheme.surfaceColor,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
              side: const BorderSide(color: Colors.white10),
            ),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Row(
                    children: [
                      Icon(LucideIcons.file_text, color: Colors.amberAccent, size: 20),
                      SizedBox(width: 8),
                      Text('Логи сервера (Бэкенд)', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                    ],
                  ),
                  const SizedBox(height: 12),
                  logsAsync.when(
                    loading: () => const Center(child: CircularProgressIndicator()),
                    error: (e, _) => Text('Ошибка загрузки логов сервера: $e', style: TextStyle(color: Colors.red.shade300, fontSize: 12)),
                    data: (logs) {
                      if (logs.isEmpty) {
                        return const Text('Логи отсутствуют', style: TextStyle(color: Colors.white54, fontSize: 13));
                      }
                      return Column(
                        children: logs.map((log) {
                          final lvlColor = _levelColor(log.level);
                          return Container(
                            margin: const EdgeInsets.only(bottom: 8),
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: Colors.white.withValues(alpha: 0.03),
                              borderRadius: BorderRadius.circular(6),
                            ),
                            child: Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                  decoration: BoxDecoration(
                                    color: lvlColor.withValues(alpha: 0.15),
                                    borderRadius: BorderRadius.circular(4),
                                    border: Border.all(color: lvlColor.withValues(alpha: 0.4)),
                                  ),
                                  child: Text(
                                    log.level,
                                    style: TextStyle(color: lvlColor, fontSize: 10, fontWeight: FontWeight.bold),
                                  ),
                                ),
                                const SizedBox(width: 8),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(log.message, style: const TextStyle(fontSize: 12)),
                                      Text(_formatDate(log.timestamp), style: const TextStyle(color: Colors.white54, fontSize: 10)),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          );
                        }).toList(),
                      );
                    },
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 40),
        ],
      ),
    );
  }

  Widget _infoRow(String label, String value, Color valueColor) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: Colors.white70, fontSize: 13)),
          Text(value, style: TextStyle(color: valueColor, fontWeight: FontWeight.bold, fontSize: 13)),
        ],
      ),
    );
  }
}
