package main

import "testing"

func TestEmbeddedCatalogAndDefectExpansion(t *testing.T) {
	catalog, err := LoadPartCatalog()
	if err != nil {
		t.Fatalf("load embedded catalog: %v", err)
	}
	if got, want := len(catalog.Parts), 2992; got != want {
		t.Fatalf("catalog parts count = %d, want %d", got, want)
	}

	report := DefectReportRequest{
		Brand:             "BMW",
		Model:             "E90",
		Year:              2011,
		VIN:               "TESTVIN123",
		BodyBrand:         "BMW E90",
		EngineBrand:       "N52B30",
		BodyColor:         "Черный",
		InteriorColor:     "Бежевый",
		Transmission:      "АКПП",
		TransmissionModel: "6HP19",
		Drive:             "Задний",
	}

	parts := catalog.ExpandDefectReport(report)
	if got, want := len(parts), len(catalog.Parts); got != want {
		t.Fatalf("expanded parts count = %d, want %d", got, want)
	}

	transmission := findExpandedPart(t, parts, "Трансмиссия")
	if transmission.Transmission != report.Transmission {
		t.Errorf("transmission = %q, want %q", transmission.Transmission, report.Transmission)
	}
	if transmission.TransmissionModel != report.TransmissionModel {
		t.Errorf("transmission model = %q, want %q", transmission.TransmissionModel, report.TransmissionModel)
	}
	if transmission.Drive != report.Drive {
		t.Errorf("transmission drive = %q, want %q", transmission.Drive, report.Drive)
	}

	steering := findExpandedPart(t, parts, "Рулевое управление")
	if steering.Drive != report.Drive {
		t.Errorf("steering drive = %q, want %q", steering.Drive, report.Drive)
	}
	if steering.TransmissionModel != "" {
		t.Errorf("steering transmission model leaked: %q", steering.TransmissionModel)
	}

	glass := findExpandedPart(t, parts, "Стекла")
	if glass.Drive != "" {
		t.Errorf("glass drive leaked: %q", glass.Drive)
	}
	if glass.VIN != report.VIN {
		t.Errorf("glass VIN = %q, want %q", glass.VIN, report.VIN)
	}
	if glass.CarReleaseDate != "2011" {
		t.Errorf("glass release date = %q, want %q", glass.CarReleaseDate, "2011")
	}

	for _, part := range parts {
		if part.Price != 0 {
			t.Errorf("part %q has price %v, want 0", part.Name, part.Price)
		}
		if part.Quantity != 0 {
			t.Errorf("part %q has quantity %d, want 0", part.Name, part.Quantity)
		}
	}
}

func TestCatalogValidationRejectsUnknownDependencyCategory(t *testing.T) {
	catalog, err := LoadPartCatalog()
	if err != nil {
		t.Fatalf("load embedded catalog: %v", err)
	}
	catalog.ReportBindings[0].Categories = []string{"Несуществующая категория"}
	if err := validatePartCatalog(catalog); err == nil {
		t.Fatal("expected an unknown dependency category to fail validation")
	}
}

func TestCatalogValidationRejectsSelectWithoutOptions(t *testing.T) {
	catalog, err := LoadPartCatalog()
	if err != nil {
		t.Fatalf("load embedded catalog: %v", err)
	}
	for index := range catalog.Attributes {
		if catalog.Attributes[index].Code == "drive" {
			catalog.Attributes[index].Options = nil
			break
		}
	}
	if err := validatePartCatalog(catalog); err == nil {
		t.Fatal("expected a select attribute without options to fail validation")
	}
}

func findExpandedPart(t *testing.T, parts []DefectReportPart, category string) DefectReportPart {
	t.Helper()
	for _, part := range parts {
		if part.Category == category {
			return part
		}
	}
	t.Fatalf("expanded catalog has no part in category %q", category)
	return DefectReportPart{}
}
