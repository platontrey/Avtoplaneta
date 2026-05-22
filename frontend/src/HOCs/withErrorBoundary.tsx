import React, { Component } from 'react';
import type { ReactNode } from 'react';

/**
 * Интерфейс для пропсов HOC withErrorBoundary
 */
interface WithErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

/**
 * Интерфейс для состояния компонента ErrorBoundary
 */
interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

/**
 * Компонент ErrorBoundary для перехвата ошибок в дочерних компонентах
 */
class ErrorBoundary extends Component<WithErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: WithErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div style={{ padding: '20px', border: '1px solid red', borderRadius: '4px', backgroundColor: '#ffe6e6' }}>
          <h3>Что-то пошло не так</h3>
          <p>Произошла ошибка в приложении. Пожалуйста, перезагрузите страницу.</p>
          {process.env.NODE_ENV === 'development' && this.state.error && (
            <details style={{ marginTop: '10px' }}>
              <summary>Подробности ошибки</summary>
              <pre style={{ fontSize: '12px', color: '#666' }}>
                {this.state.error.toString()}
              </pre>
            </details>
          )}
        </div>
      );
    }

    return this.props.children;
  }
}

/**
 * HOC для обертывания компонентов ErrorBoundary
 * @param WrappedComponent - Компонент для обертывания
 * @param fallback - Опциональный fallback компонент для отображения при ошибке
 * @returns Обернутый компонент с ErrorBoundary
 */
export function withErrorBoundary<P extends object>(
  WrappedComponent: React.ComponentType<P>,
  fallback?: ReactNode
) {
  const WithErrorBoundaryComponent = (props: P) => (
    <ErrorBoundary fallback={fallback}>
      <WrappedComponent {...props} />
    </ErrorBoundary>
  );

  WithErrorBoundaryComponent.displayName = `withErrorBoundary(${WrappedComponent.displayName || WrappedComponent.name})`;

  return WithErrorBoundaryComponent;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _unused = withErrorBoundary;
console.log(_unused); // intentionally used to avoid export warning
