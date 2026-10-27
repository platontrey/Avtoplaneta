package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"parts-service/internal/domain"
)

func TestSplitMultiSearchQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   \t\n  ",
			expected: nil,
		},
		{
			name:     "single word",
			input:    "бампер",
			expected: []string{"бампер"},
		},
		{
			name:     "single phrase with multiple words",
			input:    "бампер передний toyota camry",
			expected: []string{"бампер передний toyota camry"},
		},
		{
			name:     "comma separated",
			input:    "бампер, фара, капот",
			expected: []string{"бампер", "фара", "капот"},
		},
		{
			name:     "semicolon separated",
			input:    "капот camry; крыло camry",
			expected: []string{"капот camry", "крыло camry"},
		},
		{
			name:     "newline separated",
			input:    "бампер\nфара\r\nкапот",
			expected: []string{"бампер", "фара", "капот"},
		},
		{
			name:     "pipe separated",
			input:    "дверь левая | дверь правая",
			expected: []string{"дверь левая", "дверь правая"},
		},
		{
			name:     "russian keyword ИЛИ uppercase and lowercase",
			input:    "фара левая ИЛИ фара правая или решетка",
			expected: []string{"фара левая", "фара правая", "решетка"},
		},
		{
			name:     "english keyword OR",
			input:    "bumper OR headlight",
			expected: []string{"bumper", "headlight"},
		},
		{
			name:     "embedded substrings like мотор or форсунка are not split",
			input:    "мотор форсунка",
			expected: []string{"мотор форсунка"},
		},
		{
			name:     "OEM codes list",
			input:    "81150-33630, 81110-33630; 53101-33230",
			expected: []string{"81150-33630", "81110-33630", "53101-33230"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SplitMultiSearchQuery(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestBuildElasticsearchQuery_SingleSearch(t *testing.T) {
	params := domain.InventoryQueryParams{
		Search: "бампер передний",
	}

	query := BuildElasticsearchQuery(params)
	require.NotNil(t, query)

	boolQ, ok := query["bool"].(map[string]interface{})
	require.True(t, ok)

	mustList, ok := boolQ["must"].([]map[string]interface{})
	require.True(t, ok)
	// Для одиночного поиска должно быть по 1 must-условию на каждое слово ("бампер", "передний")
	assert.Len(t, mustList, 2)

	shouldList, ok := boolQ["should"].([]map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, shouldList)
}

func TestBuildElasticsearchQuery_MultiSearch(t *testing.T) {
	params := domain.InventoryQueryParams{
		Search: "бампер передний, фара левая camry",
	}

	query := BuildElasticsearchQuery(params)
	require.NotNil(t, query)

	boolQ, ok := query["bool"].(map[string]interface{})
	require.True(t, ok)

	mustList, ok := boolQ["must"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, mustList, 1)

	// Внутри must[0] должна быть bool-обёртка со should на 2 фрагмента
	multiBool, ok := mustList[0]["bool"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, multiBool["minimum_should_match"])

	multiShould, ok := multiBool["should"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, multiShould, 2)

	// Фрагмент 1: "бампер передний"
	frag1Bool := multiShould[0]["bool"].(map[string]interface{})
	frag1Must := frag1Bool["must"].([]map[string]interface{})
	assert.Len(t, frag1Must, 2) // "бампер", "передний"

	// Фрагмент 2: "фара левая camry"
	frag2Bool := multiShould[1]["bool"].(map[string]interface{})
	frag2Must := frag2Bool["must"].([]map[string]interface{})
	assert.Len(t, frag2Must, 3) // "фара", "левая", "camry"
}

func TestCleanPartNumber(t *testing.T) {
	assert.Equal(t, "8628060321", CleanPartNumber("86280-60321"))
	assert.Equal(t, "8628060321", CleanPartNumber("86280 60321"))
	assert.Equal(t, "8628060321", CleanPartNumber("  86280-60321  "))
	assert.Equal(t, "8628060321", CleanPartNumber("8628060321"))
	assert.Equal(t, "8628015000", CleanPartNumber("86280-15000"))
	assert.NotEqual(t, CleanPartNumber("86280-60321"), CleanPartNumber("86280-15000"))
	assert.Equal(t, "a2229003300", CleanPartNumber("A-222-900-33-00"))
}

func TestBuildElasticsearchQuery_ExactCodes(t *testing.T) {
	params := domain.InventoryQueryParams{
		OEMCode:          "86280-60321",
		ManufacturerCode: "DENSO-001",
	}

	query := BuildElasticsearchQuery(params)
	require.NotNil(t, query)

	boolQ, ok := query["bool"].(map[string]interface{})
	require.True(t, ok)

	filterList, ok := boolQ["filter"].([]map[string]interface{})
	require.True(t, ok)
	require.NotEmpty(t, filterList)

	// Проверяем наличие фильтров по oem_code и manufacturer_code
	foundOEM := false
	foundMfr := false

	for _, f := range filterList {
		subBool, ok := f["bool"].(map[string]interface{})
		if !ok {
			continue
		}
		shouldList, ok := subBool["should"].([]map[string]interface{})
		if !ok {
			continue
		}
		for _, s := range shouldList {
			if termQ, ok := s["term"].(map[string]interface{}); ok {
				if _, ok := termQ["oem_code"]; ok {
					foundOEM = true
				}
				if _, ok := termQ["manufacturer_code"]; ok {
					foundMfr = true
				}
			}
			if matchQ, ok := s["match"].(map[string]interface{}); ok {
				if m, ok := matchQ["oem_code.text"].(map[string]interface{}); ok {
					assert.Equal(t, "AND", m["operator"])
				}
				if m, ok := matchQ["manufacturer_code.text"].(map[string]interface{}); ok {
					assert.Equal(t, "AND", m["operator"])
				}
			}
		}
	}

	assert.True(t, foundOEM, "Фильтр по OEM коду должен присутствовать")
	assert.True(t, foundMfr, "Фильтр по коду производителя должен присутствовать")
}

