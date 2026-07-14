package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"avtoplaneta/pkg/tracing"
	"google.golang.org/grpc"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Initialize OpenTelemetry Tracer
	tp, err := tracing.InitTracer("auth-service")
	if err != nil {
		logrus.WithError(err).Fatal("failed to initialize tracer")
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logrus.WithError(err).Error("failed to shutdown tracer")
		}
	}()

	gin.SetMode(gin.ReleaseMode)

	config := LoadConfig()
	InitDB(config)

	userRepo := NewUserRepository(dbPool)
	SetUserRepo(userRepo)

	CreateDefaultUser()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Получен сигнал завершения, начинаем graceful shutdown...")
		cancel()
	}()

	InitAuth(ctx, config)

	activityRepo := NewActivityLogRepository(dbPool)
	authService := NewAuthService(userRepo, activityRepo, store, nil)
	handler := NewHandler(authService, config)

	log.Println("Сервис аутентификации готов к работе с пользователями.")

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9083"
	}
	go func() {
		if err := StartGRPCServer(authService, config, grpcPort); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	r := gin.Default()
	r.Use(otelgin.Middleware("auth-service"))

	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Printf("Ошибка установки доверенных прокси: %v", err)
	}

	r.Use(CORSMiddleware(config))
	r.Use(csrfMiddleware)

	SetupRoutes(r, handler)

	srv := &http.Server{
		Addr:    ":" + config.Port,
		Handler: r,
	}

	errChan := make(chan error, 1)

	go func() {
		log.Printf("Сервис аутентификации запускается на порту %s", config.Port)

		certFile := "../../cert.pem"
		keyFile := "../../key.pem"
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			certFile = "cert.pem"
			keyFile = "key.pem"
		}

		if os.Getenv("NODE_ENV") == "production" {
			if _, err := os.Stat(certFile); err == nil {
				log.Println("Запуск с TLS...")
				if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && !errors.Is(http.ErrServerClosed, err) {
					errChan <- err
				}
			} else {
				log.Println("Сертификаты не найдены, запуск без TLS...")
				if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
					errChan <- err
				}
			}
		} else {
			log.Println("Режим разработки: запуск без TLS...")
			if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
				errChan <- err
			}
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Завершение работы сервиса аутентификации...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Сервер принудительно остановлен: %v", err)
		}
		log.Println("Сервис аутентификации остановлен")
	case err := <-errChan:
		log.Fatal("Ошибка сервера:", err)
	}
}
