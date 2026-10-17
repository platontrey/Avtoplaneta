package outbox

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]Event, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Event), args.Error(1)
}

func (m *MockStorage) MarkOutboxEventProcessed(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockStorage) MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error {
	return m.Called(ctx, id, lastErr).Error(0)
}

func (m *MockStorage) CleanupProcessedOutboxEvents(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockMetrics struct {
	mu        sync.Mutex
	processed map[string]int
	failed    map[string]int
}

func newMockMetrics() *MockMetrics {
	return &MockMetrics{
		processed: make(map[string]int),
		failed:    make(map[string]int),
	}
}

func (m *MockMetrics) RecordEventProcessed(eventType string, durationSec float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processed[eventType]++
}

func (m *MockMetrics) RecordEventFailed(eventType string, durationSec float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failed[eventType]++
}

func TestDispatcher_ProcessEvent_Success(t *testing.T) {
	storage := new(MockStorage)
	metrics := newMockMetrics()

	event := Event{
		ID:            101,
		AggregateType: "order",
		AggregateID:   505,
		EventType:     "order_created",
		Payload:       []byte(`{"order_id": 505}`),
	}

	storage.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]Event{event}, nil).Once()
	storage.On("MarkOutboxEventProcessed", mock.Anything, int64(101)).Return(nil).Once()
	storage.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]Event{}, nil)

	handled := false
	handler := func(ctx context.Context, e Event) error {
		assert.Equal(t, int64(101), e.ID)
		assert.Equal(t, "order_created", e.EventType)
		handled = true
		return nil
	}

	dispatcher := NewDispatcher(nil, storage, handler, Options{
		Name:         "Orders",
		BatchSize:    50,
		PollInterval: 10 * time.Millisecond,
		Metrics:      metrics,
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, handled)

	storage.AssertExpectations(t)
	assert.Equal(t, 1, metrics.processed["order_created"])
}

func TestDispatcher_ProcessEvent_Failure(t *testing.T) {
	storage := new(MockStorage)
	metrics := newMockMetrics()

	event := Event{
		ID:            102,
		AggregateType: "part",
		AggregateID:   777,
		EventType:     "part_created",
		Payload:       []byte(`{"part_id": 777}`),
	}

	storage.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]Event{event}, nil).Once()
	storage.On("MarkOutboxEventFailed", mock.Anything, int64(102), "es connection timeout").Return(nil).Once()
	storage.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]Event{}, nil)

	handler := func(ctx context.Context, e Event) error {
		return errors.New("es connection timeout")
	}

	dispatcher := NewDispatcher(nil, storage, handler, Options{
		Name:         "Parts",
		BatchSize:    50,
		PollInterval: 10 * time.Millisecond,
		Metrics:      metrics,
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	storage.AssertExpectations(t)
	assert.Equal(t, 1, metrics.failed["part_created"])
}

func TestDispatcher_CleanupLoop(t *testing.T) {
	storage := new(MockStorage)

	storage.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]Event{}, nil)
	storage.On("CleanupProcessedOutboxEvents", mock.Anything).Return(nil)

	handler := func(ctx context.Context, e Event) error {
		return nil
	}

	dispatcher := NewDispatcher(nil, storage, handler, Options{
		Name:            "Test",
		CleanupInterval: 15 * time.Millisecond,
		PollInterval:    100 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(40 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	storage.AssertCalled(t, "CleanupProcessedOutboxEvents", mock.Anything)
}
