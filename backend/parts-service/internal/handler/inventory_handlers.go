package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"parts-service/internal/domain"
)

// GetInventoryHandler получает список запчастей с фильтрами
func (h *Handler) GetInventoryHandler(c *gin.Context) {
	queryParams := domain.InventoryQueryParams{
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
	var part domain.Part
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
	part, _ := h.inventoryService.GetPartByID(ctx, partIDUint)
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
