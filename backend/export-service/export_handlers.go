package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler — HTTP-лицо сервиса. Вся работа лежит в PriceListBuilder и ExportQueue,
// здесь только перевод результата в ответ.
type Handler struct {
	builder *PriceListBuilder
	queue   *ExportQueue
	config  *Config
}

func NewHandler(builder *PriceListBuilder, queue *ExportQueue, config *Config) *Handler {
	return &Handler{builder: builder, queue: queue, config: config}
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

// sendPriceListToDrom ставит задачу на экспорт прайс-листа на API Drom.ru в очередь.
// Отвечает 202 Accepted сразу, предотвращая долгие HTTP-таймауты и пиковые всплески RAM.
func (h *Handler) sendPriceListToDrom(c *gin.Context) {
	if h.queue == nil {
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
		return
	}

	task, err := h.queue.Enqueue(c.Request.Context(), "drom", c.ClientIP())
	if err != nil {
		logrus.WithError(err).Warn("Попытка запуска экспорта во время активной задачи")
		c.JSON(http.StatusConflict, gin.H{
			"error":  err.Error(),
			"status": h.queue.GetStatus(c.Request.Context()),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Задача экспорта на Drom успешно поставлена в очередь",
		"task_id":    task.TaskID,
		"status":     "queued",
		"status_url": "/api/export/status",
	})
}

// getExportStatus возвращает статус последней или текущей задачи экспорта
func (h *Handler) getExportStatus(c *gin.Context) {
	if h.queue == nil {
		c.JSON(http.StatusOK, gin.H{"status": "idle"})
		return
	}
	c.JSON(http.StatusOK, h.queue.GetStatus(c.Request.Context()))
}

// healthCheck отвечает на проверку живости.
func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "export-service"})
}
