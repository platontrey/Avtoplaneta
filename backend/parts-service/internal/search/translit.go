package search

import (
	"strings"
	"unicode"
)

var (
	latinToCyrillicReplacer = strings.NewReplacer(
		"shch", "щ", "sh", "ш", "ch", "ч", "zh", "ж",
		"ya", "я", "yu", "ю", "yo", "ё", "ts", "ц",
		"a", "а", "b", "б", "v", "в", "g", "г", "d", "д",
		"e", "е", "z", "з", "i", "и", "j", "й", "k", "к",
		"l", "л", "m", "м", "n", "н", "o", "о", "p", "п",
		"r", "р", "s", "с", "t", "т", "u", "у", "f", "ф",
		"h", "х", "c", "к", "y", "ы", "w", "в", "x", "кс",
	)

	qwertyToRussianTable [128]rune
)

func init() {
	qwertyMap := map[rune]rune{
		'q': 'й', 'w': 'ц', 'e': 'у', 'r': 'к', 't': 'е', 'y': 'н', 'u': 'г', 'i': 'ш', 'o': 'щ', 'p': 'з', '[': 'х', ']': 'ъ',
		'a': 'ф', 's': 'ы', 'd': 'в', 'f': 'а', 'g': 'п', 'h': 'р', 'j': 'о', 'k': 'л', 'l': 'д', ';': 'ж', '\'': 'э',
		'z': 'я', 'x': 'ч', 'c': 'с', 'v': 'м', 'b': 'и', 'n': 'т', 'm': 'ь', ',': 'б', '.': 'ю',
	}
	for k, v := range qwertyMap {
		if k < 128 {
			qwertyToRussianTable[k] = v
		}
	}
}

// TransliterateLatinToCyrillic преобразует транслит латиницы в кириллицу (sirena -> сирена, bamper -> бампер)
func TransliterateLatinToCyrillic(text string) string {
	return latinToCyrillicReplacer.Replace(strings.ToLower(text))
}

// ConvertQwertyToRussian переводит текст с неверной QWERTY раскладки на кириллицу (gthtlybq -> передний)
func ConvertQwertyToRussian(text string) string {
	var builder strings.Builder
	builder.Grow(len(text) * 2)
	for _, char := range text {
		lowerChar := unicode.ToLower(char)
		if lowerChar < 128 {
			if ruChar := qwertyToRussianTable[lowerChar]; ruChar != 0 {
				builder.WriteRune(ruChar)
				continue
			}
		}
		builder.WriteRune(lowerChar)
	}
	return builder.String()
}

// CleanPartNumber очищает артикул / код детали от пробелов, дефисов и спецсимволов, переводя в нижний регистр.
func CleanPartNumber(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r >= 'а' && r <= 'я') || r == 'ё' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// escapeESQuery экранирует спецсимволы синтаксиса Lucene для безопасного использования в query_string
func escapeESQuery(s string) string {
	needEscape := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '+', '-', '=', '!', '(', ')', '{', '}', '[', ']', '^', '"', '~', '*', '?', ':', '\\', '/', '&', '|', '<', '>':
			needEscape = true
			break
		}
	}
	if !needEscape {
		return s
	}

	var sb strings.Builder
	sb.Grow(len(s) + 4)
	for _, r := range s {
		switch r {
		case '+', '-', '=', '!', '(', ')', '{', '}', '[', ']', '^', '"', '~', '*', '?', ':', '\\', '/', '&', '|', '<', '>':
			sb.WriteRune('\\')
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
