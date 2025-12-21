package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// --- КОНФИГУРАЦИЯ ---
const (
	SessionFile = "drom_session.json"
	// Получаем список (JSON с briefs)
	ListURL = "https://my.drom.ru/personal/messaging/inbox-list?ajax=1&fromIndex=0&count=50&list=personal"
	// Просмотр конкретного диалога
	ViewURL   = "https://my.drom.ru/personal/messaging/view?dialogId=%s&json=true&flat-layout=false&ajax=1"
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// --- СТРУКТУРЫ ДЛЯ СПИСКА ДИАЛОГОВ (твоя находка) ---
type InboxBrief struct {
	DialogID     int    `json:"dialogId"`     // Тот самый ID
	Interlocutor string `json:"interlocutor"` // Имя собеседника
}

type InboxListResponse struct {
	Briefs []InboxBrief `json:"briefs"`
}

// --- СТРУКТУРЫ ДЛЯ ПРОСМОТРА ДИАЛОГА ---
type DromApiResponse struct {
	Interlocutor string `json:"interlocutor"`
	DialogHTML   string `json:"dialog"`
}

// --- СТРУКТУРЫ ДЛЯ СОХРАНЕНИЯ ---
type CleanMessage struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	Direction string `json:"direction"`
	Time      string `json:"time"`
	Text      string `json:"text"`
	IsRead    bool   `json:"is_read"`
}

type DialogDump struct {
	DialogID     string         `json:"dialog_id"`
	Interlocutor string         `json:"interlocutor"`
	SavedAt      string         `json:"saved_at"`
	Messages     []CleanMessage `json:"messages"`
}

type SessionCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type SessionData struct {
	Cookies []SessionCookie `json:"cookies"`
}

func main() {
	if _, err := os.Stat(SessionFile); os.IsNotExist(err) {
		fmt.Printf("❌ Ошибка: Файл %s не найден! Запустите Python-скрипт для входа.\n", SessionFile)
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// 1. ПОЛУЧЕНИЕ СПИСКА
	fmt.Println("🔍 Запрашиваю список диалогов...")
	briefs, err := fetchBriefs(client)
	if err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
		return
	}
	fmt.Printf("✅ Найдено диалогов: %d\n", len(briefs))

	// 2. ПАРСИНГ КАЖДОГО ДИАЛОГА
	for i, brief := range briefs {
		dID := strconv.Itoa(brief.DialogID) // Конвертируем int в string

		fmt.Printf("[%d/%d] Скачиваю переписку с %s (ID: %s)... ", i+1, len(briefs), brief.Interlocutor, dID)

		err := parseAndSaveDialog(client, dID, brief.Interlocutor)
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
		} else {
			fmt.Println("OK")
		}

		// Пауза, чтобы не нагружать сервер
		time.Sleep(400 * time.Millisecond)
	}

	fmt.Println("\n🎉 Готово! Все файлы сохранены.")
}

// --- ФУНКЦИИ ---

func makeRequest(client *http.Client, url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	data, _ := os.ReadFile(SessionFile)
	var session SessionData
	json.Unmarshal(data, &session)
	for _, c := range session.Cookies {
		req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value})
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("статус %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func fetchBriefs(client *http.Client) ([]InboxBrief, error) {
	body, err := makeRequest(client, ListURL)
	if err != nil {
		return nil, err
	}

	// Парсим JSON структуру {"briefs": [...]}
	var response InboxListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// Сохраним debug, если JSON не совпал
		os.WriteFile("debug_error_list.html", body, 0644)
		return nil, fmt.Errorf("не удалось разобрать JSON списка: %v", err)
	}

	return response.Briefs, nil
}

func parseAndSaveDialog(client *http.Client, dialogID string, knownInterlocutor string) error {
	url := fmt.Sprintf(ViewURL, dialogID)
	body, err := makeRequest(client, url)
	if err != nil {
		return err
	}

	var apiResp DromApiResponse
	// Дром может вернуть JSON с полем dialog (HTML), или просто HTML
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// Если это не JSON, считаем, что пришел HTML
		apiResp.DialogHTML = string(body)
		apiResp.Interlocutor = knownInterlocutor
	}

	// Если имя собеседника пришло пустым, берем из списка briefs
	if apiResp.Interlocutor == "" {
		apiResp.Interlocutor = knownInterlocutor
	}

	cleanMsgs := extractMessagesFromHTML(apiResp.DialogHTML, apiResp.Interlocutor)

	dump := DialogDump{
		DialogID:     dialogID,
		Interlocutor: apiResp.Interlocutor,
		SavedAt:      time.Now().Format("2006-01-02 15:04:05"),
		Messages:     cleanMsgs,
	}

	fileData, _ := json.MarshalIndent(dump, "", "    ")
	filename := fmt.Sprintf("dialog_%s.json", dialogID)
	return os.WriteFile(filename, fileData, 0644)
}

func extractMessagesFromHTML(htmlContent, interlocutorName string) []CleanMessage {
	var msgs []CleanMessage
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return msgs
	}

	doc.Find(".bzr-dialog__msg-container").Each(func(i int, s *goquery.Selection) {
		id, _ := s.Attr("data-message-id")

		block := s.Find(".bzr-dialog__message")
		if block.Length() == 0 {
			return
		}

		text := strings.TrimSpace(block.Find(".bzr-dialog__text").Text())
		if text == "" {
			text = "[Вложение]"
		}

		timeVal := strings.TrimSpace(block.Find(".bzr-dialog__message-dt").Text())

		author := interlocutorName
		direction := "incoming"

		if block.HasClass("bzr-dialog__message_out") {
			author = "Я"
			direction = "outgoing"
		}

		isRead := false
		if val, ok := block.Find(".bzr-dialog__message-check").Attr("data-state"); ok && val == "read" {
			isRead = true
		}

		msgs = append(msgs, CleanMessage{
			ID:        id,
			Author:    author,
			Direction: direction,
			Time:      timeVal,
			Text:      text,
			IsRead:    isRead,
		})
	})
	return msgs
}
