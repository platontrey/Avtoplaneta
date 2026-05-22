import React from 'react';
import type { ReactNode } from 'react';

/**
 * Интерфейс для пропсов HOC withLoading
 */
interface WithLoadingProps {
  children: ReactNode;
  isLoading: boolean;
  loadingComponent?: ReactNode;
}

/**
 * Компонент для отображения состояния загрузки
 */
const LoadingSpinner: React.FC = () => (
  <div style={{
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    padding: '20px',
    minHeight: '100px'
  }}>
    <div style={{
      width: '40px',
      height: '40px',
      border: '4px solid #f3f3f3',
      borderTop: '4px solid #3498db',
      borderRadius: '50%',
      animation: 'spin 1s linear infinite'
    }} />
    <style>{`
      @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
      }
    `}</style>
  </div>
);

/**
 * Компонент LoadingWrapper для условного рендеринга
 */
const LoadingWrapper: React.FC<WithLoadingProps> = ({
  children,
  isLoading,
  loadingComponent
}) => {
  if (isLoading) {
    return <>{loadingComponent || <LoadingSpinner />}</>;
  }

  return <>{children}</>;
};

/**
 * HOC для добавления состояния загрузки к компонентам
 * @param WrappedComponent - Компонент для обертывания
 * @param loadingComponent - Опциональный компонент загрузки
 * @returns Обернутый компонент с логикой загрузки
 */
export function withLoading<P extends object>(
  WrappedComponent: React.ComponentType<P>,
  loadingComponent?: ReactNode
) {
  const WithLoadingComponent = (props: P & { isLoading?: boolean }) => {
    const { isLoading = false, ...restProps } = props;

    return (
      <LoadingWrapper
        isLoading={isLoading}
        loadingComponent={loadingComponent}
      >
        <WrappedComponent {...(restProps as P)} />
      </LoadingWrapper>
    );
  };

  WithLoadingComponent.displayName = `withLoading(${WrappedComponent.displayName || WrappedComponent.name})`;

  return WithLoadingComponent;
}

/**
 * HOC для компонентов, которые сами управляют состоянием загрузки
 * @param WrappedComponent - Компонент для обертывания
 * @param loadingProp - Название пропса, который содержит состояние загрузки
 * @param loadingComponent - Опциональный компонент загрузки
 * @returns Обернутый компонент с логикой загрузки
 */
export function withLoadingFromProp<P extends object>(
  WrappedComponent: React.ComponentType<P>,
  loadingProp: keyof P,
  loadingComponent?: ReactNode
) {
  const WithLoadingComponent = (props: P) => {
    const isLoading = Boolean(props[loadingProp]);

    return (
      <LoadingWrapper
        isLoading={isLoading}
        loadingComponent={loadingComponent}
      >
        <WrappedComponent {...props} />
      </LoadingWrapper>
    );
  };

  WithLoadingComponent.displayName = `withLoadingFromProp(${WrappedComponent.displayName || WrappedComponent.name})`;

  return WithLoadingComponent;
}

export default withLoading;