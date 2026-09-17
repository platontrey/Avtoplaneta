package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	errPartCatalogUnavailable     = errors.New("каталог шаблонов запчастей недоступен")
	errDefectPublisherUnavailable = errors.New("очередь дефектных ведомостей недоступна")
)

const (
	defectReportEventVersion = 1
	defectReportEventType    = "defect_report_created"
	defectReportStream       = "events:orders"
)

// defectReportUnavailable отличает временную недоступность инфраструктуры
// (каталог не загружен, очередь не сконфигурирована) от ошибки публикации:
// первое — 503 / codes.Unavailable, второе — 500 / codes.Internal.
func defectReportUnavailable(err error) bool {
	return errors.Is(err, errPartCatalogUnavailable) || errors.Is(err, errDefectPublisherUnavailable)
}

type DefectReportEventPublisher interface {
	PublishDefectReport(context.Context, DefectReportRequest) error
}

type RedisDefectReportEventPublisher struct {
	client *redis.Client
}

func NewRedisDefectReportEventPublisher(client *redis.Client) *RedisDefectReportEventPublisher {
	return &RedisDefectReportEventPublisher{client: client}
}

func (p *RedisDefectReportEventPublisher) PublishDefectReport(ctx context.Context, report DefectReportRequest) error {
	if p == nil || p.client == nil {
		return errDefectPublisherUnavailable
	}

	payload, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("сериализовать дефектную ведомость: %w", err)
	}

	if err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: defectReportStream,
		Values: map[string]interface{}{
			"type": defectReportEventType,
			"data": string(payload),
		},
	}).Err(); err != nil {
		return fmt.Errorf("опубликовать дефектную ведомость: %w", err)
	}

	return nil
}

// DefectReportWorkflow — единственное место, где входные данные ведомости
// превращаются в запчасти и помещаются в очередь. HTTP и gRPC служат
// только транспортными адаптерами.
type DefectReportWorkflow struct {
	catalog   *PartCatalog
	publisher DefectReportEventPublisher
}

func NewDefectReportWorkflow(catalog *PartCatalog, publisher DefectReportEventPublisher) *DefectReportWorkflow {
	return &DefectReportWorkflow{catalog: catalog, publisher: publisher}
}

// prepare разворачивает набор запчастей по каталогу.
//
// allowLegacyClientParts=true означает «довериться selectedParts из запроса».
// TODO(legacy): убрать параметр и всегда строить набор сервером. Web и Flutter
// уже присылают только характеристики автомобиля; параметр нужен лишь старым
// установленным сборкам мобильного приложения. Снять после того, как они вымоются.
func (w *DefectReportWorkflow) prepare(report *DefectReportRequest, allowLegacyClientParts bool) error {
	if w == nil || w.catalog == nil {
		return errPartCatalogUnavailable
	}

	if !allowLegacyClientParts || len(report.SelectedParts) == 0 {
		report.SelectedParts = w.catalog.ExpandDefectReport(*report)
	} else {
		w.catalog.ApplyBindingsToParts(report.SelectedParts, *report)
	}
	report.CatalogVersion = w.catalog.Version
	report.EventVersion = defectReportEventVersion
	return nil
}

func (w *DefectReportWorkflow) Preview(report DefectReportRequest) (DefectReportRequest, error) {
	// Preview не доверяет selectedParts, присланным клиентом.
	report.SelectedParts = nil
	if err := w.prepare(&report, false); err != nil {
		return DefectReportRequest{}, err
	}
	return report, nil
}

func (w *DefectReportWorkflow) Enqueue(ctx context.Context, report *DefectReportRequest, allowLegacyClientParts bool) error {
	if err := w.prepare(report, allowLegacyClientParts); err != nil {
		return err
	}
	if w.publisher == nil {
		return errDefectPublisherUnavailable
	}
	return w.publisher.PublishDefectReport(ctx, *report)
}
