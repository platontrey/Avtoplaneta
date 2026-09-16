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
		CarReleasePeriod:  "2005-2011",
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

	// 1. Трансмиссия: должна содержать car_release_period, transmission, transmission_model, drive
	transmission := findExpandedPart(t, parts, "Трансмиссия")
	if transmission.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("transmission release period = %q, want %q", transmission.CarReleasePeriod, report.CarReleasePeriod)
	}
	if transmission.Transmission != report.Transmission {
		t.Errorf("transmission = %q, want %q", transmission.Transmission, report.Transmission)
	}
	if transmission.TransmissionModel != report.TransmissionModel {
		t.Errorf("transmission model = %q, want %q", transmission.TransmissionModel, report.TransmissionModel)
	}
	if transmission.Drive != report.Drive {
		t.Errorf("transmission drive = %q, want %q", transmission.Drive, report.Drive)
	}

	// 2. Двигатель: должен содержать car_release_period, drive, но НЕ transmission и transmission_model
	engine := findExpandedPart(t, parts, "Двигатель")
	if engine.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("engine release period = %q, want %q", engine.CarReleasePeriod, report.CarReleasePeriod)
	}
	if engine.Drive != report.Drive {
		t.Errorf("engine drive = %q, want %q", engine.Drive, report.Drive)
	}
	if engine.Transmission != "" {
		t.Errorf("engine transmission leaked: %q", engine.Transmission)
	}
	if engine.TransmissionModel != "" {
		t.Errorf("engine transmission model leaked: %q", engine.TransmissionModel)
	}

	// 3. Подвеска передних колес: должна содержать transmission_model и drive, но НЕ transmission
	suspFront := findExpandedPart(t, parts, "Подвеска передних колес")
	if suspFront.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("suspension front release period = %q, want %q", suspFront.CarReleasePeriod, report.CarReleasePeriod)
	}
	if suspFront.TransmissionModel != report.TransmissionModel {
		t.Errorf("suspension front transmission model = %q, want %q", suspFront.TransmissionModel, report.TransmissionModel)
	}
	if suspFront.Drive != report.Drive {
		t.Errorf("suspension front drive = %q, want %q", suspFront.Drive, report.Drive)
	}
	if suspFront.Transmission != "" {
		t.Errorf("suspension front transmission leaked: %q", suspFront.Transmission)
	}

	// 4. Подвеска ДВС/КПП: должна содержать transmission_model и drive, но НЕ transmission
	suspEngine := findExpandedPart(t, parts, "Подвеска ДВС/КПП")
	if suspEngine.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("suspension engine release period = %q, want %q", suspEngine.CarReleasePeriod, report.CarReleasePeriod)
	}
	if suspEngine.TransmissionModel != report.TransmissionModel {
		t.Errorf("suspension engine transmission model = %q, want %q", suspEngine.TransmissionModel, report.TransmissionModel)
	}
	if suspEngine.Drive != report.Drive {
		t.Errorf("suspension engine drive = %q, want %q", suspEngine.Drive, report.Drive)
	}
	if suspEngine.Transmission != "" {
		t.Errorf("suspension engine transmission leaked: %q", suspEngine.Transmission)
	}

	// 5. Рулевое управление: должно содержать drive, но НЕ transmission_model и transmission
	steering := findExpandedPart(t, parts, "Рулевое управление")
	if steering.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("steering release period = %q, want %q", steering.CarReleasePeriod, report.CarReleasePeriod)
	}
	if steering.Drive != report.Drive {
		t.Errorf("steering drive = %q, want %q", steering.Drive, report.Drive)
	}
	if steering.TransmissionModel != "" {
		t.Errorf("steering transmission model leaked: %q", steering.TransmissionModel)
	}
	if steering.Transmission != "" {
		t.Errorf("steering transmission leaked: %q", steering.Transmission)
	}

	// 6. Стекла: должны содержать car_release_period, VIN, год, но НЕ drive, transmission_model, transmission
	glass := findExpandedPart(t, parts, "Стекла")
	if glass.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("glass release period = %q, want %q", glass.CarReleasePeriod, report.CarReleasePeriod)
	}
	if glass.Drive != "" {
		t.Errorf("glass drive leaked: %q", glass.Drive)
	}
	if glass.Transmission != "" {
		t.Errorf("glass transmission leaked: %q", glass.Transmission)
	}
	if glass.TransmissionModel != "" {
		t.Errorf("glass transmission model leaked: %q", glass.TransmissionModel)
	}
	if glass.VIN != report.VIN {
		t.Errorf("glass VIN = %q, want %q", glass.VIN, report.VIN)
	}
	if glass.CarReleaseDate != "2011" {
		t.Errorf("glass release date = %q, want %q", glass.CarReleaseDate, "2011")
	}

	// 7. Диски и шины: должны содержать car_release_period, но НЕ drive, transmission, transmission_model
	wheels := findExpandedPart(t, parts, "Диски и шины")
	if wheels.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("wheels release period = %q, want %q", wheels.CarReleasePeriod, report.CarReleasePeriod)
	}
	if wheels.Drive != "" {
		t.Errorf("wheels drive leaked: %q", wheels.Drive)
	}

	// Проверяем, что car_release_period применился ко ВСЕМ созданным запчастям
	for _, part := range parts {
		if part.CarReleasePeriod != report.CarReleasePeriod {
			t.Errorf("part %q (category %q) has car_release_period = %q, want %q", part.Name, part.Category, part.CarReleasePeriod, report.CarReleasePeriod)
			break
		}
		if part.Price != 0 {
			t.Errorf("part %q has price %v, want 0", part.Name, part.Price)
		}
		if part.Quantity != 0 {
			t.Errorf("part %q has quantity %d, want 0", part.Name, part.Quantity)
		}
	}
}

func TestApplyBindingsToParts(t *testing.T) {
	catalog, err := LoadPartCatalog()
	if err != nil {
		t.Fatalf("load embedded catalog: %v", err)
	}

	report := DefectReportRequest{
		Brand:             "BMW",
		Model:             "E90",
		Year:              2011,
		CarReleasePeriod:  "2005-2011",
		VIN:               "TESTVIN123",
		Transmission:      "АКПП",
		TransmissionModel: "6HP19",
		Drive:             "Задний",
	}

	// Имитируем запчасти, отправленные клиентом без характеристик
	parts := []DefectReportPart{
		{Name: "Коробка передач", Category: "Трансмиссия"},
		{Name: "Рычаг подвески", Category: "Подвеска передних колес"},
		{Name: "Глушитель", Category: "Выхлопная система"},
		{Name: "Лобовое стекло", Category: "Стекла"},
	}

	catalog.ApplyBindingsToParts(parts, report)

	// Трансмиссия должна получить все 4 характеристики
	if parts[0].CarReleasePeriod != "2005-2011" {
		t.Errorf("parts[0].CarReleasePeriod = %q, want 2005-2011", parts[0].CarReleasePeriod)
	}
	if parts[0].Transmission != "АКПП" {
		t.Errorf("parts[0].Transmission = %q, want АКПП", parts[0].Transmission)
	}
	if parts[0].TransmissionModel != "6HP19" {
		t.Errorf("parts[0].TransmissionModel = %q, want 6HP19", parts[0].TransmissionModel)
	}
	if parts[0].Drive != "Задний" {
		t.Errorf("parts[0].Drive = %q, want Задний", parts[0].Drive)
	}

	// Подвеска передних колес: car_release_period, transmission_model, drive, но НЕ transmission
	if parts[1].CarReleasePeriod != "2005-2011" {
		t.Errorf("parts[1].CarReleasePeriod = %q, want 2005-2011", parts[1].CarReleasePeriod)
	}
	if parts[1].Transmission != "" {
		t.Errorf("parts[1].Transmission leaked: %q", parts[1].Transmission)
	}
	if parts[1].TransmissionModel != "6HP19" {
		t.Errorf("parts[1].TransmissionModel = %q, want 6HP19", parts[1].TransmissionModel)
	}
	if parts[1].Drive != "Задний" {
		t.Errorf("parts[1].Drive = %q, want Задний", parts[1].Drive)
	}

	// Выхлопная система: car_release_period, drive, но НЕ transmission, transmission_model
	if parts[2].CarReleasePeriod != "2005-2011" {
		t.Errorf("parts[2].CarReleasePeriod = %q, want 2005-2011", parts[2].CarReleasePeriod)
	}
	if parts[2].Drive != "Задний" {
		t.Errorf("parts[2].Drive = %q, want Задний", parts[2].Drive)
	}
	if parts[2].Transmission != "" {
		t.Errorf("parts[2].Transmission leaked: %q", parts[2].Transmission)
	}
	if parts[2].TransmissionModel != "" {
		t.Errorf("parts[2].TransmissionModel leaked: %q", parts[2].TransmissionModel)
	}

	// Стекла: car_release_period, но НЕ drive, transmission, transmission_model
	if parts[3].CarReleasePeriod != "2005-2011" {
		t.Errorf("parts[3].CarReleasePeriod = %q, want 2005-2011", parts[3].CarReleasePeriod)
	}
	if parts[3].Drive != "" {
		t.Errorf("parts[3].Drive leaked: %q", parts[3].Drive)
	}
	if parts[3].Transmission != "" {
		t.Errorf("parts[3].Transmission leaked: %q", parts[3].Transmission)
	}
	if parts[3].TransmissionModel != "" {
		t.Errorf("parts[3].TransmissionModel leaked: %q", parts[3].TransmissionModel)
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
