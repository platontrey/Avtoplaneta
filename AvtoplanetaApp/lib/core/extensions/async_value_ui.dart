import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../errors/failure.dart';

/// Расширение над [AsyncValue] для лаконичной и единообразной обработки UI-состояний.
extension AsyncValueUI<T> on AsyncValue<T> {
  /// Отображает SnackBar с текстом ошибки, если [AsyncValue] находится в состоянии ошибки и контекст валиден.
  void showSnackbarOnError(BuildContext context) {
    if (!hasError || isLoading) return;

    final failure = AppFailure.from(error!);
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          content: Text(failure.message),
          backgroundColor: Colors.red.shade700,
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(10),
          ),
          action: SnackBarAction(
            label: 'OK',
            textColor: Colors.white,
            onPressed: () {
              ScaffoldMessenger.of(context).hideCurrentSnackBar();
            },
          ),
        ),
      );
  }

  /// Удобный хелпер рендеринга состояния с готовыми красивыми заглушками для загрузки и ошибок.
  Widget toWidget({
    required Widget Function(T data) data,
    Widget Function()? loading,
    Widget Function(AppFailure failure)? error,
    VoidCallback? onRetry,
  }) {
    return when(
      data: data,
      loading: () =>
          loading?.call() ??
          const Center(
            child: CircularProgressIndicator(),
          ),
      error: (err, stack) {
        final failure = AppFailure.from(err);
        if (error != null) {
          return error(failure);
        }

        return Center(
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  Icons.error_outline_rounded,
                  size: 48,
                  color: Colors.red.shade400,
                ),
                const SizedBox(height: 16),
                Text(
                  failure.message,
                  textAlign: TextAlign.center,
                  style: const TextStyle(
                    fontSize: 15,
                    color: Colors.black87,
                    height: 1.4,
                  ),
                ),
                if (onRetry != null) ...[
                  const SizedBox(height: 16),
                  ElevatedButton.icon(
                    onPressed: onRetry,
                    icon: const Icon(Icons.refresh_rounded, size: 18),
                    label: const Text('Повторить попытку'),
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 20,
                        vertical: 12,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        );
      },
    );
  }
}
