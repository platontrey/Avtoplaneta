/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import React, { Component, type ErrorInfo, type ReactNode } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Button } from './ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
  // Счётчик перемонтирований для тихого восстановления от транзиентных ошибок
  // (например, конфликт автопереводчика с React-реконсиляцией: removeChild/insertBefore).
  remountKey: number;
  attempts: number;
  firstAttemptAt: number;
}

// Сигнатуры ошибок, вызванных вмешательством расширений-переводчиков в DOM.
// Такие ошибки транзиентны — перемонтирование поддерева восстанавливает
// согласованное состояние DOM без необходимости показывать экран ошибки.
const TRANSLATE_GLITCH_RE = /removeChild|insertBefore|appendChild|NotFoundError|failed to execute a 'removeChild'/i;

const MAX_QUIET_ATTEMPTS = 3;
const ATTEMPT_WINDOW_MS = 5000;

class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    remountKey: 0,
    attempts: 0,
    firstAttemptAt: 0,
  };

  public static getDerivedStateFromError(error: Error): Partial<State> {
    // Для ошибок автопереводчика НЕ устанавливаем hasError — вместо этого
    // тихо запрашиваем перемонтирование поддерева через remountKey.
    if (TRANSLATE_GLITCH_RE.test(error.message)) {
      return { hasError: false };
    }
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Тихое восстановление для транзиентных ошибок автопереводчика.
    if (TRANSLATE_GLITCH_RE.test(error.message)) {
      const now = Date.now();
      const windowExpired =
        now - this.state.firstAttemptAt > ATTEMPT_WINDOW_MS;

      const attempts = windowExpired ? 1 : this.state.attempts + 1;
      const firstAttemptAt = windowExpired ? now : this.state.firstAttemptAt;

      if (attempts <= MAX_QUIET_ATTEMPTS) {
        // Сбрасываем ошибку и перемонтируем детей, чтобы DOM пришёл в согласованное состояние.
        this.setState((prev) => ({
          hasError: false,
          error: undefined,
          errorInfo: undefined,
          remountKey: prev.remountKey + 1,
          attempts,
          firstAttemptAt,
        }));
        console.warn('ErrorBoundary: восстановление после транзиентной DOM-ошибки (возможно, автоперевод).', error);
        return;
      }

      // Превышен лимит попыток в окне — показываем полноценный экран ошибки.
      console.error('ErrorBoundary: предел попыток тихого восстановления исчерпан.', error, errorInfo);
    } else {
      console.error('ErrorBoundary caught an error:', error, errorInfo);
    }

    this.setState({
      hasError: true,
      error,
      errorInfo,
    });
  }

  private handleRetry = () => {
    this.setState({
      hasError: false,
      error: undefined,
      errorInfo: undefined,
      attempts: 0,
      firstAttemptAt: 0,
    });
  };

  public render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="min-h-screen flex items-center justify-center p-4">
          <Card className="w-full max-w-md">
            <CardHeader className="text-center">
              <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-100">
                <AlertTriangle className="h-6 w-6 text-red-600" />
              </div>
              <CardTitle className="text-red-900">Что-то пошло не так</CardTitle>
              <CardDescription>
                Произошла непредвиденная ошибка. Попробуйте перезагрузить страницу.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <Button onClick={this.handleRetry} className="w-full">
                <RefreshCw className="mr-2 h-4 w-4" />
                Попробовать снова
              </Button>
              <details className="text-sm text-gray-600">
                <summary className="cursor-pointer hover:text-gray-800">
                  Технические детали
                </summary>
                <pre className="mt-2 whitespace-pre-wrap text-xs bg-gray-100 p-2 rounded">
                  {this.state.error && this.state.error.toString()}
                  {this.state.errorInfo?.componentStack}
                </pre>
              </details>
            </CardContent>
          </Card>
        </div>
      );
    }

    // key={remountKey} заставляет React перемонтировать поддерево при тихом
    // восстановлении, приводя DOM в согласованное состояние.
    return (
      <React.Fragment key={this.state.remountKey}>
        {this.props.children}
      </React.Fragment>
    );
  }
}

export default ErrorBoundary;