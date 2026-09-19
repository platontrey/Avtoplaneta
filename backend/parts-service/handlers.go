package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/httpcache"
)

// Handler содержит все HTTP handlers для сервиса запчастей
// Отвечает только за обработку HTTP запросов и ответов, делегируя бизнес-логику сервису
type Handler struct {
	inventoryService InventoryService
	partCatalog      *PartCatalog
	vehicleCatalog   *VehicleCatalog
	defectReports    *DefectReportWorkflow
}

// getUpdatedFields возвращает строку с именами обновленных полей
func getUpdatedFields(updates map[string]interface{}) string {
	var fields []string
	for key := range updates {
		fields = append(fields, key)
	}
	return strings.Join(fields, ", ")
}

// NewHandler создает новый handler с dependency injection.
// Каталог и workflow передаются явно: конструктор ничего не собирает сам
// и не обращается к глобальному состоянию.
func NewHandler(inventoryService InventoryService, catalog *PartCatalog, vehicles *VehicleCatalog, defectReports *DefectReportWorkflow) *Handler {
	return &Handler{
		inventoryService: inventoryService,
		partCatalog:      catalog,
		vehicleCatalog:   vehicles,
		defectReports:    defectReports,
	}
}

// logUserActivity логирует активность пользователя через gRPC (с HTTP fallback)
func (h *Handler) logUserActivity(c *gin.Context, action, resourceType, details string, resourceID *int64) {
	userIDStr := c.GetHeader("X-User-ID")
	userEmail := c.GetHeader("X-User-Email")
	userName := c.GetHeader("X-User-Name")

	if userIDStr == "" {
		logrus.Warn("Cannot log activity - no user ID in headers")
		return
	}

	// Пробуем gRPC
	var rid uint32
	if resourceID != nil {
		rid = uint32(*resourceID)
	}
	logUserActivityGRPC(c.Request.Context(), userIDStr, userName, userEmail, action, resourceType, details, rid)
}

// GetInventoryHandler получает список запчастей с фильтрами
// @Summary Получить инвентарь запчастей
// @Description Возвращает список запчастей с возможностью фильтрации по различным параметрам
// @Tags inventory
// @Accept json
// @Produce json
// @Param search query string false "Поиск по названию или описанию"
// @Param category query string false "Фильтр по категории"
// @Param brand query string false "Фильтр по бренду"
// @Param model query string false "Фильтр по модели"
// @Param location query string false "Фильтр по расположению"
// @Param address query string false "Фильтр по адресу склада"
// @Param salesman query string false "Фильтр по продавцу"
// @Param status query string false "Фильтр по статусу"
// @Param hasPhoto query string false "Фильтр по наличию фото (with/without/all)"
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(50)
// @Success 200 {array} Part
// @Failure 500 {object} map[string]string
// @Router /api/inventory [get]
func (h *Handler) GetInventoryHandler(c *gin.Context) {
	// Парсим параметры запроса
	queryParams := InventoryQueryParams{
		Search:             c.Query("search"),
		Category:           c.Query("category"),
		Brand:              c.Query("brand"),
		Model:              c.Query("model"),
		Location:           c.Query("location"),
		Address:            c.Query("address"),
		Salesman:           c.Query("salesman"),
		Status:             c.Query("status"),
		HasPhoto:           c.Query("hasPhoto"),
		Number:             c.Query("number"),
		OEMCode:            c.Query("oem_code"),
		VIN:                c.Query("vin"),
		BodyBrand:          c.Query("body_brand"),
		EngineBrand:        c.Query("engine_brand"),
		CarReleaseDate:     c.Query("car_release_date"),
		CarReleasePeriod:   c.Query("car_release_period"),
		Transmission:       c.Query("transmission"),
		Drive:              c.Query("drive"),
		Condition:          c.Query("condition"),
		Manufacturer:       c.Query("manufacturer"),
		Defect:             c.Query("defect"),
		Color:              c.Query("color"),
		MinPrice:           c.Query("min_price"),
		MaxPrice:           c.Query("max_price"),
		MinQuantity:        c.Query("min_quantity"),
		MaxQuantity:        c.Query("max_quantity"),
		FrontRear:          c.Query("front_rear"),
		LeftRight:          c.Query("left_right"),
		TopBottom:          c.Query("top_bottom"),
		ManufacturerCode:   c.Query("manufacturer_code"),
		SupplierCode:       c.Query("supplier_code"),
		TransmissionModel:  c.Query("transmission_model"),
		WearPercentage:     c.Query("wear_percentage"),
		Season:             c.Query("season"),
		Diameter:           c.Query("diameter"),
		Width:              c.Query("width"),
		Profile:            c.Query("profile"),
		TireQuantity:       c.Query("tire_quantity"),
		Drilling:           c.Query("drilling"),
		Offset:             c.Query("offset"),
		CenterHoleDiameter: c.Query("center_hole_diameter"),
		TireModel:          c.Query("tire_model"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	queryParams.Page = page
	queryParams.Limit = limit

	ctx := c.Request.Context()
	parts, err := h.inventoryService.GetInventory(ctx, queryParams)
	if err != nil {
		fmt.Printf("GetInventory FAILED with error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить инвентарь: " + err.Error()})
		return
	}

	if strings.HasPrefix(c.Request.URL.Path, "/api/v1") {
		c.JSON(http.StatusOK, gin.H{
			"parts": parts,
			"total": len(parts),
			"page":  page,
			"limit": limit,
		})
		return
	}

	c.JSON(http.StatusOK, parts)
}

// AddPartHandler добавляет новую запчасть
func (h *Handler) AddPartHandler(c *gin.Context) {
	var part Part
	if err := c.ShouldBindJSON(&part); err != nil {
		fmt.Printf("Invalid JSON in AddPart: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Received part data: %+v\n", part)

	ctx := c.Request.Context()
	createdPart, err := h.inventoryService.AddPart(ctx, &part)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add part"})
		return
	}

	// Логируем создание запчасти
	partID := createdPart.ID
	h.logUserActivity(c, "create_part", "part", fmt.Sprintf("Created part: %s (ID: %d, Category: %s, Quantity: %d, Price: %.2f)", part.Name, partID, part.Category, part.Quantity, part.Price), &partID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Часть добавлена успешно",
		"id":      createdPart.ID,
	})
}

// UpdatePartHandler обновляет запчасть по ID
func (h *Handler) UpdatePartHandler(c *gin.Context) {
	partIDStr := c.Param("id")
	partID, err := strconv.ParseUint(partIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		fmt.Printf("Invalid JSON in UpdatePart: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Received updates for part %s: %+v\n", partIDStr, updateData)

	ctx := c.Request.Context()
	if err := h.inventoryService.UpdatePart(ctx, int64(partID), updateData); err != nil {
		fmt.Printf("UpdatePart FAILED with error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось обновить часть: %v", err)})
		return
	}

	// Логируем обновление запчасти
	partIDUint := int64(partID)
	h.logUserActivity(c, "update_part", "part", fmt.Sprintf("Updated part ID: %d with fields: %v", partID, getUpdatedFields(updateData)), &partIDUint)

	c.JSON(http.StatusOK, gin.H{"message": "Часть обновлена успешно"})
}

// DeletePartHandler удаляет запчасть по ID
func (h *Handler) DeletePartHandler(c *gin.Context) {
	idStr := c.Param("id")
	partID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	ctx := c.Request.Context()
	if err := h.inventoryService.DeletePart(ctx, int64(partID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить часть"})
		return
	}

	// Логируем удаление запчасти
	partIDUint := int64(partID)
	part, _ := h.inventoryService.GetPartByID(ctx, partIDUint) // Получить данные запчасти перед удалением
	partName := "Unknown"
	if part != nil {
		partName = part.Name
	}
	h.logUserActivity(c, "delete_part", "part", fmt.Sprintf("Deleted part: %s (ID: %d)", partName, partID), &partIDUint)

	c.JSON(http.StatusOK, gin.H{"message": "Запчасть удалена успешно"})
}

// MarkPartForDeletionHandler отмечает запчасть для удаления
func (h *Handler) MarkPartForDeletionHandler(c *gin.Context) {
	idStr := c.Param("id")
	partID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	ctx := c.Request.Context()
	if err := h.inventoryService.MarkPartForDeletion(ctx, int64(partID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось отметить часть для удаления"})
		return
	}

	// Логируем отметку для удаления
	partIDUint := int64(partID)
	h.logUserActivity(c, "mark_part_for_deletion", "part", fmt.Sprintf("Marked part ID: %d for deletion", partID), &partIDUint)

	c.JSON(http.StatusOK, gin.H{
		"message": "Часть отмечена для удаления через 14 дней",
	})
}

// GetStatisticsHandler возвращает статистику по запчастям
func (h *Handler) GetStatisticsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Проверка версии стоит один индексный запрос, а сбор статистики — агрегаты
	// по всему складу и поход в orders-service.
	if version, err := h.inventoryService.InventoryVersion(ctx); err == nil {
		if httpcache.ServeVersioned(c.Writer, c.Request, version, time.Minute) {
			return
		}
	}

	stats, err := h.inventoryService.GetStatistics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить статистику"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UploadPartPhotoHandler загружает фото для запчасти
func (h *Handler) UploadPartPhotoHandler(c *gin.Context) {
	idStr := c.Param("id")
	partID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	ctx := c.Request.Context()
	photoPath, err := h.inventoryService.UploadPartPhoto(ctx, int64(partID), c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем загрузку фото
	partIDUint := int64(partID)
	h.logUserActivity(c, "upload_photo", "part", fmt.Sprintf("Uploaded photo for part ID: %d", partID), &partIDUint)

	// Получить файл для размера и типа
	file, _ := c.FormFile("photo")

	// Получить обновленную часть для возврата всех фото
	updatedPart, err := h.inventoryService.GetPartByID(ctx, int64(partID))
	if err != nil {
		fmt.Printf("DEBUG UploadPartPhotoHandler: Failed to fetch updated part: %v\n", err)
		updatedPart = &Part{} // Используем пустую структуру, если не удалось получить
	} else {
		fmt.Printf("DEBUG UploadPartPhotoHandler: Updated part photos: %v\n", updatedPart.Photos)
	}

	response := gin.H{
		"message": "Фото загружено успешно",
		"photo":   photoPath,
		"photos":  updatedPart.Photos,
	}

	// Добавить информацию о файле, если он существует
	if file != nil {
		response["filename"] = file.Filename
		response["size"] = file.Size
		response["type"] = file.Header.Get("Content-Type")
	}

	c.JSON(http.StatusOK, response)
}

// DeletePartPhotoHandler удаляет фото запчасти
func (h *Handler) DeletePartPhotoHandler(c *gin.Context) {
	idStr := c.Param("id")
	partID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	// Получить путь к конкретному фото из query параметров (опционально)
	photoPath := c.Query("photo")
	if photoPath == "" {
		photoPath = c.Query("photo_url")
	}
	fmt.Printf("DeletePartPhotoHandler: partID=%d, photoPath='%s'\n", partID, photoPath)

	ctx := c.Request.Context()
	if err := h.inventoryService.DeletePartPhoto(ctx, int64(partID), photoPath); err != nil {
		fmt.Printf("DeletePartPhotoHandler: Error deleting photo: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем удаление фото
	partIDUint := int64(partID)
	photoInfo := "all photos"
	if photoPath != "" {
		photoInfo = photoPath
	}
	fmt.Printf("DeletePartPhotoHandler: Successfully deleted %s for part ID %d\n", photoInfo, partID)
	h.logUserActivity(c, "delete_photo", "part", fmt.Sprintf("Deleted photo %s for part ID: %d", photoInfo, partID), &partIDUint)

	c.JSON(http.StatusOK, gin.H{"message": "Фото удалено успешно"})
}

// BulkDeletePartsHandler удаляет несколько запчастей
func (h *Handler) BulkDeletePartsHandler(c *gin.Context) {
	var requestData struct {
		IDs []int64 `json:"ids"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		logrus.WithError(err).Error("BulkDeletePartsHandler: Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"ids":   requestData.IDs,
		"count": len(requestData.IDs),
	}).Info("BulkDeletePartsHandler: Received request")

	ctx := c.Request.Context()
	if err := h.inventoryService.BulkDeleteParts(ctx, requestData.IDs); err != nil {
		logrus.WithError(err).Error("BulkDeletePartsHandler: Failed to bulk delete parts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить запчасти"})
		return
	}

	// Логируем массовое удаление
	h.logUserActivity(c, "bulk_delete_parts", "part", fmt.Sprintf("Bulk deleted %d parts", len(requestData.IDs)), nil)

	logrus.Info("BulkDeletePartsHandler: Successfully deleted parts")
	c.JSON(http.StatusOK, gin.H{"message": "Запчасти удалены успешно"})
}

// BulkUpdatePartsHandler обновляет несколько запчастей
func (h *Handler) BulkUpdatePartsHandler(c *gin.Context) {
	var updates []map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		logrus.WithError(err).Error("BulkUpdatePartsHandler: Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"updates": updates,
		"count":   len(updates),
	}).Info("BulkUpdatePartsHandler: Received request")

	ctx := c.Request.Context()
	updatedCount, err := h.inventoryService.BulkUpdateParts(ctx, updates)
	if err != nil {
		logrus.WithError(err).Error("BulkUpdatePartsHandler: Failed to bulk update parts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить запчасти"})
		return
	}

	// Логируем массовое обновление
	h.logUserActivity(c, "bulk_update_parts", "part", fmt.Sprintf("Bulk updated %d parts", updatedCount), nil)

	logrus.Info("BulkUpdatePartsHandler: Successfully updated parts")
	c.JSON(http.StatusOK, gin.H{
		"message":       "Запчасти обновлены успешно",
		"updated_count": updatedCount,
	})
}

// DeleteZeroQuantityPartsBySupplierHandler удаляет запчасти с нулевым количеством по поставщику
func (h *Handler) DeleteZeroQuantityPartsBySupplierHandler(c *gin.Context) {
	supplierCode := c.Param("supplier_code")

	fmt.Printf("DeleteZeroQuantityPartsBySupplierHandler: supplier_code='%s'\n", supplierCode)

	ctx := c.Request.Context()
	deletedCount, err := h.inventoryService.DeleteZeroQuantityPartsBySupplier(ctx, supplierCode)
	if err != nil {
		fmt.Printf("DeleteZeroQuantityPartsBySupplierHandler: error deleting parts for supplier_code='%s': %v\n", supplierCode, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить запчасти"})
		return
	}

	fmt.Printf("DeleteZeroQuantityPartsBySupplierHandler: successfully deleted %d parts for supplier_code='%s'\n", deletedCount, supplierCode)

	// Логируем массовое удаление запчастей с quantity=0
	h.logUserActivity(c, "delete_zero_quantity_parts", "part", fmt.Sprintf("Deleted %d zero-quantity parts for supplier code: %s", deletedCount, supplierCode), nil)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Запчасти с нулевым количеством удалены",
		"deleted_count": deletedCount,
	})
}

// GetSupplierCodesHandler получает коды поставщиков
func (h *Handler) GetSupplierCodesHandler(c *gin.Context) {
	ctx := c.Request.Context()

	if version, err := h.inventoryService.InventoryVersion(ctx); err == nil {
		if httpcache.ServeVersioned(c.Writer, c.Request, version, 5*time.Minute) {
			return
		}
	}

	codes, err := h.inventoryService.GetSupplierCodes(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить коды поставщиков"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"supplier_codes": codes})
}

// UpdateEarningsHandler обновляет общий заработок
func (h *Handler) UpdateEarningsHandler(c *gin.Context) {
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	ctx := c.Request.Context()
	if err := h.inventoryService.UpdateEarnings(ctx, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить заработок"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заработок обновлен"})
}

// GetPartByIDHandler получает запчасть по ID
func (h *Handler) GetPartByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID запчасти"})
		return
	}

	ctx := c.Request.Context()
	part, err := h.inventoryService.GetPartByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Запчасть не найдена"})
		return
	}

	c.JSON(http.StatusOK, part)
}
