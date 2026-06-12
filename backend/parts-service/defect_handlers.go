package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// CreateDefectReportHandler отправляет дефектную ведомость в Redis Streams для асинхронной обработки
func (h *Handler) CreateDefectReportHandler(c *gin.Context) {
	var defectReportData struct {
		Brand         string `json:"brand"`
		Model         string `json:"model"`
		Year          int    `json:"year"`
		VIN           string `json:"vin"`
		Mileage       int    `json:"mileage"`
		Description   string `json:"description"`
		SelectedParts []struct {
			Name               string  `json:"name"`
			Category           string  `json:"category"`
			Description        string  `json:"description"`
			Quantity           int     `json:"quantity"`
			Price              float64 `json:"price"`
			BodyBrand          string  `json:"body_brand,omitempty"`
			EngineBrand        string  `json:"engine_brand,omitempty"`
			CarReleaseDate     string  `json:"car_release_date,omitempty"`
			FrontRear          string  `json:"front_rear,omitempty"`
			LeftRight          string  `json:"left_right,omitempty"`
			TopBottom          string  `json:"top_bottom,omitempty"`
			Number             string  `json:"number,omitempty"`
			Manufacturer       string  `json:"manufacturer,omitempty"`
			ManufacturerCode   string  `json:"manufacturer_code,omitempty"`
			OEMCode            string  `json:"oem_code,omitempty"`
			Color              string  `json:"color,omitempty"`
			Condition          string  `json:"condition,omitempty"`
			SupplierCode       string  `json:"supplier_code,omitempty"`
			Defect             string  `json:"defect,omitempty"`
			Transmission       string  `json:"transmission,omitempty"`
			Drive              string  `json:"drive,omitempty"`
			WearPercentage     string  `json:"wear_percentage,omitempty"`
			Season             string  `json:"season,omitempty"`
			Diameter           string  `json:"diameter,omitempty"`
			Width              string  `json:"width,omitempty"`
			Profile            string  `json:"profile,omitempty"`
			TireQuantity       string  `json:"tire_quantity,omitempty"`
			Drilling           string  `json:"drilling,omitempty"`
			Offset             string  `json:"offset,omitempty"`
			CenterHoleDiameter string  `json:"center_hole_diameter,omitempty"`
			TireModel          string  `json:"tire_model,omitempty"`
			VIN                string  `json:"vin,omitempty"`
		} `json:"selectedParts"`
	}

	if err := c.ShouldBindJSON(&defectReportData); err != nil {
		fmt.Printf("Invalid JSON in createDefectReport: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Queueing defect report for %s %s %d with %d parts\n", defectReportData.Brand, defectReportData.Model, defectReportData.Year, len(defectReportData.SelectedParts))

	payload, err := json.Marshal(defectReportData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сериализации данных"})
		return
	}

	// Отправляем дефектную ведомость в Redis Stream
	err = redisClient.XAdd(c.Request.Context(), &redis.XAddArgs{
		Stream: "events:orders",
		Values: map[string]interface{}{
			"type": "defect_report_created",
			"data": string(payload),
		},
	}).Err()
	if err != nil {
		fmt.Printf("Failed to publish defect report event to Redis: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось отправить ведомость в очередь обработки"})
		return
	}

	// Логируем активность пользователя
	h.logUserActivity(c, "queue_defect_report", "part", fmt.Sprintf("Queued defect report with %d parts for %s %s %d", len(defectReportData.SelectedParts), defectReportData.Brand, defectReportData.Model, defectReportData.Year), nil)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Дефектная ведомость отправлена в очередь обработки",
	})
}
