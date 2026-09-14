import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../../app/theme.dart';
import '../../../core/services/error_reporter.dart';

class AppLogsScreen extends ConsumerStatefulWidget {
  const AppLogsScreen({super.key});

  @override
  ConsumerState<AppLogsScreen> createState() => _AppLogsScreenState();
}

class _AppLogsScreenState extends ConsumerState<AppLogsScreen> {
  String _selectedFilter = 'all'; // 'all', 'network', 'flutter', 'fatal'
  bool _isSending = false;

  @override
  Widget build(BuildContext context) {
    final reporter = ref.watch(errorReporterProvider);

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (!didPop) {
          if (context.canPop()) {
            context.pop();
          } else {
            context.go('/admin');
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
                context.go('/admin');
              }
            },
          ),
          title: const Text('Журнал ошибок приложения'),
        actions: [
          IconButton(
            icon: const Icon(Icons.copy_rounded),
            tooltip: 'Скопировать текстовый отчет',
            onPressed: () => _copyReport(context, reporter),
          ),
          IconButton(
            icon: _isSending
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                  )
                : const Icon(Icons.cloud_upload_outlined),
            tooltip: 'Отправить логи на сервер',
            onPressed: _isSending ? null : () => _sendLogsToServer(context, reporter),
          ),
          IconButton(
            icon: const Icon(Icons.delete_outline_rounded),
            tooltip: 'Очистить журнал',
            onPressed: () => _confirmClear(context, reporter),
          ),
        ],
      ),
      body: ValueListenableBuilder<List<ClientLogEntry>>(
        valueListenable: reporter.logsNotifier,
        builder: (context, logs, _) {
          final filteredLogs = logs.where((e) {
            if (_selectedFilter == 'all') return true;
            return e.type == _selectedFilter;
          }).toList();

          final networkCount = logs.where((e) => e.type == 'network').length;
          final flutterCount = logs.where((e) => e.type == 'flutter').length;
          final fatalCount = logs.where((e) => e.type == 'fatal').length;

          return Column(
            children: [
              // Информационная карточка устройства и версии
              Container(
                margin: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                decoration: BoxDecoration(
                  color: AppTheme.surfaceColor,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.white10),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.smartphone_rounded, color: AppTheme.primaryColor, size: 28),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Версия: v${reporter.appVersion}+${reporter.buildNumber}',
                            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            reporter.deviceInfo,
                            style: const TextStyle(color: AppTheme.mutedColor, fontSize: 11),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ],
                      ),
                    ),
                    Text(
                      'Всего: ${logs.length}',
                      style: const TextStyle(color: Colors.white70, fontWeight: FontWeight.w600),
                    ),
                  ],
                ),
              ),

              // Фильтры
              SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                child: Row(
                  children: [
                    _buildFilterChip('all', 'Все (${logs.length})'),
                    const SizedBox(width: 8),
                    _buildFilterChip('network', 'Сеть ($networkCount)'),
                    const SizedBox(width: 8),
                    _buildFilterChip('flutter', 'Flutter UI ($flutterCount)'),
                    const SizedBox(width: 8),
                    _buildFilterChip('fatal', 'Фатальные ($fatalCount)'),
                  ],
                ),
              ),
              const Divider(height: 16),

              // Список ошибок
              Expanded(
                child: filteredLogs.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.check_circle_outline_rounded,
                                size: 64, color: Colors.green.shade400),
                            const SizedBox(height: 16),
                            const Text(
                              'Ошибок не зафиксировано',
                              style: TextStyle(fontSize: 16, color: Colors.white70),
                            ),
                            const SizedBox(height: 4),
                            const Text(
                              'Приложение работает стабильно',
                              style: TextStyle(fontSize: 12, color: AppTheme.mutedColor),
                            ),
                          ],
                        ),
                      )
                    : ListView.builder(
                        padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                        itemCount: filteredLogs.length,
                        itemBuilder: (context, index) {
                          final log = filteredLogs[index];
                          return _ErrorCard(log: log);
                        },
                      ),
              ),
            ],
          );
        },
      ),
    ),
  );
}

  Widget _buildFilterChip(String filterId, String label) {
    final isSelected = _selectedFilter == filterId;
    return FilterChip(
      selected: isSelected,
      label: Text(label),
      onSelected: (_) => setState(() => _selectedFilter = filterId),
      selectedColor: AppTheme.primaryColor.withValues(alpha: 0.25),
      checkmarkColor: AppTheme.primaryColor,
      backgroundColor: AppTheme.surfaceColor,
      labelStyle: TextStyle(
        color: isSelected ? Colors.white : Colors.white70,
        fontSize: 12,
        fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
      ),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(20),
        side: BorderSide(
          color: isSelected ? AppTheme.primaryColor : Colors.white12,
        ),
      ),
    );
  }

  Future<void> _copyReport(BuildContext context, ErrorReporter reporter) async {
    final text = reporter.exportReportText();
    await Clipboard.setData(ClipboardData(text: text));
    if (!context.mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Отчет об ошибках скопирован в буфер обмена'),
        backgroundColor: Colors.green,
        duration: Duration(seconds: 2),
      ),
    );
  }

  Future<void> _sendLogsToServer(BuildContext context, ErrorReporter reporter) async {
    setState(() => _isSending = true);
    final scaffoldMessenger = ScaffoldMessenger.of(context);

    try {
      final success = await reporter.flush();
      if (!context.mounted) return;

      if (success) {
        scaffoldMessenger.showSnackBar(
          const SnackBar(
            content: Text('Логи успешно отправлены на сервер'),
            backgroundColor: Colors.green,
          ),
        );
      } else {
        scaffoldMessenger.showSnackBar(
          const SnackBar(
            content: Text('Не удалось доставить логи на сервер. Проверьте интернет.'),
            backgroundColor: Colors.redAccent,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isSending = false);
    }
  }

  Future<void> _confirmClear(BuildContext context, ErrorReporter reporter) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Очистить журнал?'),
        content: const Text('Все зафиксированные локальные ошибки будут удалены.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Отмена'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: FilledButton.styleFrom(backgroundColor: Colors.redAccent),
            child: const Text('Очистить'),
          ),
        ],
      ),
    );

    if (confirm == true) {
      await reporter.clearLogs();
      if (!context.mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Журнал ошибок очищен')),
      );
    }
  }
}

class _ErrorCard extends StatefulWidget {
  final ClientLogEntry log;
  const _ErrorCard({required this.log});

  @override
  State<_ErrorCard> createState() => _ErrorCardState();
}

class _ErrorCardState extends State<_ErrorCard> {
  bool _expanded = false;

  Color _badgeColor(ClientLogEntry log) {
    if (log.type == 'fatal') return Colors.red.shade700;
    if (log.type == 'flutter') return Colors.purple.shade700;
    if (log.statusCode != null) {
      if (log.statusCode! >= 500) return Colors.red.shade700;
      if (log.statusCode! == 401 || log.statusCode! == 403) return Colors.orange.shade800;
      if (log.statusCode! >= 400) return Colors.amber.shade800;
    }
    return Colors.blueGrey.shade700;
  }

  String _badgeText(ClientLogEntry log) {
    if (log.statusCode != null) return '${log.statusCode}';
    if (log.type == 'fatal') return 'CRASH';
    if (log.type == 'flutter') return 'UI';
    return log.type.toUpperCase();
  }

  @override
  Widget build(BuildContext context) {
    final log = widget.log;
    final timeStr = DateFormat('HH:mm:ss dd.MM').format(log.timestamp.toLocal());
    final badgeColor = _badgeColor(log);

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: _expanded ? badgeColor.withValues(alpha: 0.5) : Colors.white10,
        ),
      ),
      color: AppTheme.surfaceColor,
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () => setState(() => _expanded = !_expanded),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                    decoration: BoxDecoration(
                      color: badgeColor,
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      _badgeText(log),
                      style: const TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  if (log.method != null) ...[
                    Text(
                      log.method!,
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 11,
                        color: Colors.white70,
                      ),
                    ),
                    const SizedBox(width: 6),
                  ],
                  Expanded(
                    child: Text(
                      log.endpoint ?? log.type,
                      style: const TextStyle(
                        fontSize: 12,
                        color: Colors.white60,
                        fontFamily: 'monospace',
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  Text(
                    timeStr,
                    style: const TextStyle(fontSize: 11, color: AppTheme.mutedColor),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Text(
                log.message,
                style: const TextStyle(fontSize: 13, height: 1.3),
                maxLines: _expanded ? null : 2,
                overflow: _expanded ? TextOverflow.visible : TextOverflow.ellipsis,
              ),

              if (_expanded) ...[
                const SizedBox(height: 12),
                if (log.responseBody != null && log.responseBody!.isNotEmpty) ...[
                  const Text(
                    'Ответ сервера:',
                    style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.amberAccent),
                  ),
                  const SizedBox(height: 4),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: Colors.black45,
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: SelectableText(
                      log.responseBody!,
                      style: const TextStyle(fontSize: 11, fontFamily: 'monospace', color: Colors.white70),
                    ),
                  ),
                  const SizedBox(height: 8),
                ],
                if (log.stackTrace != null && log.stackTrace!.isNotEmpty) ...[
                  const Text(
                    'Стек вызовов:',
                    style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.white70),
                  ),
                  const SizedBox(height: 4),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: Colors.black45,
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: SelectableText(
                      log.stackTrace!,
                      style: const TextStyle(fontSize: 10, fontFamily: 'monospace', color: Colors.white60),
                    ),
                  ),
                ],
              ],
            ],
          ),
        ),
      ),
    );
  }
}
