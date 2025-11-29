package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Структуры для тестирования Google Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content GeminiContent `json:"content"`
	} `json:"candidates"`
}

func main() {
	fmt.Println("🧪 Тестирование Google Gemini API интеграции...")

	// Проверяем API ключ
	apiKey := "AIzaSyDg3kcxf-V7pxmhBCQ-zjXm6OAj66LOuXQ"
	if apiKey == "" {
		fmt.Println("❌ GEMINI_API_KEY не установлен")
		fmt.Println("💡 Установите переменную окружения: set GEMINI_API_KEY=your_api_key")
		fmt.Println("🔑 Получить ключ: https://makersuite.google.com/app/apikey")
		return
	}

	// Сначала проверим доступные модели
	fmt.Println("📋 Получение списка доступных моделей...")
	modelsURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)

	modelsReq, err := http.NewRequest("GET", modelsURL, nil)
	if err != nil {
		fmt.Printf("❌ Ошибка создания запроса моделей: %v\n", err)
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	modelsResp, err := client.Do(modelsReq)
	if err != nil {
		fmt.Printf("❌ Ошибка получения списка моделей: %v\n", err)
		return
	}
	defer modelsResp.Body.Close()

	if modelsResp.StatusCode != 200 {
		fmt.Printf("❌ Ошибка получения моделей (статус %d)\n", modelsResp.StatusCode)
		body, _ := io.ReadAll(modelsResp.Body)
		fmt.Printf("Ответ: %s\n", string(body))
		return
	}

	var modelsData struct {
		Models []struct {
			Name             string   `json:"name"`
			Description      string   `json:"description"`
			SupportedMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}

	if err := json.NewDecoder(modelsResp.Body).Decode(&modelsData); err != nil {
		fmt.Printf("❌ Ошибка декодирования списка моделей: %v\n", err)
		return
	}

	fmt.Println("✅ Доступные модели Gemini:")
	for _, model := range modelsData.Models {
		if strings.Contains(model.Name, "gemini") {
			fmt.Printf("  📌 %s - %s\n", model.Name, model.Description)
			fmt.Printf("     Методы: %v\n", model.SupportedMethods)
		}
	}
	fmt.Println()

	// Найдем рабочую модель для generateContent
	var workingModel string
	for _, model := range modelsData.Models {
		if strings.Contains(model.Name, "gemini") {
			for _, method := range model.SupportedMethods {
				if method == "generateContent" {
					workingModel = strings.TrimPrefix(model.Name, "models/")
					break
				}
			}
			if workingModel != "" {
				break
			}
		}
	}

	if workingModel == "" {
		fmt.Println("❌ Не найдено подходящих моделей для generateContent")
		return
	}

	fmt.Printf("🎯 Используем модель: %s\n\n", workingModel)

	// Тестовое сообщение
	req := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: "Привет! Расскажи кратко о системе управления автозапчастями Avtoplaneta"},
				},
			},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("❌ Ошибка маршалинга: %v\n", err)
		return
	}

	// URL для Gemini API с найденной моделью
	geminiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", workingModel, apiKey)

	fmt.Printf("📡 Отправка запроса на Google Gemini API...\n")

	httpReq, err := http.NewRequest("POST", geminiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Ошибка создания запроса: %v\n", err)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("❌ Ошибка подключения к Gemini API: %v\n", err)
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

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		fmt.Printf("❌ Ошибка декодирования ответа: %v\n", err)
		return
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		fmt.Println("❌ Gemini вернул пустой ответ")
		return
	}

	fmt.Println("✅ Успешное подключение к Google Gemini API!")
	fmt.Printf("🤖 Ответ ИИ: %s\n", geminiResp.Candidates[0].Content.Parts[0].Text)
	fmt.Println("\n🎉 Интеграция Google Gemini работает корректно!")
}
