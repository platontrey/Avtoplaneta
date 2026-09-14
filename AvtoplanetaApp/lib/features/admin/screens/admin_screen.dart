import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../app/theme.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/user.dart';
import '../../../core/services/update_service.dart';
import '../../updater/widgets/update_dialog.dart';

final usersListProvider = FutureProvider<List<User>>((ref) async {
  final response = await apiClient.dio.get('/admin/users');
  final data = response.data as Map<String, dynamic>;
  final list = data['users'] as List<dynamic>? ?? [];
  return list.map((e) => User.fromJson(e as Map<String, dynamic>)).toList();
});

class AdminScreen extends ConsumerWidget {
  const AdminScreen({super.key});

  Future<void> _checkAppUpdates(BuildContext context, WidgetRef ref) async {
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
  Widget build(BuildContext context, WidgetRef ref) {
    final usersAsync = ref.watch(usersListProvider);

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
            icon: const Icon(Icons.arrow_back),
            tooltip: 'Назад',
            onPressed: () {
              if (context.canPop()) {
                context.pop();
              } else {
                context.go('/inventory');
              }
            },
          ),
          title: const Text('Администрирование'),
          actions: [
          IconButton(
            icon: const Icon(Icons.receipt_long_rounded),
            tooltip: 'Журнал ошибок',
            onPressed: () => context.go('/admin/logs'),
          ),
          IconButton(
            icon: const Icon(Icons.system_update_rounded),
            tooltip: 'Проверить обновления',
            onPressed: () => _checkAppUpdates(context, ref),
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Обновить список',
            onPressed: () => ref.invalidate(usersListProvider),
          ),
        ],
      ),
      body: usersAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.error_outline_rounded, size: 64, color: Colors.orange.shade400),
                const SizedBox(height: 16),
                const Text(
                  'Не удалось загрузить пользователей',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),
                Text(
                  _formatError(e),
                  style: const TextStyle(color: Colors.white70, fontSize: 13),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 24),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    FilledButton.icon(
                      onPressed: () => ref.invalidate(usersListProvider),
                      icon: const Icon(Icons.refresh_rounded),
                      label: const Text('Повторить'),
                    ),
                    const SizedBox(width: 12),
                    OutlinedButton.icon(
                      onPressed: () => context.go('/admin/logs'),
                      icon: const Icon(Icons.receipt_long_rounded),
                      label: const Text('Журнал ошибок'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
        data: (users) => ListView(
          padding: const EdgeInsets.all(8),
          children: [
            Card(
              margin: const EdgeInsets.symmetric(horizontal: 4, vertical: 6),
              color: AppTheme.surfaceColor,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
                side: const BorderSide(color: Colors.white10),
              ),
              child: ListTile(
                leading: const Icon(Icons.receipt_long_rounded, color: AppTheme.primaryColor),
                title: const Text('Журнал ошибок приложения', style: TextStyle(fontWeight: FontWeight.w600)),
                subtitle: const Text('Логи сетевых сбоев, крашей и экспорт отчета', style: TextStyle(fontSize: 12)),
                trailing: const Icon(Icons.chevron_right, color: Colors.white54),
                onTap: () => context.go('/admin/logs'),
              ),
            ),
            const SizedBox(height: 8),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              child: Text(
                'Пользователи (${users.length})',
                style: Theme.of(
                  context,
                ).textTheme.titleMedium?.copyWith(color: Colors.white70),
              ),
            ),
            ...users.map(
              (u) => _UserTile(
                user: u,
                onDeleted: () => ref.invalidate(usersListProvider),
              ),
            ),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showCreateUser(context, ref),
        child: const Icon(Icons.person_add),
      ),
    ),
  );
}

  String _formatError(dynamic e) {
    if (e is DioException) {
      final code = e.response?.statusCode;
      if (code == 401) {
        return 'Ошибка 401: Сессия устарела или требуется повторная авторизация.';
      }
      if (code == 403) {
        return 'Ошибка 403: Недостаточно прав для просмотра списка пользователей.';
      }
      return 'Сетевой сбой ($code): ${e.message}';
    }
    return e.toString();
  }

  void _showCreateUser(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (_) =>
          _CreateUserDialog(onCreated: () => ref.invalidate(usersListProvider)),
    );
  }
}

class _UserTile extends StatelessWidget {
  final User user;
  final VoidCallback onDeleted;
  const _UserTile({required this.user, required this.onDeleted});

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

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 3, horizontal: 4),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: _roleColor(user.role).withValues(alpha: 0.2),
          child: Text(
            user.initials.isNotEmpty
                ? user.initials
                : user.name.isNotEmpty
                ? user.name[0]
                : '?',
            style: TextStyle(
              color: _roleColor(user.role),
              fontWeight: FontWeight.bold,
            ),
          ),
        ),
        title: Text(user.name, style: const TextStyle(color: Colors.white)),
        subtitle: Text(
          user.email,
          style: const TextStyle(color: Colors.white54, fontSize: 12),
        ),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
              decoration: BoxDecoration(
                color: _roleColor(user.role).withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(6),
              ),
              child: Text(
                user.role,
                style: TextStyle(color: _roleColor(user.role), fontSize: 11),
              ),
            ),
            const SizedBox(width: 4),
            IconButton(
              icon: const Icon(
                Icons.delete_outline,
                color: Colors.red,
                size: 20,
              ),
              onPressed: () => _confirmDelete(context),
            ),
          ],
        ),
      ),
    );
  }

  void _confirmDelete(BuildContext context) {
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Удалить пользователя?'),
        content: Text('${user.name} (${user.email})'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Отмена'),
          ),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              try {
                await apiClient.dio.delete('/admin/users/${user.id}');
                onDeleted();
              } catch (_) {}
            },
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Удалить'),
          ),
        ],
      ),
    );
  }
}

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
  String _role = 'operator';
  bool _loading = false;

  @override
  void dispose() {
    _nameCtrl.dispose();
    _emailCtrl.dispose();
    _passCtrl.dispose();
    super.dispose();
  }

  Future<void> _create() async {
    if (_nameCtrl.text.isEmpty ||
        _emailCtrl.text.isEmpty ||
        _passCtrl.text.isEmpty) {
      return;
    }
    setState(() => _loading = true);
    try {
      await apiClient.dio.post(
        '/admin/users',
        data: {
          'name': _nameCtrl.text.trim(),
          'email': _emailCtrl.text.trim(),
          'password': _passCtrl.text,
          'role': _role,
        },
      );
      widget.onCreated();
      if (mounted) {
        Navigator.pop(context);
      }
    } catch (_) {
      if (mounted) {
        setState(() => _loading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Создать пользователя'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          TextField(
            controller: _nameCtrl,
            decoration: const InputDecoration(labelText: 'Имя'),
          ),
          const SizedBox(height: 8),
          TextField(
            controller: _emailCtrl,
            decoration: const InputDecoration(labelText: 'Email'),
            keyboardType: TextInputType.emailAddress,
          ),
          const SizedBox(height: 8),
          TextField(
            controller: _passCtrl,
            decoration: const InputDecoration(labelText: 'Пароль'),
            obscureText: true,
          ),
          const SizedBox(height: 8),
          DropdownButtonFormField<String>(
            initialValue: _role,
            items: const [
              DropdownMenuItem(value: 'operator', child: Text('Оператор')),
              DropdownMenuItem(value: 'manager', child: Text('Менеджер')),
              DropdownMenuItem(value: 'admin', child: Text('Администратор')),
            ],
            onChanged: (v) => setState(() => _role = v!),
            decoration: const InputDecoration(labelText: 'Роль'),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Отмена'),
        ),
        FilledButton(
          onPressed: _loading ? null : _create,
          child: const Text('Создать'),
        ),
      ],
    );
  }
}
