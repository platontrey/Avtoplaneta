package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// v1 и legacy маршруты для инвентаря
	for _, path := range []string{"/api/v1/inventory", "/api/inventory"} {
		r.GET(path, handler.GetInventoryHandler)
	}

	// v1 и legacy маршруты для конкретной детали
	for _, path := range []string{"/api/v1/parts/item/:id", "/api/v1/parts/:id", "/api/inventory/:id"} {
		r.GET(path, handler.GetPartByIDHandler)
	}

	// v1 и legacy маршруты создания деталей
	for _, path := range []string{"/api/v1/parts", "/api/addpart"} {
		r.POST(path, handler.AddPartHandler)
	}

	// v1 и legacy маршруты обновления деталей
	for _, path := range []string{"/api/v1/parts/:id", "/api/updatepart/:id"} {
		r.PUT(path, handler.UpdatePartHandler)
	}

	// v1 и legacy маршруты удаления деталей
	for _, path := range []string{"/api/v1/parts/:id", "/api/deletepart/:id"} {
		r.DELETE(path, handler.DeletePartHandler)
	}

	// v1 и legacy маршруты отметки на удаление
	for _, path := range []string{"/api/v1/parts/:id/mark-deletion", "/api/markpartfordeletion/:id"} {
		r.POST(path, handler.MarkPartForDeletionHandler)
	}

	// v1 и legacy маршруты загрузки фото
	for _, path := range []string{"/api/v1/uploadpartphoto/:id", "/api/v1/parts/:id/photos", "/api/uploadpartphoto/:id"} {
		r.POST(path, handler.UploadPartPhotoHandler)
	}

	// v1 и legacy маршруты удаления фото
	for _, path := range []string{"/api/v1/deletepartphoto/:id", "/api/v1/parts/:id/photo", "/api/deletepartphoto/:id"} {
		r.DELETE(path, handler.DeletePartPhotoHandler)
	}

	// v1 и legacy маршруты статистики
	for _, path := range []string{"/api/v1/statistics", "/api/statistics"} {
		r.GET(path, handler.GetStatisticsHandler)
	}
	for _, path := range []string{"/api/v1/statistics/earnings", "/api/statistics/update-earnings"} {
		r.POST(path, handler.UpdateEarningsHandler)
	}

	// v1 и legacy маршруты каталогов
	for _, path := range []string{"/api/v1/part-catalog", "/api/v1/catalogs/parts", "/api/part-catalog"} {
		r.GET(path, handler.GetPartCatalogHandler)
	}
	for _, path := range []string{"/api/v1/vehicle-catalog", "/api/v1/catalogs/vehicles", "/api/vehicle-catalog"} {
		r.GET(path, handler.GetVehicleCatalogHandler)
	}

	// v1 и legacy маршруты дефектных ведомостей
	for _, path := range []string{"/api/v1/defect-reports/preview", "/api/defect-reports/preview"} {
		r.POST(path, handler.PreviewDefectReportHandler)
	}
	for _, path := range []string{"/api/v1/defect-reports", "/api/defect-reports"} {
		r.POST(path, handler.CreateDefectReportHandler)
	}

	// v1 и legacy админ-маршруты
	for _, path := range []string{"/api/v1/admin/parts/bulk-delete", "/api/admin/bulk-delete-parts"} {
		r.DELETE(path, handler.BulkDeletePartsHandler)
		r.POST(path, handler.BulkDeletePartsHandler)
	}
	for _, path := range []string{"/api/v1/admin/parts/bulk-update", "/api/admin/bulk-update-parts"} {
		r.PUT(path, handler.BulkUpdatePartsHandler)
		r.POST(path, handler.BulkUpdatePartsHandler)
	}
	for _, path := range []string{"/api/v1/admin/supplier-codes", "/api/admin/supplier-codes"} {
		r.GET(path, handler.GetSupplierCodesHandler)
	}
	for _, path := range []string{"/api/v1/admin/parts/zero-quantity", "/api/admin/delete-zero-quantity-parts/:supplier_code"} {
		r.DELETE(path, handler.DeleteZeroQuantityPartsBySupplierHandler)
		r.POST(path, handler.DeleteZeroQuantityPartsBySupplierHandler)
	}

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
