import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../app/theme.dart';
import '../providers/messaging_providers.dart';

class ChatSettingsTab extends ConsumerStatefulWidget {
  const ChatSettingsTab({super.key});

  @override
  ConsumerState<ChatSettingsTab> createState() => _ChatSettingsTabState();
}

class _ChatSettingsTabState extends ConsumerState<ChatSettingsTab> {
  late TextEditingController _intervalController;

  @override
  void initState() {
    super.initState();
    final settings = ref.read(chatSettingsProvider);
    _intervalController = TextEditingController(text: '${settings.refreshInterval}');
  }

  @override
  void dispose() {
    _intervalController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final settings = ref.watch(chatSettingsProvider);

    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: AppTheme.primaryColor.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(14),
              ),
              child: const Icon(LucideIcons.settings, color: AppTheme.primaryColor, size: 24),
            ),
            const SizedBox(width: 14),
            const Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Настройки сообщений',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
                Text(
                  'Параметры синхронизации и уведомлений',
                  style: TextStyle(
                    fontSize: 12,
                    color: AppTheme.mutedColor,
                  ),
                ),
              ],
            ),
          ],
        ),
        const SizedBox(height: 24),

        // Auto Refresh Switch
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: AppTheme.cardColor,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: AppTheme.borderColor),
          ),
          child: Column(
            children: [
              SwitchListTile(
                contentPadding: EdgeInsets.zero,
                activeThumbColor: AppTheme.primaryColor,
                title: const Text(
                  'Автообновление Drom сообщений',
                  style: TextStyle(fontWeight: FontWeight.w600, color: Colors.white),
                ),
                subtitle: const Text(
                  'Автоматически опрашивать новые диалоги и входящие сообщения',
                  style: TextStyle(color: AppTheme.mutedColor, fontSize: 12),
                ),
                value: settings.enableAutoRefresh,
                onChanged: (val) {
                  ref.read(chatSettingsProvider.notifier).setAutoRefresh(val);
                },
              ),
              const Divider(color: AppTheme.borderColor, height: 24),

              // Refresh Interval
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Интервал обновления',
                          style: TextStyle(fontWeight: FontWeight.w600, color: Colors.white),
                        ),
                        Text(
                          settings.enableAutoRefresh
                              ? 'Частота опроса в секундах (от 10 до 300)'
                              : 'Отключено (включите автообновление)',
                          style: const TextStyle(color: AppTheme.mutedColor, fontSize: 12),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 16),
                  SizedBox(
                    width: 90,
                    child: TextField(
                      controller: _intervalController,
                      enabled: settings.enableAutoRefresh,
                      keyboardType: TextInputType.number,
                      textAlign: TextAlign.center,
                      decoration: const InputDecoration(
                        suffixText: 'сек',
                        contentPadding: EdgeInsets.symmetric(horizontal: 10, vertical: 10),
                      ),
                      onChanged: (val) {
                        final parsed = int.tryParse(val);
                        if (parsed != null && parsed >= 5 && parsed <= 600) {
                          ref.read(chatSettingsProvider.notifier).setRefreshInterval(parsed);
                        }
                      },
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),

        const SizedBox(height: 16),

        // Notifications Switch
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: AppTheme.cardColor,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: AppTheme.borderColor),
          ),
          child: SwitchListTile(
            contentPadding: EdgeInsets.zero,
            activeThumbColor: AppTheme.primaryColor,
            secondary: Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: AppTheme.primaryColor.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(10),
              ),
              child: const Icon(LucideIcons.bell, color: AppTheme.primaryColor, size: 20),
            ),
            title: const Text(
              'Уведомления',
              style: TextStyle(fontWeight: FontWeight.w600, color: Colors.white),
            ),
            subtitle: const Text(
              'Показывать уведомления о новых сообщениях в чатах',
              style: TextStyle(color: AppTheme.mutedColor, fontSize: 12),
            ),
            value: settings.enableNotifications,
            onChanged: (val) {
              ref.read(chatSettingsProvider.notifier).setNotifications(val);
            },
          ),
        ),

        const SizedBox(height: 24),

        // Reset Settings Button
        OutlinedButton.icon(
          style: OutlinedButton.styleFrom(
            padding: const EdgeInsets.symmetric(vertical: 14),
            side: const BorderSide(color: AppTheme.borderColor),
          ),
          icon: const Icon(LucideIcons.rotate_ccw, size: 18),
          label: const Text('Сбросить настройки к стандартным'),
          onPressed: () {
            ref.read(chatSettingsProvider.notifier).resetToDefaults();
            final defaultSettings = ref.read(chatSettingsProvider);
            _intervalController.text = '${defaultSettings.refreshInterval}';
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Настройки сброшены')),
            );
          },
        ),
      ],
    );
  }
}
