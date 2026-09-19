package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	partsv1 "avtoplaneta/gen/parts/v1"
)

// SetupRoutes настраивает маршруты для приложения с dependency injection
func SetupRoutes(r *gin.Engine, handler *Handler) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Add metrics middleware
	r.Use(metricsMiddleware())

	// Add /metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Монтируем grpc-gateway для /api/v1/* (совместимость с фронтендом)
	setupGRPCGatewayRoutes(r)

	// Основные маршруты сервиса запчастей
	r.GET("/api/inventory", handler.GetInventoryHandler)
	r.GET("/api/inventory/:id", handler.GetPartByIDHandler)
	r.POST("/api/addpart", handler.AddPartHandler)
	r.DELETE("/api/deletepart/:id", handler.DeletePartHandler)
	r.PUT("/api/updatepart/:id", handler.UpdatePartHandler)
	r.POST("/api/uploadpartphoto/:id", handler.UploadPartPhotoHandler)
	r.DELETE("/api/deletepartphoto/:id", handler.DeletePartPhotoHandler)
	r.POST("/api/markpartfordeletion/:id", handler.MarkPartForDeletionHandler)
	r.GET("/api/statistics", handler.GetStatisticsHandler)
	r.GET("/api/part-catalog", handler.GetPartCatalogHandler)
	r.POST("/api/statistics/update-earnings", handler.UpdateEarningsHandler)

	// v1 маршруты и алиасы
	r.POST("/api/v1/uploadpartphoto/:id", handler.UploadPartPhotoHandler)
	r.POST("/api/v1/parts/:id/photos", handler.UploadPartPhotoHandler)
	r.DELETE("/api/v1/deletepartphoto/:id", handler.DeletePartPhotoHandler)
	r.GET("/api/v1/part-catalog", handler.GetPartCatalogHandler)
	r.GET("/api/v1/catalogs/parts", handler.GetPartCatalogHandler)
	r.GET("/api/v1/vehicle-catalog", handler.GetVehicleCatalogHandler)
	r.GET("/api/v1/catalogs/vehicles", handler.GetVehicleCatalogHandler)

	// Маршруты для характеристик запчастей больше не нужны - характеристики хранятся в основной таблице Part

	// Маршруты для дефектных ведомостей
	r.GET("/api/vehicle-catalog", handler.GetVehicleCatalogHandler)
	r.POST("/api/defect-reports/preview", handler.PreviewDefectReportHandler)
	r.POST("/api/defect-reports", handler.CreateDefectReportHandler)

	// Админ маршруты
	r.DELETE("/api/admin/delete-zero-quantity-parts/:supplier_code", handler.DeleteZeroQuantityPartsBySupplierHandler)
	r.GET("/api/admin/supplier-codes", handler.GetSupplierCodesHandler)
	r.DELETE("/api/admin/bulk-delete-parts", handler.BulkDeletePartsHandler)
	r.PUT("/api/admin/bulk-update-parts", handler.BulkUpdatePartsHandler)

	// Статическое обслуживание файлов для загрузок
	uploads := r.Group("/uploads", uploadsCacheControl())
	uploads.Static("", "./uploads")
}

// uploadsCacheControl проставляет политику кэширования для файлов из /uploads.
//
// Здесь лежат только фотографии запчастей. Они сохраняются под уникальным
// именем вида <id>_<unix>.jpg и никогда не перезаписываются: при замене
// фотографии меняется сам адрес. Значит, старый адрес всегда указывает на одно
// и то же содержимое, и его можно кэшировать навсегда — это самый большой кусок
// трафика в приложении.
//
// Прайс-лист раньше лежал здесь же и требовал исключения, потому что живёт по
// постоянному адресу. Теперь его собирает и отдаёт export-service.
func uploadsCacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Next()
	}
}

func setupGRPCGatewayRoutes(r *gin.Engine) {
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9081"
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: false,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	err := partsv1.RegisterPartsServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		"localhost:"+grpcPort,
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to register parts service handler in grpc-gateway")
		return
	}

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/") {
			gwmux.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
}
