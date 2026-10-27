package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"export-service/internal/config"
	"export-service/internal/events"
	"export-service/internal/service"
)

// Handler инкапсулирует обработчики входящих HTTP-запросов сервиса экспорта
type Handler struct {
	svc    *service.PriceListService
	queue  *events.ExportQueue
	config *config.Config
}

// NewHandler создает экземпляр HTTP-обработчика
func NewHandler(svc *service.PriceListService, queue *events.ExportQueue, cfg *config.Config) *Handler {
	return &Handler{
		svc:    svc,
		queue:  queue,
		config: cfg,
	}
}

// ExportXMLPriceList инициирует сборку XML прайс-листа и возвращает метаданные файла
func (h *Handler) ExportXMLPriceList(c *gin.Context) {
	meta, rebuilt, err := h.svc.Build(c.Request.Context())
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
		"file_url":    service.PriceListURL,
		"parts_count": meta.PartsCount,
		"rebuilt":     rebuilt,
	})
}

// SendPriceListToDrom ставит асинхронную задачу на экспорт прайс-листа на API Drom.ru
func (h *Handler) SendPriceListToDrom(c *gin.Context) {
	if h.queue == nil {
		xmlData, count, err := h.svc.Generate(c.Request.Context())
		if err != nil {
			logrus.WithError(err).Error("Ошибка подготовки XML для Drom")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось подготовить данные для отправки"})
			return
		}

		if err := service.SendToDromAPI(c.Request.Context(), h.config, xmlData); err != nil {
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

// GetExportStatus возвращает статус последней или текущей задачи экспорта
func (h *Handler) GetExportStatus(c *gin.Context) {
	if h.queue == nil {
		c.JSON(http.StatusOK, gin.H{"status": "idle"})
		return
	}
	c.JSON(http.StatusOK, h.queue.GetStatus(c.Request.Context()))
}

// HealthCheck отвечает на liveness/readiness проверки
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "export-service"})
}

// ServePriceList отдает с диска готовый файл прайс-листа
func (h *Handler) ServePriceList(c *gin.Context) {
	path := h.svc.Path()

	fi, err := os.Stat(path)
	if err != nil || fi.Size() == 0 {
		// Файла еще нет — пытаемся собрать
		if _, _, buildErr := h.svc.Build(c.Request.Context()); buildErr != nil {
			c.Header("Retry-After", "120")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Прайс-лист формируется, повторите попытку через 2 минуты"})
			return
		}
		fi, err = os.Stat(path)
	}

	if err != nil || fi.Size() < 1024 {
		c.Header("Retry-After", "120")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Прайс-лист еще не готов, повторите попытку через 2 минуты"})
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.File(path)
}
