package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Вызов OpenRouter API
func callOpenRouter(message string, context map[string]interface{}) (string, error) {
	log.Printf("callOpenRouter: Начинаем вызов OpenRouter API для сообщения: %s", message)
	systemPrompt := `Ты ИИ-помощник системы управления автозапчастями Avtoplaneta.
Твоя задача - помогать пользователям с управлением инвентарем, заказами и другими функциями системы.

У тебя есть доступ к реальной базе данных запчастей. Используй предоставленный контекст для точного поиска и операций.

ПОНИМАЙ СИНОНИМЫ И СОКРАЩЕНИЯ:
- "левый" = "L", "left", "лев"
- "правый" = "R", "right", "прав"
- "передний" = "front", "перед", "Ф"
- "задний" = "rear", "back", "зад", "З"
- "тормоз" = "тормозной", "brake"
- "амортизатор" = "стойка", "shock", "аморт"
- "фильтр" = "filter", "фильт"

АНАЛИЗИРУЙ ЗАПРОСЫ ПОЛЬЗОВАТЕЛЯ И ОПРЕДЕЛЯЙ НАМЕРЕНИЕ:
- "удали амортизатор BMW E90 левый" → найти "Амортизатор BMW E90 L" и удалить
- "добавь тормозные колодки цена 1500" → добавить запчасть
- "покажи все запчасти BMW" → поиск запчастей BMW
- "измени цену на амортизатор левый" → обновить цену
- "измени категорию на амортизатор левый на Трансмиссия" → изменить категорию

ЕСЛИ НАЙДЕНО НЕСКОЛЬКО ЗАПЧАСТЕЙ - УТОЧНИ:
- Проверь реальную базу данных в контексте
- Предложи конкретные варианты из базы
- "Я нашел несколько вариантов. Какой именно вы имеете в виду: [вариант1], [вариант2]?"

КОМАНДЫ НА РУССКОМ ЯЗЫКЕ:

УПРАВЛЕНИЕ ЗАПЧАСТЯМИ:
- "добавить [название] цена [цена] количество [кол-во]" - добавить новую запчасть
- "обновить [название] [поле] [значение]" - обновить существующую запчасть (цена, количество, название, категория)
- "удалить [название]" - удалить запчасть
- "показать [фильтр]" - показать список запчастей
- "найти [запчасть]" - поиск запчастей

УПРАВЛЕНИЕ ЗАКАЗАМИ:
- "заказы", "ордера" - показать заказы
- "создать заказ" - создать новый заказ

СТАТИСТИКА И АДМИНИСТРИРОВАНИЕ:
- "статистика", "отчеты" - показать статистику
- "админ", "панель администратора" - открыть админ-панель

ПРАВИЛА ЗАПОЛНЕНИЯ ПОЛЕЙ ЗАПЧАСТИ:

КОГДА ПОЛЬЗОВАТЕЛЬ ГОВОРИТ "добавь абсорбер BMW E5 цена 1000":
- name: "Абсорбер" (только название запчасти)
- brand: "BMW" (марка автомобиля)
- model: "E5" (модель автомобиля)
- price: 1000
- quantity: 1 (по умолчанию, если не указано)

КОГДА ПОЛЬЗОВАТЕЛЬ ГОВОРИТ "добавь тормозные колодки передние Audi A4 цена 2500 4 штуки":
- name: "Тормозные колодки передние" (название + характеристики)
- brand: "Audi"
- model: "A4"
- price: 2500
- quantity: 4

ПРИМЕРЫ ДЕЙСТВИЙ:

{"type": "navigate", "path": "/add-part"} - навигация
{"type": "search", "query": "тормозные колодки"} - поиск
{"type": "add_part", "data": {"name": "Абсорбер", "brand": "BMW", "model": "E5", "price": 1000, "quantity": 1}} - добавить запчасть
{"type": "update_part", "id": 123, "data": {"price": 1600}} - обновить запчасть
{"type": "update_part", "id": 123, "data": {"category": "Трансмиссия"}} - изменить категорию запчасти
{"type": "delete_part", "id": 123} - удалить запчасть
{"type": "search_and_delete", "part_name": "Амортизатор BMW E90 L"} - поиск и удаление
{"type": "search_and_update", "part_name": "Тормозные колодки", "field": "price", "value": "1600"} - поиск и обновление
{"type": "clarify", "options": ["Амортизатор BMW E90 L", "Амортизатор BMW E90 R"]} - уточнение

ВАЖНО:
- Используй реальные данные из контекста базы данных
- Понимай синонимы и сокращения
- Если пользователь сказал "левый", ищи "L" в названиях
- Если пользователь сказал "правый", ищи "R" в названиях
- Будь точным в поиске и операциях
- Отвечай на русском языке. Будь полезным и дружелюбным.

ФОРМАТ ОТВЕТА:
- Твой ответ ДОЛЖЕН быть в формате JSON
- Структура: {"response": "твой текстовый ответ пользователю", "action": {"type": "тип действия", ...}}
- Если действия нет, используй "action": null
- Примеры действий: {"type": "navigate", "path": "/add-part"}, {"type": "search", "query": "запрос"}
- Всегда возвращай валидный JSON, без дополнительного текста вне JSON`

	// Добавляем контекст базы данных
	dbContext := getDatabaseContext()
	userMessage := fmt.Sprintf("Сообщение пользователя: %s\n\nКонтекст: %+v\n\nКонтекст базы данных: %+v", message, context, dbContext)

	// OpenRouter API endpoint
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	log.Printf("callOpenRouter: OPENROUTER_API_KEY установлен: %t", apiKey != "")
	if apiKey == "" {
		log.Printf("OPENROUTER_API_KEY не установлен, использую fallback логику")
		response, _ := fallbackAICommand(message, context) // Игнорируем action, так как возвращаем только текст
		log.Printf("callOpenRouter: Fallback вернул: %s", response)
		return response, nil // Возвращаем nil как ошибку, поскольку fallback всегда работает
	}

	openRouterURL := "https://openrouter.ai/api/v1/chat/completions"

	reqPayload := map[string]interface{}{
		"model": "x-ai/grok-4.1-fast:free", // Grok 4.1 Fast Free модель от xAI
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userMessage,
			},
		},
		"max_tokens":  500,
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		log.Printf("callOpenRouter: Ошибка маршалинга запроса: %v", err)
		return "", fmt.Errorf("ошибка маршалинга: %v", err)
	}
	log.Printf("callOpenRouter: Отправляемый JSON: %s", string(jsonData))

	req, err := http.NewRequest("POST", openRouterURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("callOpenRouter: Ошибка создания запроса: %v", err)
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "https://avtoplaneta.local") // Для статистики OpenRouter
	req.Header.Set("X-Title", "Avtoplaneta AI Assistant")

	log.Printf("callOpenRouter: Отправка запроса к %s", openRouterURL)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("callOpenRouter: Ошибка вызова OpenRouter API: %v", err)
		return "", fmt.Errorf("ошибка сети: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	log.Printf("callOpenRouter: Получен ответ со статусом: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("callOpenRouter: OpenRouter API вернул статус %d: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API вернул статус %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("callOpenRouter: Ошибка чтения тела ответа: %v", err)
		return "", fmt.Errorf("ошибка чтения ответа: %v", err)
	}
	log.Printf("callOpenRouter: Получено тело ответа длиной %d байт: %s", len(body), string(body))

	var openRouterResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		log.Printf("callOpenRouter: Ошибка декодирования ответа OpenRouter: %v", err)
		return "", fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	if len(openRouterResp.Choices) == 0 {
		log.Printf("callOpenRouter: OpenRouter вернул пустой choices")
		return "", fmt.Errorf("пустой ответ от API")
	}

	content := openRouterResp.Choices[0].Message.Content
	log.Printf("callOpenRouter: Извлеченный content: '%s'", content)
	return content, nil
}

// Fallback функция с rule-based логикой
func fallbackAICommand(message string, context map[string]interface{}) (string, map[string]interface{}) {
	_ = context // Может пригодиться для расширения функционала
	message = strings.ToLower(strings.TrimSpace(message))

	switch {
	case strings.Contains(message, "привет") || strings.Contains(message, "здравствуй"):
		return "Привет! Я ИИ-помощник системы Avtoplaneta. Чем могу помочь?", nil

	case strings.Contains(message, "поиск") || strings.Contains(message, "найди"):
		query := extractSearchQuery(message)
		if query != "" {
			return fmt.Sprintf("Выполняю поиск '%s'...", query), map[string]interface{}{
				"type":  "search",
				"query": query,
			}
		}
		return "Что именно вы хотите найти?", nil

	case strings.Contains(message, "добавить") || strings.Contains(message, "создать"):
		if strings.Contains(message, "запчасть") || isPartAddition(message) {
			// Простой парсинг параметров из сообщения
			name := extractPartName(message)
			price := extractPrice(message)
			quantity := extractQuantity(message)

			if name != "" {
				if price > 0 && quantity > 0 {
					return fmt.Sprintf("Добавляю запчасть '%s'...", name), map[string]interface{}{
						"type": "add_part",
						"data": map[string]interface{}{
							"name":     name,
							"price":    price,
							"quantity": quantity,
							"category": "Автозапчасти",
							"status":   true,
						},
					}
				} else {
					// Если не все параметры указаны, открываем форму
					return fmt.Sprintf("Открываю форму для добавления запчасти '%s'...", name), map[string]interface{}{
						"type": "navigate",
						"path": "/add-part",
						"prefill": map[string]interface{}{
							"name": name,
						},
					}
				}
			}
			return "Открываю форму добавления запчасти...", map[string]interface{}{
				"type": "navigate",
				"path": "/add-part",
			}
		}
		return "Что именно вы хотите добавить?", nil

	case strings.Contains(message, "обновить") || strings.Contains(message, "изменить"):
		if strings.Contains(message, "запчасть") || isPartUpdate(message) {
			partName := extractPartName(message)
			field := extractField(message)
			value := extractValue(message)

			if partName != "" && field != "" && value != "" {
				// Ищем запчасть по имени
				return fmt.Sprintf("Ищу запчасть '%s' для обновления...", partName), map[string]interface{}{
					"type":      "search_and_update",
					"part_name": partName,
					"field":     field,
					"value":     value,
				}
			}
			return "Укажите название запчасти и что нужно изменить", nil
		}

	case strings.Contains(message, "удалить") || strings.Contains(message, "удали"):
		partName := extractPartName(message)
		if partName != "" {
			// Ищем запчасть по имени для удаления
			return fmt.Sprintf("Ищу запчасть '%s' для удаления...", partName), map[string]interface{}{
				"type":      "search_and_delete",
				"part_name": partName,
			}
		}
		return "Укажите название запчасти для удаления", nil

	case strings.Contains(message, "показать") || strings.Contains(message, "список"):
		if strings.Contains(message, "запчаст") || isPartsList(message) {
			filter := extractFilter(message)
			if filter != "" {
				return fmt.Sprintf("Показываю запчасти с фильтром '%s'...", filter), map[string]interface{}{
					"type":  "search",
					"query": filter,
				}
			}
			return "Показываю список запчастей...", map[string]interface{}{
				"type": "navigate",
				"path": "/inventory",
			}
		}
		return "Что именно вы хотите показать?", nil

	case strings.Contains(message, "заказ") || strings.Contains(message, "ордер"):
		return "Перехожу к управлению заказами...", map[string]interface{}{
			"type": "navigate",
			"path": "/orders",
		}

	case strings.Contains(message, "статистика") || strings.Contains(message, "отчет"):
		return "Показываю статистику системы...", map[string]interface{}{
			"type": "navigate",
			"path": "/statistics",
		}

	case strings.Contains(message, "админ") || strings.Contains(message, "панель"):
		return "Открываю панель администратора...", map[string]interface{}{
			"type": "navigate",
			"path": "/admin",
		}
	}
	// Пытаемся распознать естественные команды
	if isNaturalCommand(message) {
		return processNaturalCommand(message)
	}
	return "Извините, я не понял вашу команду. Попробуйте сказать 'удали амортизатор BMW E90' или 'добавь тормозные колодки цена 1500'.", nil
}

// Новые функции для обработки естественных команд
func extractSearchQuery(message string) string {
	// Извлекаем поисковый запрос после слов "поиск", "найди", "покажи"
	patterns := []string{
		`поиск (.+)`,
		`найди (.+)`,
		`покажи (.+)`,
	}

	for _, pattern := range patterns {
		if match := regexp.MustCompile(pattern).FindStringSubmatch(message); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func extractPartName(message string) string {
	// Извлекаем название запчасти из естественных команд
	// Удаляем глаголы и предлоги в начале, марки автомобилей и числа в конце
	words := strings.Fields(message)
	var partName []string

	// Пропускаем глаголы и предлоги в начале
	skipWords := map[string]bool{
		"удали": true, "удалить": true, "добавь": true, "добавить": true,
		"обнови": true, "обновить": true, "измени": true, "изменить": true,
		"покажи": true, "показать": true, "найди": true, "найти": true,
		"запчасть": true, "запчасти": true,
	}

	// Марки автомобилей (пропускаем их, они пойдут в отдельные поля)
	carBrands := map[string]bool{
		"bmw": true, "mercedes": true, "audi": true, "volkswagen": true, "vw": true,
		"toyota": true, "nissan": true, "honda": true, "mazda": true, "mitsubishi": true,
		"ford": true, "chevrolet": true, "opel": true, "renault": true, "peugeot": true,
		"citroen": true, "fiat": true, "alfa": true, "lancia": true, "ferrari": true,
		"lamborghini": true, "maserati": true, "bentley": true, "rolls": true, "royce": true,
		"aston": true, "martin": true, "jaguar": true, "land": true, "rover": true,
		"volvo": true, "saab": true, "skoda": true, "seat": true, "porsche": true,
		"lada": true, "vaz": true, "gaz": true, "uaz": true, "kamaz": true,
		"zil": true, "moskvich": true, "izh": true,
	}

	// Слова, которые указывают на конец названия (цена, количество и т.д.)
	endWords := map[string]bool{
		"цена": true, "ценой": true, "руб": true, "рублей": true, "рубля": true, "рубль": true,
		"количество": true, "количеством": true, "штук": true, "штуки": true, "штука": true,
		"единиц": true, "единица": true, "шт": true, "шт.": true,
	}

	// Синонимы для замены на сокращения
	synonyms := map[string]string{
		"левый": "L", "левая": "L", "лев": "L",
		"правый": "R", "правая": "R", "прав": "R",
		"передний": "перед", "передняя": "перед",
		"задний": "зад", "задняя": "зад",
		"верхний": "верх", "верхняя": "верх",
		"нижний": "низ", "нижняя": "низ",
		"тормозной": "тормоз", "тормозные": "тормоз",
		"амортизатор": "аморт", "стойка": "аморт",
		"фильтр": "фильт", "масляный": "масло", "масляного": "масло",
		"воздушный": "воздух", "воздушного": "воздух",
	}

	for _, word := range words {
		lowerWord := strings.ToLower(word)

		// Пропускаем слова в начале
		if skipWords[lowerWord] {
			continue
		}

		// Пропускаем марки автомобилей (они пойдут в отдельные поля)
		if carBrands[lowerWord] {
			continue
		}

		// Останавливаемся на словах, указывающих на конец названия
		if endWords[lowerWord] {
			break
		}

		// Пропускаем чистые числа (цены, количества, модели типа E90, A4 и т.д.)
		if _, err := strconv.Atoi(word); err == nil {
			continue
		}

		// Пропускаем комбинации букв и цифр (модели типа E90, A4, X5)
		if matched, _ := regexp.MatchString(`^[A-Za-z]+\d+$`, word); matched {
			continue
		}
		if matched, _ := regexp.MatchString(`^\d+[A-Za-z]+$`, word); matched {
			continue
		}

		// Заменяем синонимы на сокращения
		if replacement, exists := synonyms[lowerWord]; exists {
			partName = append(partName, replacement)
		} else {
			partName = append(partName, word)
		}
	}

	result := strings.Join(partName, " ")
	// Убираем лишние пробелы и возвращаем
	return strings.TrimSpace(result)
}

func extractPrice(message string) float64 {
	// Ищем цену в сообщении - более гибкие паттерны
	patterns := []string{
		`цена (\d+(?:\.\d+)?)`,
		`(\d+(?:\.\d+)?) руб`,
		`(\d+(?:\.\d+)?)р`,
		`(\d+(?:\.\d+)?) рублей`,
		`(\d+(?:\.\d+)?) рубля`,
		`(\d+(?:\.\d+)?) рубль`,
	}

	for _, pattern := range patterns {
		if match := regexp.MustCompile(pattern).FindStringSubmatch(message); len(match) > 1 {
			if price, err := strconv.ParseFloat(match[1], 64); err == nil {
				return price
			}
		}
	}

	// Ищем просто числа в сообщении (если нет явных указателей на цену)
	// Но только если есть слова "цена" или "руб"
	if strings.Contains(message, "цена") || strings.Contains(message, "руб") || strings.Contains(message, "р") {
		// Ищем все числа в сообщении
		numberRegex := regexp.MustCompile(`(\d+(?:\.\d+)?)`)
		matches := numberRegex.FindAllStringSubmatch(message, -1)

		// Возвращаем первое найденное число
		for _, match := range matches {
			if price, err := strconv.ParseFloat(match[1], 64); err == nil && price > 0 {
				return price
			}
		}
	}

	return 0
}

func extractQuantity(message string) int {
	// Ищем количество в сообщении - более гибкие паттерны
	patterns := []string{
		`количество (\d+)`,
		`(\d+) шт`,
		`(\d+)штук`,
		`(\d+) штук`,
		`(\d+) шт\.`,
		`(\d+) штуки`,
		`(\d+) единиц`,
		`(\d+) единица`,
	}

	for _, pattern := range patterns {
		if match := regexp.MustCompile(pattern).FindStringSubmatch(message); len(match) > 1 {
			if qty, err := strconv.Atoi(match[1]); err == nil {
				return qty
			}
		}
	}

	// Ищем просто числа в сообщении (если нет явных указателей на количество)
	// Но только если есть слова "штук", "шт", "количество"
	if strings.Contains(message, "штук") || strings.Contains(message, "шт") || strings.Contains(message, "количество") || strings.Contains(message, "единиц") {
		// Ищем все числа в сообщении
		numberRegex := regexp.MustCompile(`(\d+)`)
		matches := numberRegex.FindAllStringSubmatch(message, -1)

		// Возвращаем первое найденное число
		for _, match := range matches {
			if qty, err := strconv.Atoi(match[1]); err == nil && qty > 0 {
				return qty
			}
		}
	}

	return 0
}

func extractField(message string) string {
	// Извлекаем поле для обновления
	fields := map[string]string{
		"цена":       "price",
		"цену":       "price",
		"количество": "quantity",
		"название":   "name",
		"имя":        "name",
		"категория":  "category",
		"категорию":  "category",
	}

	for word, field := range fields {
		if strings.Contains(message, word) {
			return field
		}
	}
	return ""
}

func extractValue(message string) string {
	// Извлекаем значение для обновления (после "на")
	if idx := strings.Index(message, " на "); idx != -1 {
		return strings.TrimSpace(message[idx+4:])
	}
	return ""
}

func extractBrand(message string) string {
	// Извлекаем марку автомобиля из сообщения
	brands := []string{
		"BMW", "Mercedes", "Audi", "Volkswagen", "VW", "Toyota", "Nissan", "Honda",
		"Mazda", "Mitsubishi", "Ford", "Chevrolet", "Opel", "Renault", "Peugeot",
		"Citroen", "Fiat", "Alfa", "Lancia", "Ferrari", "Lamborghini", "Maserati",
		"Bentley", "Rolls-Royce", "Aston Martin", "Jaguar", "Land Rover", "Volvo",
		"Saab", "Skoda", "Seat", "Porsche", "Lada", "VAZ", "GAZ", "UAZ", "KAMAZ",
		"ZIL", "Moskvich", "IZH",
	}

	messageUpper := strings.ToUpper(message)
	for _, brand := range brands {
		if strings.Contains(messageUpper, brand) {
			return brand
		}
	}
	return ""
}

func extractModel(message string) string {
	// Извлекаем модель автомобиля из сообщения
	// Ищем комбинации типа "E90", "A4", "X5", "C-Class" и т.д.
	patterns := []string{
		`([A-Z]\d+)`,       // E90, A4, X5
		`([A-Z]\d+[A-Z]?)`, // E90, A4, X5, C180
		`([A-Z]-[A-Z]\w*)`, // C-Class, E-Class
		`(\d+\w+)`,         // 316i, 528i
	}

	for _, pattern := range patterns {
		if match := regexp.MustCompile(pattern).FindStringSubmatch(message); len(match) > 1 {
			model := match[1]
			// Проверяем, что это не часть другого слова
			if len(model) >= 2 && len(model) <= 10 {
				return model
			}
		}
	}

	return ""
}

func extractFilter(message string) string {
	// Извлекаем фильтр (марка автомобиля и т.д.)
	patterns := []string{
		`(.+) BMW`,
		`(.+) Mercedes`,
		`(.+) Audi`,
		`(.+) Volkswagen`,
	}

	for _, pattern := range patterns {
		if match := regexp.MustCompile(pattern).FindStringSubmatch(message); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func isPartAddition(message string) bool {
	return strings.Contains(message, "добав") || strings.Contains(message, "созда")
}

func isPartUpdate(message string) bool {
	return strings.Contains(message, "обнови") || strings.Contains(message, "измени") || strings.Contains(message, "поменяй")
}

func isPartsList(message string) bool {
	return strings.Contains(message, "список") || strings.Contains(message, "все") || strings.Contains(message, "запчаст")
}

func isNaturalCommand(message string) bool {
	// Проверяем, является ли команда естественной (не явной командой)
	naturalIndicators := []string{
		"удали ", "добавь ", "покажи ", "найди ", "измени ",
		"BMW", "Mercedes", "Audi", "амортизатор", "тормоз",
	}

	for _, indicator := range naturalIndicators {
		if strings.Contains(message, indicator) {
			return true
		}
	}
	return false
}

func processNaturalCommand(message string) (string, map[string]interface{}) {
	// Обработка естественных команд
	// context может пригодиться для расширения
	if strings.Contains(message, "удали") || strings.Contains(message, "удалить") {
		partName := extractPartName(message)
		if partName != "" {
			// Ищем запчасть в базе данных
			matches := searchPartsInDatabase(partName)
			if len(matches) == 1 {
				// Найдена ровно одна запчасть
				return fmt.Sprintf("Удаляю запчасть '%s'...", matches[0]["name"]), map[string]interface{}{
					"type": "delete_part",
					"id":   matches[0]["id"],
				}
			} else if len(matches) > 1 {
				// Найдено несколько запчастей, уточняем
				options := make([]string, len(matches))
				for i, match := range matches {
					options[i] = fmt.Sprintf("%v", match["name"])
				}
				return fmt.Sprintf("Найдено несколько запчастей. Какую именно удалить: %s?", strings.Join(options, ", ")), map[string]interface{}{
					"type":    "clarify",
					"options": options,
					"action":  "delete",
				}
			} else {
				return fmt.Sprintf("Запчасть '%s' не найдена в базе данных", partName), nil
			}
		}
	} else if strings.Contains(message, "добавь") || strings.Contains(message, "добавить") {
		partName := extractPartName(message)
		price := extractPrice(message)
		quantity := extractQuantity(message)
		brand := extractBrand(message)
		model := extractModel(message)

		if partName != "" {
			// Создаем данные для запчасти
			partData := map[string]interface{}{
				"name":     partName,
				"price":    price,
				"quantity": quantity,
				"category": "Автозапчасти",
				"status":   true,
			}

			// Добавляем бренд и модель, если найдены
			if brand != "" {
				partData["brand"] = brand
			}
			if model != "" {
				partData["model"] = model
			}

			// Если цена и количество указаны явно, создаем запчасть
			if price > 0 && quantity > 0 {
				return fmt.Sprintf("Добавляю запчасть '%s'...", partName), map[string]interface{}{
					"type": "add_part",
					"data": partData,
				}
			} else {
				// Если не все параметры указаны, пытаемся создать с разумными значениями по умолчанию
				// или открываем форму для уточнения
				if price > 0 {
					// Цена указана, количество по умолчанию 1
					partData["quantity"] = 1
					return fmt.Sprintf("Добавляю запчасть '%s' с ценой %.2f руб. (количество: 1)...", partName, price), map[string]interface{}{
						"type": "add_part",
						"data": partData,
					}
				} else if quantity > 0 {
					// Количество указано, цена по умолчанию 0 (нужно будет указать позже)
					partData["price"] = 0
					return fmt.Sprintf("Добавляю запчасть '%s' в количестве %d шт. (цена будет установлена позже)...", partName, quantity), map[string]interface{}{
						"type": "add_part",
						"data": partData,
					}
				} else {
					// Ничего не указано, открываем форму
					return fmt.Sprintf("Открываю форму для добавления запчасти '%s'...", partName), map[string]interface{}{
						"type":    "navigate",
						"path":    "/add-part",
						"prefill": partData,
					}
				}
			}
		}
	} else if strings.Contains(message, "покажи") || strings.Contains(message, "показать") {
		filter := extractFilter(message)
		if filter != "" {
			return fmt.Sprintf("Показываю запчасти '%s'...", filter), map[string]interface{}{
				"type":  "search",
				"query": filter,
			}
		}
	}

	return "Извините, я не понял вашу команду. Попробуйте сказать 'удали амортизатор BMW E90 левый' или 'добавь тормозные колодки цена 1500'.", nil
}

// Поиск запчастей в базе данных
func searchPartsInDatabase(query string) []map[string]interface{} {
	// query может быть расширен для использования context в будущем
	parts := getPartsFromDatabase()
	var matches []map[string]interface{}

	query = strings.ToLower(query)

	for _, part := range parts {
		name, ok := part["name"].(string)
		if !ok {
			continue
		}

		nameLower := strings.ToLower(name)

		// Проверяем точное совпадение или частичное
		if strings.Contains(nameLower, query) || strings.Contains(query, nameLower) {
			matches = append(matches, part)
		}

		// Проверяем синонимы
		if strings.Contains(query, "левый") || strings.Contains(query, "L") {
			if strings.Contains(nameLower, "l") || strings.Contains(nameLower, "лев") {
				matches = append(matches, part)
			}
		}
		if strings.Contains(query, "правый") || strings.Contains(query, "R") {
			if strings.Contains(nameLower, "r") || strings.Contains(nameLower, "прав") {
				matches = append(matches, part)
			}
		}
	}

	// Убираем дубликаты
	seen := make(map[interface{}]bool)
	var unique []map[string]interface{}
	for _, match := range matches {
		if id := match["id"]; !seen[id] {
			seen[id] = true
			unique = append(unique, match)
		}
	}

	return unique
}

// Парсинг ответа ИИ и извлечение действия
func parseAIResponse(aiResponse string) (string, map[string]interface{}) {
	log.Printf("parseAIResponse: Входной ответ ИИ: '%s'", aiResponse)

	// Логируем длину ответа и проверяем на наличие фигурных скобок
	log.Printf("parseAIResponse: Длина ответа: %d символов", len(aiResponse))
	hasOpeningBrace := strings.Contains(aiResponse, "{")
	hasClosingBrace := strings.Contains(aiResponse, "}")
	log.Printf("parseAIResponse: Содержит открывающую скобку '{': %t, закрывающую '}': %t", hasOpeningBrace, hasClosingBrace)

	// Пытаемся найти JSON в ответе
	start := strings.Index(aiResponse, "{")
	end := strings.LastIndex(aiResponse, "}")

	log.Printf("parseAIResponse: Найден JSON с start=%d, end=%d", start, end)
	if start != -1 && end != -1 && end > start {
		jsonPart := aiResponse[start : end+1]
		log.Printf("parseAIResponse: Извлеченный JSON: '%s'", jsonPart)
		log.Printf("parseAIResponse: Длина JSON части: %d символов", len(jsonPart))

		// Проверяем, является ли JSON валидным
		if !json.Valid([]byte(jsonPart)) {
			log.Printf("parseAIResponse: JSON часть не является валидным JSON")
		} else {
			log.Printf("parseAIResponse: JSON часть является валидным JSON")
		}

		// Сначала пробуем парсить как объект с response и action
		var result struct {
			Response string                 `json:"response"`
			Action   map[string]interface{} `json:"action,omitempty"`
		}

		if err := json.Unmarshal([]byte(jsonPart), &result); err == nil {
			log.Printf("parseAIResponse: Успешно распарсено как result: response='%s', action=%+v", result.Response, result.Action)
			if result.Response != "" {
				return result.Response, result.Action
			}
		} else {
			log.Printf("parseAIResponse: Ошибка парсинга как result: %v", err)
		}

		// Если не получилось, пробуем парсить как прямое действие (новый формат ИИ)
		var directAction map[string]interface{}
		if err := json.Unmarshal([]byte(jsonPart), &directAction); err == nil {
			log.Printf("parseAIResponse: Успешно распарсено как directAction: %+v", directAction)
			// Проверяем, есть ли тип действия
			if actionType, exists := directAction["type"]; exists && actionType != nil {
				// Это прямое действие, возвращаем текст до JSON как response
				responseText := strings.TrimSpace(aiResponse[:start])
				log.Printf("parseAIResponse: Возвращаем responseText='%s', directAction=%+v", responseText, directAction)
				return responseText, directAction
			} else {
				log.Printf("parseAIResponse: directAction не содержит поле 'type' или оно nil")
			}
		} else {
			log.Printf("parseAIResponse: Ошибка парсинга как directAction: %v", err)
		}
	} else {
		log.Printf("parseAIResponse: JSON не найден в ответе")
		// Дополнительные логи для анализа ответа
		log.Printf("parseAIResponse: Проверяем на ключевые слова действий...")
		actionKeywords := []string{"navigate", "search", "add_part", "update_part", "delete_part"}
		for _, keyword := range actionKeywords {
			if strings.Contains(strings.ToLower(aiResponse), keyword) {
				log.Printf("parseAIResponse: Найдено ключевое слово действия: '%s'", keyword)
			}
		}
	}

	// Если JSON не найден или не распарсился, возвращаем весь ответ как текст
	log.Printf("parseAIResponse: Возвращаем весь ответ как текст: '%s'", aiResponse)
	return aiResponse, nil
}

// Получить контекст базы данных для ИИ
func getDatabaseContext() map[string]interface{} {
	// Получаем реальные данные из базы данных через API parts-service
	partsData := getPartsFromDatabase()

	return map[string]interface{}{
		"available_operations": []string{
			"add_part", "update_part", "delete_part", "search_parts", "list_parts",
		},
		"categories": []string{
			"Двигатель", "Трансмиссия", "Ходовая часть", "Электрика",
			"Кузов", "Интерьер", "Шины и диски", "Автозапчасти",
		},
		"existing_parts": partsData,
		"search_synonyms": map[string][]string{
			"левый":       {"L", "left", "лев", "Л"},
			"правый":      {"R", "right", "прав", "П"},
			"передний":    {"front", "перед", "Ф"},
			"задний":      {"rear", "back", "зад", "З"},
			"верхний":     {"upper", "up", "верх"},
			"нижний":      {"lower", "down", "низ"},
			"тормоз":      {"тормозной", "brake", "торм"},
			"амортизатор": {"стойка", "shock", "аморт"},
			"фильтр":      {"filter", "фильт"},
			"масло":       {"oil", "масл"},
			"воздух":      {"air", "возд"},
		},
		"note": "ИИ имеет доступ к полному списку запчастей и может выполнять CRUD операции",
	}
}

// Получить данные о запчастях из базы данных
func getPartsFromDatabase() []map[string]interface{} {
	// Делаем HTTP запрос к parts-service для получения списка запчастей
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", "http://localhost:8081/api/inventory", nil)
	if err != nil {
		log.Printf("Ошибка создания запроса к БД: %v", err)
		return []map[string]interface{}{}
	}

	// Добавляем необходимые заголовки для аутентификации
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка запроса к parts-service: %v", err)
		return []map[string]interface{}{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		log.Printf("Ошибка ответа от parts-service: статус %d", resp.StatusCode)
		return []map[string]interface{}{}
	}

	// Читаем тело ответа
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return []map[string]interface{}{}
	}

	var response struct {
		Parts []map[string]interface{} `json:"parts"`
	}

	// Сначала пробуем декодировать как объект с полем parts
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		// Если не получилось, пробуем декодировать как массив напрямую
		var partsArray []map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &partsArray); err != nil {
			log.Printf("Ошибка декодирования ответа: %v", err)
			return []map[string]interface{}{}
		}
		response.Parts = partsArray
	}

	// Возвращаем только основные поля для контекста ИИ
	var partsContext []map[string]interface{}
	for _, part := range response.Parts {
		partContext := map[string]interface{}{
			"id":       part["id"],
			"name":     part["name"],
			"category": part["category"],
			"brand":    part["brand"],
			"model":    part["model"],
			"quantity": part["quantity"],
			"price":    part["price"],
		}
		partsContext = append(partsContext, partContext)
	}

	return partsContext
}

// Основная функция обработки команд ИИ агента
func processAICommand(message string, context map[string]interface{}) (string, map[string]interface{}) {
	log.Printf("processAICommand: Начинаем обработку сообщения: '%s'", message)
	// Сначала пытаемся вызвать OpenRouter
	aiResponse, err := callOpenRouter(message, context)
	if err != nil {
		log.Printf("processAICommand: Ошибка ИИ: %v, использую fallback", err)
		// Используем fallback логику
		fallbackResponse, fallbackAction := fallbackAICommand(message, context)
		log.Printf("processAICommand: Fallback ответил: response='%s', action=%+v", fallbackResponse, fallbackAction)
		return fallbackResponse, fallbackAction
	}

	// Логируем ответ ИИ для отладки
	log.Printf("processAICommand: ИИ ответил: '%s'", aiResponse)

	// Парсим ответ ИИ
	response, action := parseAIResponse(aiResponse)
	log.Printf("processAICommand: После парсинга: response='%s', action=%+v", response, action)
	return response, action
}
