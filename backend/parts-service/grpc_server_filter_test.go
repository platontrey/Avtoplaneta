package main

import (
	"context"
	"net/url"
	"testing"

	partsv1 "avtoplaneta/gen/parts/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"github.com/stretchr/testify/require"
)

type inventoryParamsCaptureService struct {
	*recordingInventoryService
	params InventoryQueryParams
}

type recordingDefectReportPublisher struct {
	report DefectReportRequest
}

func (publisher *recordingDefectReportPublisher) PublishDefectReport(_ context.Context, report DefectReportRequest) error {
	publisher.report = report
	return nil
}

func (service *inventoryParamsCaptureService) GetInventory(_ context.Context, params InventoryQueryParams) ([]Part, error) {
	service.params = params
	return []Part{{
		PartCore:           PartCore{ID: 1, Name: "Test"},
		PartSpecifications: PartSpecifications{CarReleasePeriod: "2001-2007"},
	}}, nil
}

func TestPartsGRPCServerGetInventoryForwardsEveryFilter(t *testing.T) {
	service := &inventoryParamsCaptureService{recordingInventoryService: &recordingInventoryService{}}
	server := NewPartsGRPCServer(service, nil)

	response, err := server.GetInventory(context.Background(), &partsv1.GetInventoryRequest{
		Search:             "search",
		Category:           "category",
		Brand:              "brand",
		Model:              "model",
		Location:           "location",
		Address:            "address",
		Salesman:           "salesman",
		Status:             "true",
		HasPhoto:           "with",
		Number:             "number",
		OemCode:            "oem",
		Vin:                "vin",
		BodyBrand:          "body",
		EngineBrand:        "engine",
		CarReleaseDate:     "2004",
		CarReleasePeriod:   "2001-2007",
		Transmission:       "АКПП",
		Drive:              "Полный",
		Condition:          "Б/у",
		Manufacturer:       "manufacturer",
		Defect:             "defect",
		Color:              "color",
		MinPrice:           "100",
		MaxPrice:           "200",
		MinQuantity:        "1",
		MaxQuantity:        "2",
		FrontRear:          "Передний",
		LeftRight:          "Левый",
		TopBottom:          "Верхний",
		ManufacturerCode:   "manufacturer-code",
		SupplierCode:       "supplier-code",
		TransmissionModel:  "U660E",
		WearPercentage:     "10",
		Season:             "Зима",
		Diameter:           "17",
		Width:              "225",
		Profile:            "55",
		TireQuantity:       "4",
		Drilling:           "5x114.3",
		Offset:             "45",
		CenterHoleDiameter: "60.1",
		TireModel:          "model-name",
		Page:               2,
		Limit:              100,
	})
	require.NoError(t, err)

	require.Equal(t, InventoryQueryParams{
		Search:             "search",
		Category:           "category",
		Brand:              "brand",
		Model:              "model",
		Location:           "location",
		Address:            "address",
		Salesman:           "salesman",
		Status:             "true",
		HasPhoto:           "with",
		Number:             "number",
		OEMCode:            "oem",
		VIN:                "vin",
		BodyBrand:          "body",
		EngineBrand:        "engine",
		CarReleaseDate:     "2004",
		CarReleasePeriod:   "2001-2007",
		Transmission:       "АКПП",
		Drive:              "Полный",
		Condition:          "Б/у",
		Manufacturer:       "manufacturer",
		Defect:             "defect",
		Color:              "color",
		MinPrice:           "100",
		MaxPrice:           "200",
		MinQuantity:        "1",
		MaxQuantity:        "2",
		FrontRear:          "Передний",
		LeftRight:          "Левый",
		TopBottom:          "Верхний",
		ManufacturerCode:   "manufacturer-code",
		SupplierCode:       "supplier-code",
		TransmissionModel:  "U660E",
		WearPercentage:     "10",
		Season:             "Зима",
		Diameter:           "17",
		Width:              "225",
		Profile:            "55",
		TireQuantity:       "4",
		Drilling:           "5x114.3",
		Offset:             "45",
		CenterHoleDiameter: "60.1",
		TireModel:          "model-name",
		Page:               2,
		Limit:              100,
	}, service.params)
	require.Len(t, response.Parts, 1)
	require.Equal(t, "2001-2007", response.Parts[0].CarReleasePeriod)
}

func TestInventoryGatewayContractAcceptsAdvancedQueryParameters(t *testing.T) {
	request := &partsv1.GetInventoryRequest{}
	err := runtime.PopulateQueryParameters(request, url.Values{
		"car_release_period":   {"2001-2007"},
		"transmission":         {"АКПП"},
		"transmission_model":   {"U660E"},
		"manufacturer_code":    {"M-1"},
		"center_hole_diameter": {"60.1"},
		"hasPhoto":             {"with"},
	}, &utilities.DoubleArray{})
	require.NoError(t, err)
	require.Equal(t, "2001-2007", request.CarReleasePeriod)
	require.Equal(t, "АКПП", request.Transmission)
	require.Equal(t, "U660E", request.TransmissionModel)
	require.Equal(t, "M-1", request.ManufacturerCode)
	require.Equal(t, "60.1", request.CenterHoleDiameter)
	require.Equal(t, "with", request.HasPhoto)
}

func TestPartsGRPCServerAddPartKeepsCarReleasePeriod(t *testing.T) {
	service := &recordingInventoryService{}
	server := NewPartsGRPCServer(service, nil)

	created, err := server.AddPart(context.Background(), &partsv1.AddPartRequest{
		Name:             "Test part",
		CarReleasePeriod: "2001-2007",
	})
	require.NoError(t, err)
	require.Equal(t, "2001-2007", created.CarReleasePeriod)
	require.Equal(t, "2001-2007", service.snapshot()[0].CarReleasePeriod)
}

func TestPartsGRPCServerQueuesPreparedDefectReport(t *testing.T) {
	service := &recordingInventoryService{}
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)
	publisher := &recordingDefectReportPublisher{}
	server := NewPartsGRPCServer(service, NewDefectReportWorkflow(catalog, publisher))

	response, err := server.CreateDefectReport(context.Background(), &partsv1.CreateDefectReportRequest{
		Brand:             "Toyota",
		Model:             "Camry",
		Year:              "2015",
		Vin:               "TESTVIN",
		CarReleasePeriod:  "2011-2017",
		BodyBrand:         "XV50",
		EngineBrand:       "2AR-FE",
		Transmission:      "АКПП",
		TransmissionModel: "U660E",
		Drive:             "Передний",
		SelectedParts: []*partsv1.DefectReportPart{{
			Name:     "Бампер",
			Category: "Кузов снаружи",
		}},
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, response.PartsQueued)
	require.NotEmpty(t, response.Message)
	require.Empty(t, service.snapshot(), "gRPC adapter must not bypass the asynchronous workflow")
	require.Equal(t, defectReportEventVersion, publisher.report.EventVersion)
	require.Len(t, publisher.report.SelectedParts, 1)

	queued := publisher.report.SelectedParts[0]
	require.Equal(t, "TESTVIN", queued.VIN)
	require.Equal(t, "2015", queued.CarReleaseDate)
	require.Equal(t, "2011-2017", queued.CarReleasePeriod)
	require.Equal(t, "XV50", queued.BodyBrand)
	require.Equal(t, "2AR-FE", queued.EngineBrand)
	require.Equal(t, "АКПП", queued.Transmission)
	require.Equal(t, "U660E", queued.TransmissionModel)
	require.Equal(t, "Передний", queued.Drive)
}
