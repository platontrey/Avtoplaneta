package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestExportQueue_EnqueueAndConflict(t *testing.T) {
	builder := &PriceListBuilder{}
	queue := NewExportQueue(nil, builder, &Config{})

	ctx := context.Background()

	// 1. Первая постановка в очередь должна пройти успешно
	task, err := queue.Enqueue(ctx, "drom", "127.0.0.1")
	if err != nil {
		t.Fatalf("ожидался успешный enqueue, получена ошибка: %v", err)
	}
	if task.TaskID == "" {
		t.Fatal("task_id не должен быть пустым")
	}

	status := queue.GetStatus(ctx)
	if status.Status != "queued" {
		t.Fatalf("ожидался статус 'queued', получен: %s", status.Status)
	}

	// 2. Повторная попытка пока задача в очереди/выполняется должна вернуть ошибку конфликта
	_, err = queue.Enqueue(ctx, "drom", "127.0.0.1")
	if err == nil {
		t.Fatal("ожидалась ошибка конфликта при параллельном enqueue")
	}

	// 3. После сброса running флаг освобождается
	queue.running.Store(false)
	task2, err := queue.Enqueue(ctx, "drom", "127.0.0.1")
	if err != nil {
		t.Fatalf("после освобождения ожидался успешный enqueue: %v", err)
	}
	if task2.TaskID == task.TaskID {
		t.Fatal("task_id должен быть уникальным")
	}
}

func TestExportHandler_HTTP202AndStatus(t *testing.T) {
	builder := &PriceListBuilder{}
	queue := NewExportQueue(nil, builder, &Config{})
	handler := NewHandler(builder, queue, &Config{})

	router := gin.New()
	SetupRoutes(router, handler)

	// 1. Первый POST /api/export/drom -> 202 Accepted
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/api/export/drom", nil)
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusAccepted {
		t.Fatalf("ожидался HTTP 202, получен: %d, body: %s", w1.Code, w1.Body.String())
	}

	var resp1 map[string]interface{}
	if err := json.Unmarshal(w1.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("ошибка разбора JSON ответа: %v", err)
	}
	if resp1["status"] != "queued" {
		t.Fatalf("ожидался status=queued, получен: %v", resp1["status"])
	}
	if resp1["task_id"] == nil || resp1["task_id"] == "" {
		t.Fatalf("task_id отсутствует в ответе")
	}

	// 2. Второй параллельный POST /api/export/drom -> 409 Conflict
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/api/export/drom", nil)
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("ожидался HTTP 409 Conflict, получен: %d, body: %s", w2.Code, w2.Body.String())
	}

	// 3. GET /api/export/status -> 200 OK со статусом
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/api/export/status", nil)
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("ожидался HTTP 200 OK для /api/export/status, получен: %d", w3.Code)
	}
	var statusResp ExportStatus
	if err := json.Unmarshal(w3.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("ошибка парсинга статуса: %v", err)
	}
	if statusResp.TaskID != resp1["task_id"] {
		t.Fatalf("task_id в статусе не совпадает: %s vs %s", statusResp.TaskID, resp1["task_id"])
	}
}
