import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:open_filex/open_filex.dart';

import '../../../app/theme.dart';
import '../../../core/services/update_service.dart';

class UpdateDialog extends StatefulWidget {
  final AppUpdateInfo info;
  final UpdateService updateService;

  const UpdateDialog({
    super.key,
    required this.info,
    required this.updateService,
  });

  /// Удобный метод для показа диалога
  static Future<void> show({
    required BuildContext context,
    required AppUpdateInfo info,
    required UpdateService updateService,
  }) {
    return showDialog<void>(
      context: context,
      barrierDismissible: !info.forceUpdate,
      builder: (_) => UpdateDialog(
        info: info,
        updateService: updateService,
      ),
    );
  }

  @override
  State<UpdateDialog> createState() => _UpdateDialogState();
}

class _UpdateDialogState extends State<UpdateDialog> {
  bool _isDownloading = false;
  double _progress = 0.0;
  int _receivedBytes = 0;
  int _totalBytes = 0;
  String? _errorMessage;
  CancelToken? _cancelToken;

  @override
  void dispose() {
    _cancelToken?.cancel('Диалог закрыт пользователем');
    super.dispose();
  }

  String _formatBytes(int bytes) {
    if (bytes <= 0) return '';
    final mb = bytes / (1024 * 1024);
    return '${mb.toStringAsFixed(1)} МБ';
  }

  Future<void> _startDownload() async {
    setState(() {
      _isDownloading = true;
      _errorMessage = null;
      _progress = 0.0;
      _receivedBytes = 0;
      _totalBytes = 0;
      _cancelToken = CancelToken();
    });

    try {
      final result = await widget.updateService.downloadAndInstall(
        info: widget.info,
        cancelToken: _cancelToken,
        onProgress: (progress, received, total) {
          if (mounted) {
            setState(() {
              _progress = progress;
              _receivedBytes = received;
              _totalBytes = total;
            });
          }
        },
      );

      if (!mounted) return;

      if (result.type != ResultType.done) {
        setState(() {
          _errorMessage = result.message.isNotEmpty
              ? 'Ошибка запуска установщика: ${result.message}'
              : 'Не удалось открыть файл установки. Проверьте разрешение на установку приложений.';
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = _formatErrorMessage(e);
        });
      }
    } finally {
      if (mounted) {
        setState(() => _isDownloading = false);
      }
    }
  }

  String _formatErrorMessage(dynamic error) {
    if (error is DioException) {
      if (CancelToken.isCancel(error)) {
        return 'Ошибка загрузки: операция отменена';
      }
      final underlying = error.error?.toString() ?? '';
      if (underlying.contains('Connection closed') ||
          underlying.contains('Software caused connection abort') ||
          underlying.contains('SocketException') ||
          underlying.contains('Broken pipe')) {
        return 'Ошибка загрузки: связь с сервером прервана. Нажмите «Повторить загрузку».';
      }
      if (error.type == DioExceptionType.connectionTimeout ||
          error.type == DioExceptionType.receiveTimeout ||
          error.type == DioExceptionType.sendTimeout) {
        return 'Ошибка загрузки: превышено время ожидания ответа сервера. Нажмите «Повторить загрузку».';
      }
      if (error.response?.statusCode == 404) {
        return 'Ошибка загрузки: файл обновления не найден на сервере (код 404).';
      }
      return 'Ошибка загрузки: сетевой сбой. Нажмите «Повторить загрузку».';
    }

    final str = error.toString();
    if (str.contains('Connection closed') || str.contains('SocketException')) {
      return 'Ошибка загрузки: соединение с сервером прервано. Нажмите «Повторить загрузку».';
    }
    return 'Ошибка загрузки: $str';
  }

  @override
  Widget build(BuildContext context) {
    final info = widget.info;

    return PopScope(
      canPop: !info.forceUpdate && !_isDownloading,
      child: AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        backgroundColor: AppTheme.surfaceColor,
        titlePadding: const EdgeInsets.fromLTRB(24, 24, 24, 12),
        contentPadding: const EdgeInsets.fromLTRB(24, 0, 24, 16),
        actionsPadding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: AppTheme.primaryColor.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(
                Icons.system_update_rounded,
                color: AppTheme.primaryColor,
                size: 28,
              ),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Доступно обновление',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                  ),
                  Text(
                    'Версия ${info.version}${info.buildNumber > 0 ? ' (сборка ${info.buildNumber})' : ''}',
                    style: const TextStyle(
                      fontSize: 13,
                      color: AppTheme.mutedColor,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              if (info.forceUpdate) ...[
                Container(
                  margin: const EdgeInsets.only(bottom: 12),
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: Colors.amber.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.amber.shade700, width: 1),
                  ),
                  child: const Row(
                    children: [
                      Icon(Icons.warning_amber_rounded, color: Colors.amber, size: 20),
                      SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          'Это обязательное обновление для продолжения работы.',
                          style: TextStyle(fontSize: 12, color: Colors.amber),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
              if (info.changelog.isNotEmpty) ...[
                const Text(
                  'Что нового:',
                  style: TextStyle(
                    fontWeight: FontWeight.w600,
                    fontSize: 14,
                  ),
                ),
                const SizedBox(height: 6),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: AppTheme.cardColor,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Text(
                    info.changelog,
                    style: const TextStyle(fontSize: 13, height: 1.4),
                  ),
                ),
                const SizedBox(height: 12),
              ],
              if (_isDownloading) ...[
                const SizedBox(height: 8),
                ClipRRect(
                  borderRadius: BorderRadius.circular(8),
                  child: LinearProgressIndicator(
                    value: _progress > 0 ? _progress : null,
                    minHeight: 8,
                    backgroundColor: AppTheme.cardColor,
                    valueColor: const AlwaysStoppedAnimation(AppTheme.primaryColor),
                  ),
                ),
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      _progress > 0
                          ? '${(_progress * 100).toStringAsFixed(0)}%'
                          : 'Загрузка...',
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: AppTheme.primaryColor,
                      ),
                    ),
                    if (_totalBytes > 0)
                      Text(
                        '${_formatBytes(_receivedBytes)} / ${_formatBytes(_totalBytes)}',
                        style: const TextStyle(
                          fontSize: 12,
                          color: AppTheme.mutedColor,
                        ),
                      ),
                  ],
                ),
              ],
              if (_errorMessage != null) ...[
                const SizedBox(height: 8),
                Text(
                  _errorMessage!,
                  style: const TextStyle(
                    color: Colors.redAccent,
                    fontSize: 12,
                  ),
                ),
              ],
            ],
          ),
        ),
        actions: [
          if (!info.forceUpdate && !_isDownloading)
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('Позже', style: TextStyle(color: AppTheme.mutedColor)),
            ),
          if (!_isDownloading)
            FilledButton.icon(
              onPressed: _startDownload,
              icon: Icon(_errorMessage != null ? Icons.refresh_rounded : Icons.download_rounded, size: 18),
              label: Text(_errorMessage != null ? 'Повторить загрузку' : 'Обновить'),
            ),
        ],
      ),
    );
  }
}
