package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler — HTTP-лицо сервиса. Вся работа лежит в PriceListBuilder,
// здесь только перевод результата в ответ.
type Handler struct {
	builder *PriceListBuilder
	config  *Config
}

func NewHandler(builder *PriceListBuilder, config *Config) *Handler {
	return &Handler{builder: builder, config: config}
}

// exportXMLPriceList экспортирует прайс-лист в формате XML для Drom.
// Файл пересобирается только если склад менялся с прошлой сборки.
func (h *Handler) exportXMLPriceList(c *gin.Context) {
	meta, rebuilt, err := h.builder.Build(c.Request.Context())
	if err != nil {
		logrus.WithError(err).Error("Ошибка подготовки XML прайс-листа")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось подготовить прайс-лист"})
		return
	}

	message := "Прайс-лист актуален, пересборка не потребовалась"
	if rebuilt {
		message = "Прайс-лист успешно сгенерирован"
		logrus.WithField("parts_count", meta.PartsCount).Info("Сгенерирован XML прайс-лист")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     message,
		"file_url":    priceListURL,
		"parts_count": meta.PartsCount,
		"rebuilt":     rebuilt,
	})
}

// sendPriceListToDrom отправляет прайс-лист на API Drom.ru.
// XML собирается заново: на площадку уходит текущее состояние склада,
// а не то, что лежит на диске с прошлой сборки.
func (h *Handler) sendPriceListToDrom(c *gin.Context) {
	xmlData, count, err := h.builder.Generate(c.Request.Context())
	if err != nil {
		logrus.WithError(err).Error("Ошибка подготовки XML для Drom")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось подготовить данные для отправки"})
		return
	}

	if err := sendToDromAPI(c.Request.Context(), h.config, xmlData); err != nil {
		logrus.WithError(err).Error("Ошибка отправки на Drom API")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось отправить прайс-лист на Drom"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Прайс-лист успешно отправлен на Drom",
		"parts_count": count,
	})
}

// healthCheck отвечает на проверку живости.
func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "export-service"})
}
