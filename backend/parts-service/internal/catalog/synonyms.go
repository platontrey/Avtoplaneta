package catalog

import (
	"strings"
)

// ExpandFrontRearSynonyms разворачивает любое обозначение положения перед/зад (русское, английское, F/R)
// во все возможные синонимы для поиска в БД и Elasticsearch.
func ExpandFrontRearSynonyms(val string) []string {
	clean := strings.ToLower(strings.TrimSpace(val))
	switch {
	case clean == "f" || clean == "front" || strings.HasPrefix(clean, "перед"):
		return []string{"F", "f", "Front", "front", "перед", "передний", "передняя", "переднее", "передние", "перед / зад", "перед/зад", "F/R", "F / R"}
	case clean == "r" || clean == "rear" || strings.HasPrefix(clean, "зад"):
		return []string{"R", "r", "Rear", "rear", "зад", "задний", "задняя", "заднее", "задние", "перед / зад", "перед/зад", "F/R", "F / R"}
	case clean == "перед / зад" || clean == "перед/зад" || clean == "f/r" || clean == "f / r":
		return []string{"перед / зад", "перед/зад", "F/R", "F / R", "F", "R"}
	default:
		return []string{val}
	}
}

// ExpandLeftRightSynonyms разворачивает любое обозначение стороны право/лево (русское, английское, L/R)
// во все возможные синонимы для поиска в БД и Elasticsearch.
func ExpandLeftRightSynonyms(val string) []string {
	clean := strings.ToLower(strings.TrimSpace(val))
	switch {
	case clean == "r" || clean == "right" || strings.HasPrefix(clean, "прав"):
		return []string{"R", "r", "Right", "right", "право", "правый", "правая", "правое", "правые", "прав", "лево / право", "лево/право", "L/R", "L / R"}
	case clean == "l" || clean == "left" || strings.HasPrefix(clean, "лев"):
		return []string{"L", "l", "Left", "left", "лево", "левый", "левая", "левое", "левые", "лев", "лево / право", "лево/право", "L/R", "L / R"}
	case clean == "лево / право" || clean == "лево/право" || clean == "l/r" || clean == "l / r":
		return []string{"лево / право", "лево/право", "L/R", "L / R", "L", "R"}
	default:
		return []string{val}
	}
}

// ExpandTopBottomSynonyms разворачивает любое обозначение вертикального положения верх/низ
// во все возможные синонимы для поиска в БД и Elasticsearch.
func ExpandTopBottomSynonyms(val string) []string {
	clean := strings.ToLower(strings.TrimSpace(val))
	switch {
	case clean == "t" || clean == "u" || clean == "top" || clean == "upper" || strings.HasPrefix(clean, "верх"):
		return []string{"T", "t", "U", "u", "Top", "top", "Upper", "upper", "верх", "верхний", "верхняя", "верхнее", "верхние", "верх / низ", "верх/низ"}
	case clean == "b" || clean == "bottom" || clean == "lower" || strings.HasPrefix(clean, "низ"):
		return []string{"B", "b", "L", "l", "Bottom", "bottom", "Lower", "lower", "низ", "нижний", "нижняя", "нижнее", "нижние", "верх / низ", "верх/низ"}
	case clean == "верх / низ" || clean == "верх/низ" || clean == "t/b" || clean == "u/l":
		return []string{"верх / низ", "верх/низ", "T/B", "U/L", "T", "B"}
	default:
		return []string{val}
	}
}

// ExpandCategorySynonyms разворачивает категорию в синонимы и подкатегории.
func ExpandCategorySynonyms(val string) []string {
	clean := strings.TrimSpace(val)
	if clean == "" {
		return nil
	}

	variantsMap := make(map[string]bool)
	variantsMap[clean] = true
	lower := strings.ToLower(clean)

	switch {
	case strings.Contains(lower, "кузов"):
		variantsMap["Кузов"] = true
		variantsMap["Кузов внутри"] = true
		variantsMap["Кузов снаружи"] = true
	case strings.Contains(lower, "подвеск"):
		variantsMap["Подвеска"] = true
		variantsMap["Подвеска передних колес"] = true
		variantsMap["Подвеска задних колес"] = true
		variantsMap["Подвеска ДВС/КПП"] = true
	case strings.Contains(lower, "тормоз"):
		variantsMap["Тормоза"] = true
		variantsMap["Тормозная система"] = true
	case strings.Contains(lower, "электр"):
		variantsMap["Электрика"] = true
		variantsMap["Электрооснащение"] = true
	case strings.Contains(lower, "шин") || strings.Contains(lower, "диск"):
		variantsMap["Шины и диски"] = true
		variantsMap["Диски и шины"] = true
	case strings.Contains(lower, "выхлоп") || strings.Contains(lower, "глушител"):
		variantsMap["Выхлопная система"] = true
		variantsMap["Система выхлопа (Глушитель)"] = true
		variantsMap["Система выхлопа"] = true
	case strings.Contains(lower, "рулев"):
		variantsMap["Рулевое управление"] = true
		variantsMap["Система рулевого управления"] = true
	case strings.Contains(lower, "фильтр"):
		variantsMap["Система фильтрации (Фильтры)"] = true
		variantsMap["Система фильтрации"] = true
		variantsMap["Фильтры"] = true
	case strings.Contains(lower, "двигател") || strings.Contains(lower, "двс") || strings.Contains(lower, "мотор"):
		variantsMap["Двигатель"] = true
	case strings.Contains(lower, "трансмисс") || strings.Contains(lower, "кпп"):
		variantsMap["Трансмиссия"] = true
	case strings.Contains(lower, "стекл") || strings.Contains(lower, "стекло"):
		variantsMap["Стекла"] = true
	case strings.Contains(lower, "оптик") || strings.Contains(lower, "фар"):
		variantsMap["Оптика"] = true
	case strings.Contains(lower, "пневмо"):
		variantsMap["Пневмосистема"] = true
	case strings.Contains(lower, "кондицион"):
		variantsMap["Система кондиционирования"] = true
	case strings.Contains(lower, "охлажден") || strings.Contains(lower, "отоплен"):
		variantsMap["Система охлаждения и отопления"] = true
	case strings.Contains(lower, "сопутствующ"):
		variantsMap["Сопутствующие товары"] = true
	}

	variants := make([]string, 0, len(variantsMap))
	for v := range variantsMap {
		variants = append(variants, v)
	}
	return variants
}
