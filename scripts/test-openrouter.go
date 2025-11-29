package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Структуры для тестирования OpenRouter API
type OpenRouterRequest struct {
	Model       string              `json:"model"`
	Messages    []OpenRouterMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
}

type OpenRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message OpenRouterMessage `json:"message"`
	} `json:"choices"`
}

func main() {
	fmt.Println("🧪 Тестирование OpenRouter API интеграции...")

	// Проверяем API ключ
	apiKey := "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" // Замените на ваш реальный ключ
	if apiKey == "" {
		fmt.Println("❌ OPENROUTER_API_KEY не установлен")
		fmt.Println("💡 Установите переменную окружения: set OPENROUTER_API_KEY=your_api_key")
		fmt.Println("🔑 Получить ключ: https://openrouter.ai/keys")
		return
	}

	// Тестовое сообщение
	req := OpenRouterRequest{
		Model: "x-ai/grok-4.1-fast:free",
		Messages: []OpenRouterMessage{
			{
				Role:    "system",
				Content: "Ты ИИ-помощник системы управления автозапчастями Avtoplaneta. Отвечай кратко и по делу.",
			},
			{
				Role:    "user",
				Content: "Привет! Расскажи кратко о системе управления автозапчастями Avtoplaneta",
			},
		},
		MaxTokens:   500,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("❌ Ошибка маршалинга: %v\n", err)
		return
	}

	// OpenRouter API endpoint
	openRouterURL := "https://openrouter.ai/api/v1/chat/completions"

	fmt.Printf("📡 Отправка запроса на OpenRouter API...\n")

	httpReq, err := http.NewRequest("POST", openRouterURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Ошибка создания запроса: %v\n", err)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://avtoplaneta.local")
	httpReq.Header.Set("X-Title", "Avtoplaneta AI Assistant")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("❌ Ошибка подключения к OpenRouter API: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📊 Статус ответа: %d\n", resp.StatusCode)

	if resp.StatusCode != 200 {
		fmt.Printf("❌ Ошибка API: статус %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Ответ: %s\n", string(body))
		return
	}

	var openRouterResp OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		fmt.Printf("❌ Ошибка декодирования ответа: %v\n", err)
		return
	}

	if len(openRouterResp.Choices) == 0 {
		fmt.Println("❌ OpenRouter вернул пустой ответ")
		return
	}

	fmt.Println("✅ Успешное подключение к OpenRouter API!")
	fmt.Printf("🤖 Ответ ИИ: %s\n", openRouterResp.Choices[0].Message.Content)
	fmt.Println("\n🎉 Интеграция OpenRouter работает корректно!")

	// Показать стоимость (если доступна)
	fmt.Println("\n💰 Информация о стоимости:")
	fmt.Println("- Модель: x-ai/grok-4.1-fast:free (Grok 4.1 Fast Free)")
	fmt.Println("- Стоимость: Бесплатно (free tier)")
	fmt.Println("- Ограничения: Rate limits могут применяться")
}
