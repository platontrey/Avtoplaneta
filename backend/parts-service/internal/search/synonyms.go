package search

import (
	"strings"

	"parts-service/internal/catalog"
)

// ExpandFrontRearSynonyms разворачивает любое обозначение положения перед/зад (русское, английское, F/R)
func ExpandFrontRearSynonyms(val string) []string {
	return catalog.ExpandFrontRearSynonyms(val)
}

// ExpandLeftRightSynonyms разворачивает любое обозначение стороны право/лево (русское, английское, L/R)
func ExpandLeftRightSynonyms(val string) []string {
	return catalog.ExpandLeftRightSynonyms(val)
}

// ExpandTopBottomSynonyms разворачивает любое обозначение вертикального положения верх/низ
func ExpandTopBottomSynonyms(val string) []string {
	return catalog.ExpandTopBottomSynonyms(val)
}

// ExpandCategorySynonyms разворачивает категорию в синонимы и подкатегории.
func ExpandCategorySynonyms(val string) []string {
	return catalog.ExpandCategorySynonyms(val)
}

// GetPositionTermQueries проверяет, является ли терм маркером расположения/стороны,
// и возвращает запросы к колонкам front_rear, left_right, top_bottom.
func GetPositionTermQueries(term string) []map[string]interface{} {
	clean := strings.ToLower(strings.TrimSpace(term))
	var queries []map[string]interface{}

	// Проверка на Перед: "перед", "передний", "передняя", "переднее", "передние", "front"
	if strings.HasPrefix(clean, "перед") || clean == "front" {
		synonyms := []interface{}{"F", "f", "Front", "front", "перед", "передний", "передняя", "переднее", "перед / зад"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"front_rear": synonyms,
			},
		})
	}

	// Проверка на Зад: "зад", "задний", "задняя", "заднее", "задние", "rear"
	if strings.HasPrefix(clean, "зад") || clean == "rear" {
		synonyms := []interface{}{"R", "r", "Rear", "rear", "зад", "задний", "задняя", "заднее", "перед / зад"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"front_rear": synonyms,
			},
		})
	}

	// Проверка на Право: "прав", "правый", "правая", "правое", "правые", "право", "right"
	if strings.HasPrefix(clean, "прав") || clean == "right" {
		synonyms := []interface{}{"R", "r", "Right", "right", "право", "правый", "правая", "правое", "лево / право"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"left_right": synonyms,
			},
		})
	}

	// Проверка на Лево: "лев", "левый", "левая", "левое", "левые", "лево", "left"
	if strings.HasPrefix(clean, "лев") || clean == "left" {
		synonyms := []interface{}{"L", "l", "Left", "left", "лево", "левый", "левая", "левое", "лево / право"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"left_right": synonyms,
			},
		})
	}

	// Проверка на Верх: "верх", "верхний", "верхняя", "верхнее", "top", "upper"
	if strings.HasPrefix(clean, "верх") || clean == "top" || clean == "upper" {
		synonyms := []interface{}{"T", "t", "U", "u", "Top", "top", "Upper", "upper", "верх", "верхний", "верхняя", "верх / низ"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"top_bottom": synonyms,
			},
		})
	}

	// Проверка на Низ: "низ", "нижний", "нижняя", "нижнее", "bottom", "lower"
	if strings.HasPrefix(clean, "низ") || clean == "bottom" || clean == "lower" {
		synonyms := []interface{}{"B", "b", "L", "l", "Bottom", "bottom", "Lower", "lower", "низ", "нижний", "нижняя", "верх / низ"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"top_bottom": synonyms,
			},
		})
	}

	return queries
}
