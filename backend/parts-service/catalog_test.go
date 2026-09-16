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

	// 2. Двигатель: все автомобильные характеристики копируются в каждую запчасть
	engine := findExpandedPart(t, parts, "Двигатель")
	if engine.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("engine release period = %q, want %q", engine.CarReleasePeriod, report.CarReleasePeriod)
	}
	if engine.Drive != report.Drive {
		t.Errorf("engine drive = %q, want %q", engine.Drive, report.Drive)
	}
	if engine.Transmission != report.Transmission {
		t.Errorf("engine transmission = %q, want %q", engine.Transmission, report.Transmission)
	}
	if engine.TransmissionModel != report.TransmissionModel {
		t.Errorf("engine transmission model = %q, want %q", engine.TransmissionModel, report.TransmissionModel)
	}

	// 3. Подвеска передних колес также получает все трансмиссионные поля
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
	if suspFront.Transmission != report.Transmission {
		t.Errorf("suspension front transmission = %q, want %q", suspFront.Transmission, report.Transmission)
	}

	// 4. Подвеска ДВС/КПП
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
	if suspEngine.Transmission != report.Transmission {
		t.Errorf("suspension engine transmission = %q, want %q", suspEngine.Transmission, report.Transmission)
	}

	// 5. Рулевое управление
	steering := findExpandedPart(t, parts, "Рулевое управление")
	if steering.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("steering release period = %q, want %q", steering.CarReleasePeriod, report.CarReleasePeriod)
	}
	if steering.Drive != report.Drive {
		t.Errorf("steering drive = %q, want %q", steering.Drive, report.Drive)
	}
	if steering.TransmissionModel != report.TransmissionModel {
		t.Errorf("steering transmission model = %q, want %q", steering.TransmissionModel, report.TransmissionModel)
	}
	if steering.Transmission != report.Transmission {
		t.Errorf("steering transmission = %q, want %q", steering.Transmission, report.Transmission)
	}

	// 6. Стекла: должны содержать все автомобильные данные
	glass := findExpandedPart(t, parts, "Стекла")
	if glass.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("glass release period = %q, want %q", glass.CarReleasePeriod, report.CarReleasePeriod)
	}
	if glass.Drive != report.Drive {
		t.Errorf("glass drive = %q, want %q", glass.Drive, report.Drive)
	}
	if glass.Transmission != report.Transmission {
		t.Errorf("glass transmission = %q, want %q", glass.Transmission, report.Transmission)
	}
	if glass.TransmissionModel != report.TransmissionModel {
		t.Errorf("glass transmission model = %q, want %q", glass.TransmissionModel, report.TransmissionModel)
	}
	if glass.VIN != report.VIN {
		t.Errorf("glass VIN = %q, want %q", glass.VIN, report.VIN)
	}
	if glass.CarReleaseDate != "2011" {
		t.Errorf("glass release date = %q, want %q", glass.CarReleaseDate, "2011")
	}

	// 7. Диски и шины
	wheels := findExpandedPart(t, parts, "Диски и шины")
	if wheels.CarReleasePeriod != report.CarReleasePeriod {
		t.Errorf("wheels release period = %q, want %q", wheels.CarReleasePeriod, report.CarReleasePeriod)
	}
	if wheels.Drive != report.Drive {
		t.Errorf("wheels drive = %q, want %q", wheels.Drive, report.Drive)
	}

	// Проверяем, что car_release_period применился ко ВСЕМ созданным запчастям
	for _, part := range parts {
		if part.CarReleasePeriod != report.CarReleasePeriod {
			t.Errorf("part %q (category %q) has car_release_period = %q, want %q", part.Name, part.Category, part.CarReleasePeriod, report.CarReleasePeriod)
			break
		}
		if part.Transmission != report.Transmission || part.TransmissionModel != report.TransmissionModel || part.Drive != report.Drive {
			t.Errorf("part %q did not receive transmission data: transmission=%q model=%q drive=%q", part.Name, part.Transmission, part.TransmissionModel, part.Drive)
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

	// Подвеска передних колес: все поля из ведомости
	if parts[1].CarReleasePeriod != "2005-2011" {
		t.Errorf("parts[1].CarReleasePeriod = %q, want 2005-2011", parts[1].CarReleasePeriod)
	}
	if parts[1].Transmission != "АКПП" {
		t.Errorf("parts[1].Transmission = %q, want АКПП", parts[1].Transmission)
	}
	if parts[1].TransmissionModel != "6HP19" {
		t.Errorf("parts[1].TransmissionModel = %q, want 6HP19", parts[1].TransmissionModel)
	}
	if parts[1].Drive != "Задний" {
		t.Errorf("parts[1].Drive = %q, want Задний", parts[1].Drive)
	}

	// Выхлопная система: все поля из ведомости
	if parts[2].CarReleasePeriod != "2005-2011" {
		t.Errorf("parts[2].CarReleasePeriod = %q, want 2005-2011", parts[2].CarReleasePeriod)
	}
	if parts[2].Drive != "Задний" {
		t.Errorf("parts[2].Drive = %q, want Задний", parts[2].Drive)
	}
	if parts[2].Transmission != "АКПП" {
		t.Errorf("parts[2].Transmission = %q, want АКПП", parts[2].Transmission)
	}
	if parts[2].TransmissionModel != "6HP19" {
		t.Errorf("parts[2].TransmissionModel = %q, want 6HP19", parts[2].TransmissionModel)
	}

	// Стекла: все поля из ведомости
	if parts[3].CarReleasePeriod != "2005-2011" {
		t.Errorf("parts[3].CarReleasePeriod = %q, want 2005-2011", parts[3].CarReleasePeriod)
	}
	if parts[3].Drive != "Задний" {
		t.Errorf("parts[3].Drive = %q, want Задний", parts[3].Drive)
	}
	if parts[3].Transmission != "АКПП" {
		t.Errorf("parts[3].Transmission = %q, want АКПП", parts[3].Transmission)
	}
	if parts[3].TransmissionModel != "6HP19" {
		t.Errorf("parts[3].TransmissionModel = %q, want 6HP19", parts[3].TransmissionModel)
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
