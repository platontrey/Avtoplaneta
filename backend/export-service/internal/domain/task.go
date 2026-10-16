package domain

import "time"

// ExportTask описывает задачу на экспорт каталога в фоновой очереди
type ExportTask struct {
	TaskID      string    `json:"task_id"`
	Target      string    `json:"target"` // "drom", "xml"
	TriggeredBy string    `json:"triggered_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ExportStatus описывает состояние и результат выполнения задачи экспорта
type ExportStatus struct {
	TaskID      string     `json:"task_id"`
	Status      string     `json:"status"` // "idle", "queued", "running", "success", "failed"
	TriggeredBy string     `json:"triggered_by,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	PartsCount  int        `json:"parts_count,omitempty"`
	Error       string     `json:"error,omitempty"`
}
