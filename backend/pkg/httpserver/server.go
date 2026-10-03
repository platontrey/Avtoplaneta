package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultShutdownTimeout   = 30 * time.Second
	defaultReadHeaderTimeout = 10 * time.Second
	defaultIdleTimeout       = 120 * time.Second
)

// Server оборачивает стандартный http.Server с поддержкой graceful shutdown и TLS
type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
	certFile        string
	keyFile         string
}

// Option конфигурационный параметр сервера
type Option func(*Server)

// WithShutdownTimeout задает таймаут для graceful shutdown
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}

// WithTLS явным образом задает пути к файлам сертификата и ключа
func WithTLS(certFile, keyFile string) Option {
	return func(s *Server) {
		s.certFile = certFile
		s.keyFile = keyFile
	}
}

// New создает новый экземпляр Server
func New(handler http.Handler, port string, opts ...Option) *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr:              ":" + port,
			Handler:           handler,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			IdleTimeout:       defaultIdleTimeout,
		},
		shutdownTimeout: defaultShutdownTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Run запускает HTTP сервер и блокирует горутину до отмены ctx или возникновения фатальной ошибки
func (s *Server) Run(ctx context.Context) error {
	certFile, keyFile, useTLS := s.resolveTLS()

	errChan := make(chan error, 1)
	go func() {
		if useTLS {
			logrus.WithFields(logrus.Fields{
				"addr": s.httpServer.Addr,
				"cert": certFile,
			}).Info("Starting HTTP server with TLS")
			if err := s.httpServer.ListenAndServeTLS(certFile, keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- fmt.Errorf("HTTP server (TLS) error: %w", err)
			}
		} else {
			logrus.WithField("addr", s.httpServer.Addr).Info("Starting HTTP server (without TLS)")
			if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- fmt.Errorf("HTTP server error: %w", err)
			}
		}
	}()

	select {
	case <-ctx.Done():
		logrus.Info("Gracefully shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server forced shutdown: %w", err)
		}
		logrus.Info("HTTP server stopped gracefully")
		return nil

	case err := <-errChan:
		return err
	}
}

// Run удобный хелпер запуска одной строкой
func Run(ctx context.Context, port string, handler http.Handler, opts ...Option) error {
	return New(handler, port, opts...).Run(ctx)
}

func (s *Server) resolveTLS() (certFile, keyFile string, ok bool) {
	if s.certFile != "" && s.keyFile != "" {
		if fileExists(s.certFile) && fileExists(s.keyFile) {
			return s.certFile, s.keyFile, true
		}
	}

	candidates := [][2]string{
		{"../../cert.pem", "../../key.pem"},
		{"cert.pem", "key.pem"},
	}

	if os.Getenv("NODE_ENV") == "production" {
		for _, c := range candidates {
			if fileExists(c[0]) && fileExists(c[1]) {
				return c[0], c[1], true
			}
		}
	}

	return "", "", false
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
