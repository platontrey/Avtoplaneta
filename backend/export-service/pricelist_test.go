package main

import (
	"context"
	"errors"
	"os"
	"testing"
)

// fakeSource считает обращения: именно по ним видно, сэкономил сборщик работу
// или всё-таки сходил за всем складом.
type fakeSource struct {
	version     string
	versionErr  error
	parts       []Part
	inns        map[int64]string
	innsErr     error
	versionHits int
	partsHits   int
}

func (s *fakeSource) InventoryVersion(context.Context) (string, error) {
	s.versionHits++
	return s.version, s.versionErr
}

func (s *fakeSource) PartsForExport(context.Context) ([]Part, error) {
	s.partsHits++
	return s.parts, nil
}

func (s *fakeSource) SellerINNs(context.Context) (map[int64]string, error) {
	return s.inns, s.innsErr
}

func newBuilder(t *testing.T, source InventorySource) *PriceListBuilder {
	t.Helper()
	return NewPriceListBuilder(source, &Config{ExportDir: t.TempDir()})
}

// Первый вызов собирает файл, второй при том же отпечатке склада не должен
// вычитывать запчасти заново — ради этого отпечаток и заводился.
func TestBuildSkipsWhenInventoryUnchanged(t *testing.T) {
	source := &fakeSource{version: "17-42", parts: []Part{{Name: "Деталь", Status: true}}}
	builder := newBuilder(t, source)

	if _, rebuilt, err := builder.Build(context.Background()); err != nil || !rebuilt {
		t.Fatalf("первая сборка: rebuilt=%v err=%v", rebuilt, err)
	}
	if _, rebuilt, err := builder.Build(context.Background()); err != nil || rebuilt {
		t.Fatalf("вторая сборка должна была пропуститься: rebuilt=%v err=%v", rebuilt, err)
	}
	if source.partsHits != 1 {
		t.Fatalf("склад вычитан %d раз вместо одного", source.partsHits)
	}
}

// Отпечаток изменился — значит, склад менялся, и файл пересобирается.
func TestBuildRebuildsWhenInventoryChanged(t *testing.T) {
	source := &fakeSource{version: "17-42", parts: []Part{{Name: "Деталь", Status: true}}}
	builder := newBuilder(t, source)

	if _, _, err := builder.Build(context.Background()); err != nil {
		t.Fatalf("первая сборка: %v", err)
	}
	source.version = "18-43"
	if _, rebuilt, err := builder.Build(context.Background()); err != nil || !rebuilt {
		t.Fatalf("после изменения склада ожидалась пересборка: rebuilt=%v err=%v", rebuilt, err)
	}
}

// Файл удалили, а отпечаток остался. Доверять отпечатку в одиночку нельзя:
// прайс-лист должен появиться снова.
func TestBuildRebuildsWhenFileMissing(t *testing.T) {
	source := &fakeSource{version: "17-42", parts: []Part{{Name: "Деталь", Status: true}}}
	builder := newBuilder(t, source)

	if _, _, err := builder.Build(context.Background()); err != nil {
		t.Fatalf("первая сборка: %v", err)
	}
	if err := os.Remove(builder.Path()); err != nil {
		t.Fatalf("удалить прайс-лист: %v", err)
	}
	if _, rebuilt, err := builder.Build(context.Background()); err != nil || !rebuilt {
		t.Fatalf("пропавший файл должен собираться заново: rebuilt=%v err=%v", rebuilt, err)
	}
}

// Отпечаток не прочитался — это повод сделать лишнюю работу, а не отказать
// площадке в прайс-листе.
func TestBuildProceedsWhenVersionUnavailable(t *testing.T) {
	source := &fakeSource{versionErr: errors.New("parts-service недоступен"), parts: []Part{{Name: "Деталь", Status: true}}}
	builder := newBuilder(t, source)

	if _, rebuilt, err := builder.Build(context.Background()); err != nil || !rebuilt {
		t.Fatalf("без отпечатка ожидалась безусловная сборка: rebuilt=%v err=%v", rebuilt, err)
	}
	if _, rebuilt, err := builder.Build(context.Background()); err != nil || !rebuilt {
		t.Fatalf("без отпечатка пропуск невозможен: rebuilt=%v err=%v", rebuilt, err)
	}
}

// ИНН — обязательный реквизит, но недоступный auth-service не должен оставлять
// Drom вообще без прайс-листа.
func TestBuildSurvivesMissingSellerDirectory(t *testing.T) {
	source := &fakeSource{
		version: "17-42",
		parts:   []Part{{Name: "Деталь", SellerID: 5, Status: true}},
		innsErr: errors.New("auth-service недоступен"),
	}
	builder := newBuilder(t, source)

	meta, rebuilt, err := builder.Build(context.Background())
	if err != nil || !rebuilt {
		t.Fatalf("сборка должна была пройти: rebuilt=%v err=%v", rebuilt, err)
	}
	if meta.PartsCount != 1 {
		t.Fatalf("ожидалась 1 запчасть, получено %d", meta.PartsCount)
	}
	if _, err := os.Stat(builder.Path()); err != nil {
		t.Fatalf("файл прайс-листа не создан: %v", err)
	}
}
