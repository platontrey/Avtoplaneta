package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CreateDefectReportHandler создает дефектную ведомость и массово добавляет выбранные запчасти
func (h *Handler) CreateDefectReportHandler(c *gin.Context) {
	var defectReportData struct {
		Brand         string `json:"brand"`
		Model         string `json:"model"`
		Year          int    `json:"year"`
		VIN           string `json:"vin"`
		Mileage       int    `json:"mileage"`
		Description   string `json:"description"`
		SelectedParts []struct {
			Name        string  `json:"name"`
			Category    string  `json:"category"`
			Description string  `json:"description"`
			Quantity    int     `json:"quantity"`
			Price       float64 `json:"price"`
			// Характеристики запчасти
			BodyBrand          string `json:"body_brand,omitempty"`
			EngineBrand        string `json:"engine_brand,omitempty"`
			CarReleaseDate     string `json:"car_release_date,omitempty"`
			FrontRear          string `json:"front_rear,omitempty"`
			LeftRight          string `json:"left_right,omitempty"`
			TopBottom          string `json:"top_bottom,omitempty"`
			Number             string `json:"number,omitempty"`
			Manufacturer       string `json:"manufacturer,omitempty"`
			ManufacturerCode   string `json:"manufacturer_code,omitempty"`
			OEMCode            string `json:"oem_code,omitempty"`
			Color              string `json:"color,omitempty"`
			Condition          string `json:"condition,omitempty"`
			SupplierCode       string `json:"supplier_code,omitempty"`
			Defect             string `json:"defect,omitempty"`
			Transmission       string `json:"transmission,omitempty"`
			Drive              string `json:"drive,omitempty"`
			WearPercentage     string `json:"wear_percentage,omitempty"`
			Season             string `json:"season,omitempty"`
			Diameter           string `json:"diameter,omitempty"`
			Width              string `json:"width,omitempty"`
			Profile            string `json:"profile,omitempty"`
			TireQuantity       string `json:"tire_quantity,omitempty"`
			Drilling           string `json:"drilling,omitempty"`
			Offset             string `json:"offset,omitempty"`
			CenterHoleDiameter string `json:"center_hole_diameter,omitempty"`
			TireModel          string `json:"tire_model,omitempty"`
			VIN                string `json:"vin,omitempty"`
		} `json:"selectedParts"`
	}

	if err := c.ShouldBindJSON(&defectReportData); err != nil {
		fmt.Printf("Invalid JSON in createDefectReport: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Creating defect report for %s %s %d with %d parts\n", defectReportData.Brand, defectReportData.Model, defectReportData.Year, len(defectReportData.SelectedParts))

	// Получить первого пользователя как продавца по умолчанию
	// Поскольку User модель из другого сервиса, используем простую структуру
	type User struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	var defaultUser User
	if err := db.Table("users").First(&defaultUser).Error; err != nil {
		fmt.Printf("Error getting default user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить пользователя"})
		return
	}

	var createdParts []Part
	var createdPartIDs []uint

	// Создать запчасти массово
	for _, selectedPart := range defectReportData.SelectedParts {
		part := Part{
			PartCore: PartCore{
				Name:        selectedPart.Name,
				Quantity:    selectedPart.Quantity,
				Description: selectedPart.Description,
				Category:    selectedPart.Category,
				Price:       selectedPart.Price,
				Salesman:    defaultUser.Name,
				Location:    "", // Будет установлено позже
				Status:      true,
				Brand:       defectReportData.Brand,
				Model:       defectReportData.Model,
				Photo:       "",
				SellerID:    defaultUser.ID,
			},
			PartSpecifications: PartSpecifications{
				BodyBrand:        selectedPart.BodyBrand,
				EngineBrand:      selectedPart.EngineBrand,
				CarReleaseDate:   selectedPart.CarReleaseDate,
				FrontRear:        selectedPart.FrontRear,
				LeftRight:        selectedPart.LeftRight,
				TopBottom:        selectedPart.TopBottom,
				Number:           selectedPart.Number,
				Manufacturer:     selectedPart.Manufacturer,
				ManufacturerCode: selectedPart.ManufacturerCode,
				OEMCode:          selectedPart.OEMCode,
				Color:            selectedPart.Color,
				Condition:        selectedPart.Condition,
				SupplierCode:     selectedPart.SupplierCode,
				Defect:           selectedPart.Defect,
				Transmission:     selectedPart.Transmission,
				Drive:            selectedPart.Drive,
				WearPercentage:   selectedPart.WearPercentage,
			},
			PartTireSpecifications: PartTireSpecifications{
				Season:             selectedPart.Season,
				Diameter:           selectedPart.Diameter,
				Width:              selectedPart.Width,
				Profile:            selectedPart.Profile,
				TireQuantity:       selectedPart.TireQuantity,
				Drilling:           selectedPart.Drilling,
				Offset:             selectedPart.Offset,
				CenterHoleDiameter: selectedPart.CenterHoleDiameter,
				TireModel:          selectedPart.TireModel,
			},
		}

		// Санитизация
		part.PartCore.Name = strings.TrimSpace(part.PartCore.Name)
		part.PartCore.Description = strings.TrimSpace(part.PartCore.Description)
		part.PartCore.Category = strings.TrimSpace(part.PartCore.Category)
		part.PartCore.Salesman = strings.TrimSpace(part.PartCore.Salesman)
		part.PartCore.Location = strings.TrimSpace(part.PartCore.Location)
		part.PartCore.Brand = strings.TrimSpace(part.PartCore.Brand)
		part.PartCore.Model = strings.TrimSpace(part.PartCore.Model)

		if err := db.Create(&part).Error; err != nil {
			fmt.Printf("Error creating part %s: %v\n", selectedPart.Name, err)
			continue // Продолжить с другими запчастями
		}

		// Получить созданную запчасть для индексации
		var createdPart Part
		if err := db.Last(&createdPart).Error; err != nil {
			fmt.Printf("Error getting created part: %v\n", err)
			continue
		}

		createdParts = append(createdParts, createdPart)
		createdPartIDs = append(createdPartIDs, createdPart.ID)

		// Индексировать запчасть в Elasticsearch
		if err := IndexPart(&createdPart); err != nil {
			fmt.Printf("Warning: Failed to index part %d in Elasticsearch: %v\n", createdPart.ID, err)
		} else {
			fmt.Printf("Successfully indexed part %d in Elasticsearch\n", createdPart.ID)
		}
	}

	fmt.Printf("Successfully created %d parts for defect report\n", len(createdParts))

	// Логируем создание дефектной ведомости
	h.logUserActivity(c, "create_defect_report", "part", fmt.Sprintf("Created defect report with %d parts for %s %s %d", len(createdParts), defectReportData.Brand, defectReportData.Model, defectReportData.Year), nil)

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Дефектная ведомость создана успешно",
		"createdParts": len(createdParts),
		"partIDs":      createdPartIDs,
	})
}
