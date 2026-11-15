package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8/esapi"

	"parts-service/internal/domain"
)

func partToESPart(part *domain.Part) ElasticsearchPart {
	return ElasticsearchPart{
		ID:                 part.ID,
		Name:               part.Name,
		Quantity:           part.Quantity,
		Description:        part.Description,
		Category:           part.Category,
		Price:              part.Price,
		Salesman:           part.Salesman,
		Location:           part.Location,
		Address:            part.Address,
		Status:             part.Status,
		Brand:              part.Brand,
		Model:              part.Model,
		Photos:             part.Photos,
		FrontRear:          part.FrontRear,
		LeftRight:          part.LeftRight,
		TopBottom:          part.TopBottom,
		OEMCode:            part.OEMCode,
		ManufacturerCode:   part.ManufacturerCode,
		Manufacturer:       part.Manufacturer,
		Color:              part.Color,
		Condition:          part.Condition,
		BodyBrand:          part.BodyBrand,
		EngineBrand:        part.EngineBrand,
		CarReleaseDate:     part.CarReleaseDate,
		CarReleasePeriod:   part.CarReleasePeriod,
		Transmission:       part.Transmission,
		TransmissionModel:  part.TransmissionModel,
		Drive:              part.Drive,
		Defect:             part.Defect,
		Number:             part.Number,
		SupplierCode:       part.SupplierCode,
		WearPercentage:     part.WearPercentage,
		VIN:                part.VIN,
		Season:             part.Season,
		Diameter:           part.Diameter,
		Width:              part.Width,
		Profile:            part.Profile,
		TireQuantity:       part.TireQuantity,
		Drilling:           part.Drilling,
		Offset:             part.Offset,
		CenterHoleDiameter: part.CenterHoleDiameter,
		TireModel:          part.TireModel,
	}
}

func IndexPart(part *domain.Part) error {
	if esClient == nil {
		return nil
	}

	esPart := partToESPart(part)
	data, err := json.Marshal(esPart)
	if err != nil {
		return fmt.Errorf("error marshaling part: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      "parts",
		DocumentID: fmt.Sprintf("%d", part.ID),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		RecordESError()
		return fmt.Errorf("error indexing part: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		RecordESError()
		return fmt.Errorf("error response: %s", res.String())
	}

	return nil
}

func BulkIndexParts(ctx context.Context, parts []domain.Part) error {
	if esClient == nil || len(parts) == 0 {
		return nil
	}

	const batchSize = 2000
	total := len(parts)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		batch := parts[i:end]
		if err := indexPartsBatch(ctx, batch, i+1, end); err != nil {
			return err
		}

		log.Printf("Bulk-индексация: успешно обработано %d из %d запчастей", end, total)
	}

	return RefreshPartsIndex(ctx)
}

func buildBulkPayload(parts []domain.Part) bytes.Buffer {
	var buf bytes.Buffer
	for i := range parts {
		part := &parts[i]
		meta := fmt.Sprintf(`{"index":{"_index":"parts","_id":"%d"}}`+"\n", part.ID)
		buf.WriteString(meta)

		esPart := partToESPart(part)
		docBytes, err := json.Marshal(esPart)
		if err != nil {
			log.Printf("Предупреждение: ошибка сериализации запчасти %d: %v", part.ID, err)
			continue
		}
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}
	return buf
}

type bulkItemOp struct {
	Status int `json:"status"`
	Error  struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	} `json:"error"`
}

type bulkResponseBody struct {
	Errors bool                    `json:"errors"`
	Items  []map[string]bulkItemOp `json:"items"`
}

func indexPartsBatch(ctx context.Context, batch []domain.Part, startIdx, endIdx int) error {
	buf := buildBulkPayload(batch)
	req := esapi.BulkRequest{
		Index: "parts",
		Body:  &buf,
	}

	res, err := req.Do(ctx, esClient)
	if err != nil {
		RecordESError()
		return fmt.Errorf("ошибка отправки bulk-запроса (%d-%d): %w", startIdx, endIdx, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		RecordESError()
		return fmt.Errorf("ошибка Elasticsearch при bulk-индексации (%d-%d): %s", startIdx, endIdx, res.String())
	}

	var bulkRes bulkResponseBody
	if err := json.NewDecoder(res.Body).Decode(&bulkRes); err != nil {
		log.Printf("Предупреждение: не удалось декодировать ответ bulk: %v", err)
		return nil
	}

	if bulkRes.Errors {
		logBulkErrors(bulkRes.Items)
	}
	return nil
}

func logBulkErrors(items []map[string]bulkItemOp) {
	errCount := 0
	for _, item := range items {
		for _, op := range item {
			if op.Error.Reason == "" {
				continue
			}
			errCount++
			if errCount <= 5 {
				RecordESError()
				log.Printf("Предупреждение: ошибка индексации документа (%d): %s: %s", op.Status, op.Error.Type, op.Error.Reason)
			}
		}
	}
	if errCount > 5 {
		log.Printf("Предупреждение: всего ошибок в пакете: %d", errCount)
	}
}

func DeletePartFromIndex(partID int64) error {
	if esClient == nil {
		return nil
	}

	req := esapi.DeleteRequest{
		Index:      "parts",
		DocumentID: fmt.Sprintf("%d", partID),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		RecordESError()
		return fmt.Errorf("error deleting part: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		RecordESError()
		return fmt.Errorf("error response: %s", res.String())
	}

	return nil
}

func BulkDeletePartsFromIndex(ctx context.Context, ids []int64) error {
	if esClient == nil || len(ids) == 0 {
		return nil
	}

	const batchSize = 2000
	total := len(ids)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		var buf bytes.Buffer
		for _, id := range ids[i:end] {
			buf.WriteString(fmt.Sprintf(`{"delete":{"_index":"parts","_id":"%d"}}`+"\n", id))
		}

		req := esapi.BulkRequest{
			Index: "parts",
			Body:  &buf,
		}

		res, err := req.Do(ctx, esClient)
		if err != nil {
			RecordESError()
			return fmt.Errorf("ошибка отправки bulk-delete запроса (%d-%d): %w", i+1, end, err)
		}
		_ = res.Body.Close()
	}

	return RefreshPartsIndex(ctx)
}
