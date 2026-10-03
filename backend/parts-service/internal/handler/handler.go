package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"parts-service/internal/catalog"
	"parts-service/internal/grpc"
	"parts-service/internal/service"
)

// Handler содержит все HTTP handlers для сервиса запчастей
// Отвечает только за обработку HTTP запросов и ответов, делегируя бизнес-логику сервису
type Handler struct {
	inventoryService service.InventoryService
	partCatalog      *catalog.PartCatalog
	vehicleCatalog   *catalog.VehicleCatalog
	defectReports    *service.DefectReportWorkflow
}

// NewHandler создает новый handler с dependency injection.
func NewHandler(
	inventoryService service.InventoryService,
	partCatalog *catalog.PartCatalog,
	vehicleCatalog *catalog.VehicleCatalog,
	defectReports *service.DefectReportWorkflow,
) *Handler {
	return &Handler{
		inventoryService: inventoryService,
		partCatalog:      partCatalog,
		vehicleCatalog:   vehicleCatalog,
		defectReports:    defectReports,
	}
}

// logUserActivity логирует активность пользователя через gRPC
func (h *Handler) logUserActivity(c *gin.Context, action, resourceType, details string, resourceID *int64) {
	userIDStr := c.GetHeader("X-User-ID")
	userEmail := c.GetHeader("X-User-Email")
	userName := c.GetHeader("X-User-Name")

	if userIDStr == "" {
		logrus.Warn("Cannot log activity - no user ID in headers")
		return
	}

	var rid uint32
	if resourceID != nil {
		rid = uint32(*resourceID)
	}
	grpc.LogUserActivityGRPC(c.Request.Context(), userIDStr, userName, userEmail, action, resourceType, details, rid)
}

// getUpdatedFields возвращает строку с именами обновленных полей
func getUpdatedFields(updates map[string]interface{}) string {
	var fields []string
	for key := range updates {
		fields = append(fields, key)
	}
	return strings.Join(fields, ", ")
}
