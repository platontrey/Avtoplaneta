package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"export-service/internal/config"
	"export-service/internal/domain"
	"export-service/internal/repository"
	"export-service/internal/service"
)

// fakeSource считает обращения к зависимости InventorySource
type fakeSource struct {
	version     string
	versionErr  error
	parts       []domain.Part
	inns        map[int64]string
	innsErr     error
	versionHits int
	partsHits   int
}

func (s *fakeSource) InventoryVersion(context.Context) (string, error) {
	s.versionHits++
	return s.version, s.versionErr
}

func (s *fakeSource) PartsForExport(context.Context) ([]domain.Part, error) {
	s.partsHits++
	return s.parts, nil
}

func (s *fakeSource) GetPart(_ context.Context, partID int64) (*domain.Part, error) {
	for _, p := range s.parts {
		if p.ID == partID {
			return &p, nil
		}
	}
	return nil, errors.New("деталь не найдена")
}

func (s *fakeSource) SellerINNs(context.Context) (map[int64]string, error) {
	return s.inns, s.innsErr
}

func newTestPriceListService(t *testing.T, source service.InventorySource) *service.PriceListService {
	t.Helper()
	dir := t.TempDir()
	store, err := repository.NewBoltSnapshotRepository(filepath.Join(dir, "catalog.db"))
	if err != nil {
		t.Fatalf("NewBoltSnapshotRepository: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return service.NewPriceListService(source, store, &config.Config{ExportDir: dir})
}

// Первый вызов собирает файл, второй при том же отпечатке склада не должен вычитывать запчасти заново
func TestBuildSkipsWhenInventoryUnchanged(t *testing.T) {
	source := &fakeSource{version: "17-42", parts: []domain.Part{{Name: "Деталь", Status: true}}}
	svc := newTestPriceListService(t, source)

	if _, rebuilt, err := svc.Build(context.Background()); err != nil || !rebuilt {
		t.Fatalf("первая сборка: rebuilt=%v err=%v", rebuilt, err)
	}
	if _, rebuilt, err := svc.Build(context.Background()); err != nil || rebuilt {
		t.Fatalf("вторая сборка должна была пропуститься: rebuilt=%v err=%v", rebuilt, err)
	}
	if source.partsHits != 1 {
		t.Fatalf("склад вычитан %d раз вместо одного", source.partsHits)
	}
}

// Отпечаток изменился — склад менялся, и файл пересобирается
func TestBuildRebuildsWhenInventoryChanged(t *testing.T) {
	source := &fakeSource{version: "17-42", parts: []domain.Part{{Name: "Деталь", Status: true}}}
	svc := newTestPriceListService(t, source)

	if _, _, err := svc.Build(context.Background()); err != nil {
		t.Fatalf("первая сборка: %v", err)
	}
	source.version = "18-43"
	if _, rebuilt, err := svc.Build(context.Background()); err != nil || !rebuilt {
		t.Fatalf("после изменения склада ожидалась пересборка: rebuilt=%v err=%v", rebuilt, err)
	}
}

// Файл удалили, а отпечаток остался. Прайс-лист должен собраться снова.
func TestBuildRebuildsWhenFileMissing(t *testing.T) {
	source := &fakeSource{version: "17-42", parts: []domain.Part{{Name: "Деталь", Status: true}}}
	svc := newTestPriceListService(t, source)

	if _, _, err := svc.Build(context.Background()); err != nil {
		t.Fatalf("первая сборка: %v", err)
	}
	if err := os.Remove(svc.Path()); err != nil {
		t.Fatalf("удалить прайс-лист: %v", err)
	}
	if _, rebuilt, err := svc.Build(context.Background()); err != nil || !rebuilt {
		t.Fatalf("пропавший файл должен собираться заново: rebuilt=%v err=%v", rebuilt, err)
	}
}

// Отпечаток не прочитался — сборка выполняется безусловно
func TestBuildProceedsWhenVersionUnavailable(t *testing.T) {
	source := &fakeSource{versionErr: errors.New("parts-service недоступен"), parts: []domain.Part{{Name: "Деталь", Status: true}}}
	svc := newTestPriceListService(t, source)

	if _, rebuilt, err := svc.Build(context.Background()); err != nil || !rebuilt {
		t.Fatalf("без отпечатка ожидалась безусловная сборка: rebuilt=%v err=%v", rebuilt, err)
	}
	if _, rebuilt, err := svc.Build(context.Background()); err != nil || !rebuilt {
		t.Fatalf("без отпечатка пропуск невозможен: rebuilt=%v err=%v", rebuilt, err)
	}
}

// ИНН продавцов недоступен — сборка выдерживает сбой auth-service
func TestBuildSurvivesMissingSellerDirectory(t *testing.T) {
	source := &fakeSource{
		version: "17-42",
		parts:   []domain.Part{{Name: "Деталь", SellerID: 5, Status: true}},
		innsErr: errors.New("auth-service недоступен"),
	}
	svc := newTestPriceListService(t, source)

	meta, rebuilt, err := svc.Build(context.Background())
	if err != nil || !rebuilt {
		t.Fatalf("сборка должна была пройти: rebuilt=%v err=%v", rebuilt, err)
	}
	if meta.PartsCount != 1 {
		t.Fatalf("ожидалась 1 запчасть, получено %d", meta.PartsCount)
	}
	if _, err := os.Stat(svc.Path()); err != nil {
		t.Fatalf("файл прайс-листа не создан: %v", err)
	}
}
