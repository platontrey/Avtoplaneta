import 'dart:io';
import 'package:dio/dio.dart';

/// Базовый класс для всех ошибок приложения с понятным описанием для пользователя.
abstract class AppFailure implements Exception {
  final String message;
  final int? statusCode;
  final Object? originalError;

  const AppFailure(
    this.message, {
    this.statusCode,
    this.originalError,
  });

  @override
  String toString() => message;

  /// Фабричный метод для преобразования любого исключения в типизированный [AppFailure].
  factory AppFailure.from(Object error) {
    if (error is AppFailure) {
      return error;
    }

    if (error is DioException) {
      return AppFailure.fromDioException(error);
    }

    if (error is SocketException) {
      return const NetworkFailure(
        'Отсутствует подключение к интернету. Проверьте сеть и повторите попытку.',
      );
    }

    if (error is FormatException) {
      return ValidationFailure(
        'Ошибка формата данных: ${error.message}',
        originalError: error,
      );
    }

    return UnknownFailure(
      error.toString().replaceFirst('Exception: ', ''),
      originalError: error,
    );
  }

  /// Парсинг исключений сетевого клиента Dio.
  factory AppFailure.fromDioException(DioException dioError) {
    switch (dioError.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return NetworkFailure(
          'Превышено время ожидания ответа сервера. Попробуйте снова.',
          originalError: dioError,
        );

      case DioExceptionType.connectionError:
        return NetworkFailure(
          'Не удалось установить соединение с сервером. Проверьте интернет или настройки сети.',
          originalError: dioError,
        );

      case DioExceptionType.cancel:
        return const NetworkFailure('Запрос был отменен пользователем.');

      case DioExceptionType.badCertificate:
        return const NetworkFailure(
          'Ошибка безопасности SSL-сертификата сервера.',
        );

      case DioExceptionType.badResponse:
        final status = dioError.response?.statusCode;
        final responseData = dioError.response?.data;
        String? serverMessage;

        if (responseData is Map) {
          serverMessage = (responseData['error'] ??
                  responseData['message'] ??
                  responseData['detail'])
              ?.toString();
        } else if (responseData is String && responseData.isNotEmpty) {
          if (responseData.trim().toLowerCase().startsWith('<!doctype html') ||
              responseData.trim().toLowerCase().startsWith('<html')) {
            serverMessage = 'Сервер вернул HTML-страницу вместо ожидаемого API JSON ответа.';
          } else if (responseData.length < 200) {
            serverMessage = responseData;
          }
        }

        if (status == 401 || status == 403) {
          return AuthFailure(
            serverMessage ?? 'Сессия устарела или недостаточно прав. Требуется повторный вход.',
            statusCode: status,
            originalError: dioError,
          );
        }

        if (status == 404) {
          return NotFoundFailure(
            serverMessage ?? 'Запрашиваемый ресурс или запись не найдены.',
            statusCode: status,
            originalError: dioError,
          );
        }

        if (status == 409) {
          return ValidationFailure(
            serverMessage ?? 'Конфликт данных: запись с такими параметрами уже существует (например, дубликат телефона).',
            statusCode: status,
            originalError: dioError,
          );
        }

        if (status != null && status >= 400 && status < 500) {
          return ValidationFailure(
            serverMessage ?? 'Некорректный запрос к серверу (ошибка $status).',
            statusCode: status,
            originalError: dioError,
          );
        }

        if (status != null && status >= 500) {
          return ServerFailure(
            serverMessage ?? 'Внутренняя ошибка сервера ($status). Пожалуйста, повторите позже.',
            statusCode: status,
            originalError: dioError,
          );
        }

        return UnknownFailure(
          serverMessage ?? 'Непредвиденная ошибка ответа сервера (${dioError.message}).',
          statusCode: status,
          originalError: dioError,
        );

      case DioExceptionType.unknown:
        if (dioError.error is SocketException) {
          return const NetworkFailure(
            'Отсутствует подключение к интернету.',
          );
        }
        return UnknownFailure(
          dioError.message ?? 'Произошла непредвиденная ошибка сети.',
          originalError: dioError,
        );
    }
  }
}

/// Ошибка сетевого соединения или таймаута.
class NetworkFailure extends AppFailure {
  const NetworkFailure(
    super.message, {
    super.statusCode,
    super.originalError,
  });
}

/// Ошибка авторизации или недостатка прав.
class AuthFailure extends AppFailure {
  const AuthFailure(
    super.message, {
    super.statusCode,
    super.originalError,
  });
}

/// Ошибка валидации параметров или конфликта данных (400, 409, 422).
class ValidationFailure extends AppFailure {
  const ValidationFailure(
    super.message, {
    super.statusCode,
    super.originalError,
  });
}

/// Ошибка отсутствия записи (404 Not Found).
class NotFoundFailure extends AppFailure {
  const NotFoundFailure(
    super.message, {
    super.statusCode,
    super.originalError,
  });
}

/// Ошибка бэкенд-сервиса (5xx).
class ServerFailure extends AppFailure {
  const ServerFailure(
    super.message, {
    super.statusCode,
    super.originalError,
  });
}

/// Неизвестная или непредвиденная ошибка.
class UnknownFailure extends AppFailure {
  const UnknownFailure(
    super.message, {
    super.statusCode,
    super.originalError,
  });
}
