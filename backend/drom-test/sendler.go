package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Структуры для куки (те же самые)
type SessionCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SessionData struct {
	Cookies []SessionCookie `json:"cookies"`
}

const (
	SessionFile = "drom_session.json"
	PostURL     = "https://my.drom.ru/personal/messaging/view"
	UserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// 1. Читаем куки
	cookies, err := loadCookies()
	if err != nil {
		fmt.Println("Ошибка: не найден файл drom_session.json")
		return
	}
	fmt.Println("Куки загружены.")

	// 2. Ввод ID диалога
	fmt.Print("Введите ID диалога (Enter для 1771417825): ")
	dialogID, _ := reader.ReadString('\n')
	dialogID = strings.TrimSpace(dialogID)
	if dialogID == "" {
		dialogID = "1771417825"
	}
	fmt.Printf("--- Чат с %s (пишите 'exit' для выхода) ---\n", dialogID)

	client := &http.Client{}

	// 3. Цикл чата
	for {
		fmt.Print("Вы: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if strings.ToLower(text) == "exit" || strings.ToLower(text) == "quit" {
			break
		}
		if text == "" {
			continue
		}

		if sendMessage(client, cookies, dialogID, text) {
			fmt.Println("✅ Отправлено")
		} else {
			fmt.Println("❌ Ошибка отправки")
		}
	}
}

func loadCookies() ([]SessionCookie, error) {
	data, err := os.ReadFile(SessionFile)
	if err != nil {
		return nil, err
	}
	var session SessionData
	err = json.Unmarshal(data, &session)
	return session.Cookies, err
}

func sendMessage(client *http.Client, cookies []SessionCookie, dialogID, text string) bool {
	// Подготовка данных формы
	formData := url.Values{}
	formData.Set("message", text)
	formData.Set("post", "Отправить")

	// Формируем URL с параметрами
	reqUrl, _ := url.Parse(PostURL)
	q := reqUrl.Query()
	q.Add("dialogId", dialogID)
	q.Add("json", "true")
	q.Add("ajax", "1")
	q.Add("flat-layout", "false")
	reqUrl.RawQuery = q.Encode()

	// Создаем запрос
	req, err := http.NewRequest("POST", reqUrl.String(), strings.NewReader(formData.Encode()))
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Заголовки
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://my.drom.ru")
	// Важнейший заголовок Referer
	req.Header.Set("Referer", fmt.Sprintf("https://my.drom.ru/personal/messaging-modal/dialog-%s", dialogID))

	// Куки
	for _, c := range cookies {
		req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value})
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка сети:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true
	}

	fmt.Printf("HTTP Статус: %d\n", resp.StatusCode)
	return false
}
