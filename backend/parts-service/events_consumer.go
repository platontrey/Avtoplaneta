package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// EventConsumer определяет интерфейс для обработки событий
type EventConsumer interface {
	Start(ctx context.Context) error
}

// RedisEventConsumer реализует EventConsumer с использованием Redis Streams
type RedisEventConsumer struct {
	client   *redis.Client
	service  InventoryService
	group    string
	consumer string
	stream   string
}

// NewEventConsumer создает новый consumer событий
func NewEventConsumer(client *redis.Client, service InventoryService) EventConsumer {
	return &RedisEventConsumer{
		client:   client,
		service:  service,
		group:    "parts-service-group",
		consumer: "worker-1",
		stream:   "events:orders",
	}
}

// Start запускает обработку событий
func (c *RedisEventConsumer) Start(ctx context.Context) error {
	// Создаем consumer group, если не существует
	err := c.client.XGroupCreateMkStream(ctx, c.stream, c.group, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		logrus.WithError(err).Error("Failed to create consumer group")
		return err
	}

	logrus.Info("Starting Redis Streams consumer for order events")

	// Запускаем фоновый цикл восстановления зависших сообщений каждые 30 секунд
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.recoverPendingMessages(ctx)
			}
		}
	}()

	// Выполняем первоначальное восстановление при старте
	c.recoverPendingMessages(ctx)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping Redis Streams consumer")
			return nil
		default:
			// Читаем сообщения из потока (новые сообщения с ">")
			streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    c.group,
				Consumer: c.consumer,
				Streams:  []string{c.stream, ">"},
				Count:    10,
				Block:    time.Second * 5, // Блокируем на 5 секунд
			}).Result()

			if err != nil && !errors.Is(err, redis.Nil) {
				logrus.WithError(err).Error("Failed to read from Redis Streams")
				time.Sleep(time.Second) // Защита от частых ошибок при падении Redis
				continue
			}

			if len(streams) == 0 {
				continue
			}

			// Обрабатываем сообщения
			for _, stream := range streams {
				for _, msg := range stream.Messages {
					if err := c.processMessage(ctx, msg); err != nil {
						logrus.WithError(err).WithField("message_id", msg.ID).Error("Failed to process message, letting it stay in PEL for retry")
						continue // НЕ подтверждаем (no XAck), оно останется в PEL и будет повторно обработано фоновым циклом
					}

					// Подтверждаем обработку
					if err := c.client.XAck(ctx, c.stream, c.group, msg.ID).Err(); err != nil {
						logrus.WithError(err).WithField("message_id", msg.ID).Error("Failed to acknowledge message")
					}
				}
			}
		}
	}
}

// recoverPendingMessages обрабатывает зависшие сообщения (PEL), закрепленные за текущим воркером
func (c *RedisEventConsumer) recoverPendingMessages(ctx context.Context) {
	streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    c.group,
		Consumer: c.consumer,
		Streams:  []string{c.stream, "0"}, // 0 читает зависшие (pending) сообщения текущего воркера
		Count:    10,
		Block:    -1, // не блокируем
	}).Result()

	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logrus.WithError(err).Error("Failed to read pending messages from Redis Streams")
		}
		return
	}

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			logrus.WithField("message_id", msg.ID).Info("Recovering pending message from PEL")

			// Проверяем количество попыток доставки сообщения
			timesDelivered := c.getDeliveryCount(ctx, msg.ID)
			if timesDelivered > 5 {
				logrus.WithFields(logrus.Fields{
					"message_id":      msg.ID,
					"times_delivered": timesDelivered,
				}).Warn("Message exceeded max delivery attempts (5), sending to DLQ")

				c.sendToDLQ(ctx, msg, fmt.Errorf("exceeded max delivery attempts (%d)", timesDelivered))

				// Подтверждаем в основном стриме, чтобы удалить его из очереди
				c.client.XAck(ctx, c.stream, c.group, msg.ID)
				continue
			}

			if err := c.processMessage(ctx, msg); err != nil {
				logrus.WithError(err).WithField("message_id", msg.ID).Error("Failed to process recovered pending message")
				continue
			}

			// Успешно обработано — подтверждаем
			if err := c.client.XAck(ctx, c.stream, c.group, msg.ID).Err(); err != nil {
				logrus.WithError(err).WithField("message_id", msg.ID).Error("Failed to acknowledge recovered message")
			}
		}
	}
}

// getDeliveryCount возвращает количество доставок сообщения из XPENDING
func (c *RedisEventConsumer) getDeliveryCount(ctx context.Context, messageID string) int64 {
	pending, err := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: c.stream,
		Group:  c.group,
		Start:  messageID,
		End:    messageID,
		Count:  1,
	}).Result()

	if err != nil || len(pending) == 0 {
		return 1 // Если не удалось получить, считаем как 1 попытку
	}

	return pending[0].RetryCount
}

// sendToDLQ отправляет сообщение в Dead Letter Queue (стрим с суффиксом :dead)
func (c *RedisEventConsumer) sendToDLQ(ctx context.Context, msg redis.XMessage, err error) {
	dlqValues := make(map[string]interface{})
	for k, v := range msg.Values {
		dlqValues[k] = v
	}
	dlqValues["dlq_error"] = err.Error()
	dlqValues["dlq_time"] = time.Now().Format(time.RFC3339)
	dlqValues["original_id"] = msg.ID

	dlqStream := c.stream + ":dead"
	errAdd := c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: dlqStream,
		Values: dlqValues,
	}).Err()

	if errAdd != nil {
		logrus.WithError(errAdd).WithFields(logrus.Fields{
			"original_id": msg.ID,
			"dlq_stream":  dlqStream,
		}).Error("Failed to publish message to DLQ")
	} else {
		logrus.WithFields(logrus.Fields{
			"original_id": msg.ID,
			"dlq_stream":  dlqStream,
		}).Info("Message successfully moved to DLQ")
	}
}

// processMessage обрабатывает отдельное сообщение
func (c *RedisEventConsumer) processMessage(ctx context.Context, msg redis.XMessage) error {
	eventType, ok := msg.Values["type"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid message format: missing type")
		return nil
	}

	switch eventType {
	case "order_completed":
		return c.handleOrderCompleted(ctx, msg)
	case "defect_report_created":
		return c.handleDefectReportCreated(ctx, msg)
	case "part_index_requested":
		return c.handlePartIndexRequested(ctx, msg)
	case "part_delete_requested":
		return c.handlePartDeleteRequested(ctx, msg)
	case "user_action":
		return c.handleUserAction(msg)
	default:
		logrus.WithFields(logrus.Fields{
			"message_id": msg.ID,
			"type":       eventType,
		}).Warn("Unknown event type")
		return nil
	}
}

// handleOrderCompleted обрабатывает событие завершения заказа
func (c *RedisEventConsumer) handleOrderCompleted(ctx context.Context, msg redis.XMessage) error {
	orderIDStr, ok := msg.Values["order_id"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid order_completed event: missing order_id")
		return nil
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderIDStr).Error("Invalid order_id format")
		return err
	}

	amountStr, ok := msg.Values["amount"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid order_completed event: missing amount")
		return nil
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		logrus.WithError(err).WithField("amount", amountStr).Error("Invalid amount format")
		return err
	}

	// Обновляем статистику доходов
	if err := c.service.UpdateEarnings(ctx, amount); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to update earnings")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"amount":   amount,
	}).Info("Successfully processed order completed event")

	return nil
}

// handleUserAction обрабатывает событие действия пользователя
func (c *RedisEventConsumer) handleUserAction(msg redis.XMessage) error {
	userID, ok := msg.Values["user_id"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid user_action event: missing user_id")
		return nil
	}

	action, ok := msg.Values["action"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid user_action event: missing action")
		return nil
	}

	// Здесь можно добавить логику для обработки действия пользователя
	// Например, отправить в messaging-service или просто залогировать
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"action":  action,
		"details": msg.Values["details"],
	}).Info("User action logged via Redis Streams")

	return nil
}

func (c *RedisEventConsumer) handleDefectReportCreated(ctx context.Context, msg redis.XMessage) error {
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid defect_report_created event: missing data")
		return nil
	}

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

	if err := json.Unmarshal([]byte(dataStr), &defectReportData); err != nil {
		logrus.WithError(err).Error("Failed to deserialize defect report data")
		return err
	}

	var defaultUserID int64
	var defaultUserName string
	err := dbPool.QueryRow(ctx, "SELECT id, name FROM users ORDER BY id LIMIT 1").Scan(&defaultUserID, &defaultUserName)
	if err != nil {
		logrus.WithError(err).Error("Error getting default user for defect report")
		return err
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Ограничиваем параллелизм до 10 горутин

	for _, sp := range defectReportData.SelectedParts {
		selectedPart := sp // Захватываем переменную для горутины
		wg.Add(1)
		
		go func() {
			defer wg.Done()
			sem <- struct{}{}        // Занимаем слот
			defer func() { <-sem }() // Освобождаем слот

			part := Part{
				PartCore: PartCore{
					Name:        selectedPart.Name,
					Quantity:    selectedPart.Quantity,
					Description: selectedPart.Description,
					Category:    selectedPart.Category,
					Price:       selectedPart.Price,
					Salesman:    defaultUserName,
					Location:    "",
					Status:      true,
					Brand:       defectReportData.Brand,
					Model:       defectReportData.Model,
					Photo:       "",
					SellerID:    defaultUserID,
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

			part.PartCore.Name = strings.TrimSpace(part.PartCore.Name)
			part.PartCore.Description = strings.TrimSpace(part.PartCore.Description)
			part.PartCore.Category = strings.TrimSpace(part.PartCore.Category)
			part.PartCore.Salesman = strings.TrimSpace(part.PartCore.Salesman)
			part.PartCore.Location = strings.TrimSpace(part.PartCore.Location)
			part.PartCore.Brand = strings.TrimSpace(part.PartCore.Brand)
			part.PartCore.Model = strings.TrimSpace(part.PartCore.Model)

			if _, err := c.service.AddPart(ctx, &part); err != nil {
				logrus.WithError(err).WithField("part_name", part.PartCore.Name).Error("Failed to add part from defect report")
			}
		}()
	}

	wg.Wait()

	logrus.WithFields(logrus.Fields{
		"brand": defectReportData.Brand,
		"model": defectReportData.Model,
		"parts": len(defectReportData.SelectedParts),
	}).Info("Successfully processed defect report event and added parts")

	return nil
}

func (c *RedisEventConsumer) handlePartIndexRequested(ctx context.Context, msg redis.XMessage) error {
	partIDStr, ok := msg.Values["part_id"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid part_index_requested event: missing part_id")
		return nil
	}

	partID, err := strconv.ParseInt(partIDStr, 10, 64)
	if err != nil {
		logrus.WithError(err).WithField("part_id", partIDStr).Error("Invalid part_id format in part_index_requested")
		return err
	}

	part, err := c.service.GetPartByID(ctx, partID)
	if err != nil {
		// Запчасть могла быть удалена до индексации
		logrus.WithField("part_id", partID).Info("Part not found for indexing, skipping")
		return nil
	}

	if err := IndexPart(part); err != nil {
		logrus.WithError(err).WithField("part_id", partID).Error("Failed to index part in Elasticsearch")
		return err
	}

	logrus.WithField("part_id", partID).Info("Successfully indexed part in Elasticsearch via Redis Streams")
	return nil
}

func (c *RedisEventConsumer) handlePartDeleteRequested(ctx context.Context, msg redis.XMessage) error {
	partIDStr, ok := msg.Values["part_id"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid part_delete_requested event: missing part_id")
		return nil
	}

	partID, err := strconv.ParseInt(partIDStr, 10, 64)
	if err != nil {
		logrus.WithError(err).WithField("part_id", partIDStr).Error("Invalid part_id format in part_delete_requested")
		return err
	}

	if err := DeletePartFromIndex(partID); err != nil {
		logrus.WithError(err).WithField("part_id", partID).Error("Failed to remove part from Elasticsearch index")
		return err
	}

	logrus.WithField("part_id", partID).Info("Successfully deleted part from Elasticsearch via Redis Streams")
	return nil
}
