package main

import (
	"strings"
	"testing"
)

func TestBrandSlugAcceptsOnlyBrandPages(t *testing.T) {
	good := map[string]string{
		"/catalog/toyota/":                        "toyota",
		"https://www.drom.ru/catalog/alfa_romeo/": "alfa_romeo",
		"//www.drom.ru/catalog/mercedes-benz/":    "mercedes-benz",
	}
	for href, want := range good {
		got, ok := brandSlug(href)
		if !ok || got != want {
			t.Fatalf("brandSlug(%q) = %q, %v; ожидалось %q, true", href, got, ok, want)
		}
	}

	bad := []string{
		"/catalog/",                           // индекс
		"/catalog/toyota/camry/",              // это модель
		"/catalog/all/",                       // служебный раздел
		"/catalog/toyota/?page=2",             // с параметрами не работаем
		"/reviews/toyota/",                    // другой раздел
		"https://example.com/catalog/toyota/", // чужой домен
	}
	for _, href := range bad {
		if _, ok := brandSlug(href); ok {
			t.Fatalf("brandSlug(%q) принял ссылку, которую не должен", href)
		}
	}
}

func TestModelSlugStaysInsideItsBrand(t *testing.T) {
	if got, ok := modelSlug("/catalog/toyota/land_cruiser_prado/", "toyota"); !ok || got != "land_cruiser_prado" {
		t.Fatalf("modelSlug вернул %q, %v", got, ok)
	}
	if _, ok := modelSlug("/catalog/lexus/rx/", "toyota"); ok {
		t.Fatal("modelSlug принял модель чужой марки")
	}
	if _, ok := modelSlug("/catalog/toyota/camry/1990/", "toyota"); ok {
		t.Fatal("modelSlug принял ссылку глубже модели")
	}
}

// Версия завязана на содержимое: пустой прогон не должен менять ETag у клиентов.
func TestVersionDependsOnContentOnly(t *testing.T) {
	first := []vehicleBrand{{Name: "Toyota", Slug: "toyota", Models: []vehicleModel{{Name: "Camry", Slug: "camry"}}}}
	second := []vehicleBrand{{Name: "Toyota", Slug: "toyota", Models: []vehicleModel{{Name: "Camry", Slug: "camry"}}}}
	third := []vehicleBrand{{Name: "Toyota", Slug: "toyota", Models: []vehicleModel{{Name: "Corolla", Slug: "corolla"}}}}

	if version(first) != version(second) {
		t.Fatal("одинаковые данные дали разные версии")
	}
	if version(first) == version(third) {
		t.Fatal("разные данные дали одинаковую версию")
	}
	if parts := strings.Split(version(first), "."); len(parts) != 2 || len(parts[1]) != 8 {
		t.Fatalf("неожиданный формат версии: %q", version(first))
	}
}

func TestSortCatalogIsDeterministic(t *testing.T) {
	catalog := vehicleCatalog{Brands: []vehicleBrand{
		{Name: "Toyota", Models: []vehicleModel{{Name: "Corolla"}, {Name: "Camry"}}},
		{Name: "audi", Models: []vehicleModel{{Name: "A6"}, {Name: "A4"}}},
	}}
	sortCatalog(&catalog)

	if catalog.Brands[0].Name != "audi" || catalog.Brands[1].Name != "Toyota" {
		t.Fatalf("марки отсортированы неверно: %v", catalog.Brands)
	}
	if catalog.Brands[0].Models[0].Name != "A4" || catalog.Brands[1].Models[0].Name != "Camry" {
		t.Fatal("модели отсортированы неверно")
	}
}
