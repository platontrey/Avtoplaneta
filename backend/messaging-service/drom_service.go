package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

func makeDromRequest(reqURL string) ([]byte, error) {
	// Read session file
	data, err := os.ReadFile(DromSessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read drom session: %v", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse drom session: %v", err)
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", DromUserAgent)
	for _, cookie := range session.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("drom API returned status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (s *messagingService) GetDromBriefs(ctx context.Context) ([]DromDialog, error) {
	body, err := makeDromRequest(DromListURL)
	if err != nil {
		return nil, err
	}

	var response InboxListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse briefs: %v", err)
	}

	dialogs := make([]DromDialog, 0, len(response.Briefs))
	for _, brief := range response.Briefs {
		dialog := DromDialog{
			ID:           int64(brief.DialogID),
			DialogID:     strconv.Itoa(brief.DialogID),
			Interlocutor: brief.Interlocutor,
			IsUnread:     brief.IsUnread,
		}

		if brief.LastMessageDate > 0 {
			dialog.LastMessageAt = time.Unix(brief.LastMessageDate, 0).Format("2006-01-02 15:04:05")
		}

		if brief.HTML != "" {
			parseBriefHTML(brief.HTML, &dialog)
		}

		dialogs = append(dialogs, dialog)
	}

	return dialogs, nil
}

func (s *messagingService) GetDromMessages(ctx context.Context, dialogID string) ([]DromMessage, *DromDialog, error) {
	url := fmt.Sprintf(DromViewURL, dialogID)
	body, err := makeDromRequest(url)
	if err != nil {
		return nil, nil, err
	}

	var apiResp DromApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// If not JSON, assume raw HTML
		apiResp.DialogHTML = string(body)
	}

	dialogIDInt, _ := strconv.ParseInt(dialogID, 10, 64)
	dialogMeta := &DromDialog{
		ID:           dialogIDInt,
		DialogID:     dialogID,
		Interlocutor: apiResp.Interlocutor,
	}

	if apiResp.HeaderHTML != "" {
		parseDromHeaderHTML(apiResp.HeaderHTML, dialogMeta)
	}

	messages := extractDromMessagesFromHTML(apiResp.DialogHTML, dialogMeta.Interlocutor)
	for i := range messages {
		messages[i].DialogID = dialogID
	}

	if dialogMeta.BulletinID == "" && apiResp.DialogHTML != "" {
		if m := regexp.MustCompile(`bulletin:(\d+)`).FindStringSubmatch(apiResp.DialogHTML); len(m) > 1 {
			dialogMeta.BulletinID = m[1]
		} else if m := regexp.MustCompile(`["']bullId["']\s*:\s*(\d+)`).FindStringSubmatch(apiResp.DialogHTML); len(m) > 1 {
			dialogMeta.BulletinID = m[1]
		}
	}

	return messages, dialogMeta, nil
}

func (s *messagingService) SendDromMessage(ctx context.Context, dialogID string, content string) (*DromMessage, error) {
	msgID, err := sendDromMessageToAPI(dialogID, content)
	if err != nil {
		return nil, err
	}

	messageIDStr := strconv.FormatInt(time.Now().Unix(), 10)
	if msgID > 0 {
		messageIDStr = strconv.FormatInt(msgID, 10)
	}

	message := DromMessage{
		ID:        time.Now().Unix(),
		DialogID:  dialogID,
		MessageID: messageIDStr,
		Author:    "Я",
		Direction: "outgoing",
		Time:      time.Now().Format("15:04"),
		Text:      content,
		IsRead:    true,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
	}

	return &message, nil
}

func sendDromMessageToAPI(dialogID, text string) (int64, error) {
	postURL := fmt.Sprintf("https://my.drom.ru/personal/messaging/view?dialogId=%s&ajax=1", url.QueryEscape(dialogID))

	formData := url.Values{}
	formData.Set("dialogId", dialogID)
	formData.Set("message", text)
	formData.Set("post", "Отправить")
	formData.Set("ajax", "1")

	// Read session file
	data, err := os.ReadFile(DromSessionFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read drom session: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return 0, fmt.Errorf("failed to parse drom session: %w", err)
	}

	req, err := http.NewRequest("POST", postURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", DromUserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Origin", "https://my.drom.ru")
	req.Header.Set("Referer", fmt.Sprintf("https://my.drom.ru/personal/messaging/view?dialogId=%s", url.QueryEscape(dialogID)))
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	for _, cookie := range session.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("network error sending message to drom: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("drom post message returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		StatusCode      int    `json:"statusCode"`
		SendingStatus   string `json:"sendingStatus"`
		LatestMessageId int64  `json:"latestMessageId"`
		Error           string `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err == nil {
		if result.StatusCode != 0 && result.StatusCode != 200 {
			errMsg := result.Error
			if errMsg == "" {
				errMsg = fmt.Sprintf("status %d", result.StatusCode)
			}
			return 0, fmt.Errorf("drom rejected message: %s", errMsg)
		}
		return result.LatestMessageId, nil
	}

	return 0, nil
}

func (s *messagingService) GetDromAuthStatus(ctx context.Context) (map[string]interface{}, error) {
	data, err := os.ReadFile(DromSessionFile)
	if err != nil {
		return map[string]interface{}{
			"authenticated": false,
			"error":         "Файл сессии drom_session.json не найден",
		}, nil
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil || len(session.Cookies) == 0 {
		return map[string]interface{}{
			"authenticated": false,
			"error":         "Некорректный формат сессии",
		}, nil
	}

	// Test connection to Drom
	body, err := makeDromRequest(DromListURL)
	if err != nil {
		return map[string]interface{}{
			"authenticated": false,
			"error":         fmt.Sprintf("Ошибка проверки сессии: %v", err),
		}, nil
	}

	var response InboxListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return map[string]interface{}{
			"authenticated": false,
			"error":         "Сессия Drom истекла или недействительна",
		}, nil
	}

	return map[string]interface{}{
		"authenticated": true,
		"dialogs_count": len(response.Briefs),
		"checked_at":    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *messagingService) SaveDromSession(ctx context.Context, rawSession []byte) error {
	cookies := parseDromCookies(rawSession)
	if len(cookies) == 0 {
		return errors.New("не удалось распознать cookies в переданных данных (скопируйте строку с ring=... или json)")
	}

	session := SessionData{Cookies: cookies}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(DromSessionFile, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи %s: %w", DromSessionFile, err)
	}
	_ = os.WriteFile("../drom-test/drom_session.json", data, 0644)

	// Validate the session with Drom
	body, err := makeDromRequest(DromListURL)
	if err != nil {
		return fmt.Errorf("сессия сохранена, но Drom не принял её (проверьте, что сессия активна): %w", err)
	}

	var response InboxListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return errors.New("сессия сохранена, но Drom вернул недействительный ответ (возможно, не скопирован ключ ring или сессия истекла)")
	}

	log.Printf("[DROM SESSION] Session successfully verified with Drom, active dialogs count: %d", len(response.Briefs))
	return nil
}

func (s *messagingService) LoginDrom(ctx context.Context, login, password, code string) (map[string]interface{}, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 20 * time.Second,
	}

	// 1. Initial GET to fetch CSRF token and set initial cookies
	req1, err := http.NewRequestWithContext(ctx, "GET", "https://my.drom.ru/sign", nil)
	if err != nil {
		return nil, err
	}
	req1.Header.Set("User-Agent", DromUserAgent)
	req1.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req1.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	resp1, err := client.Do(req1)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса страницы входа: %w", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()

	// Check if Drom redirected to anti-bot verification (reCAPTCHA / SmartCaptcha)
	finalURL1 := resp1.Request.URL.String()
	if strings.Contains(resp1.Request.URL.Path, "verify") || strings.Contains(finalURL1, "/verify") {
		log.Printf("[DROM LOGIN] Drom redirected to captcha challenge: %s", finalURL1)
		return map[string]interface{}{
			"success":          false,
			"captcha_required": true,
			"error":            "Drom заблокировал страницу входа проверкой от ботов (reCAPTCHA). Автоматический вход по логину/паролю заблокирован сервером Drom. Пожалуйста, используйте вкладку «Вставка Cookie» — это занимает 30 секунд и надёжно работает.",
		}, nil
	}

	doc1, _ := goquery.NewDocumentFromReader(bytes.NewReader(body1))
	csrfToken := extractCSRFTokenFromBody(body1, doc1)

	if csrfToken == "" {
		if doc1 != nil && doc1.Find("input[name='g-recaptcha-response'], #send-button, .captcha").Length() > 0 {
			return map[string]interface{}{
				"success":          false,
				"captcha_required": true,
				"error":            "Drom заблокировал страницу входа проверкой reCAPTCHA. Программный вход через логин/пароль заблокирован. Пожалуйста, воспользуйтесь авторизацией по Cookie (вкладка «Вставка Cookie»).",
			}, nil
		}
		return map[string]interface{}{
			"success": false,
			"error":   "Не удалось получить CSRF токен со страницы входа Drom (страница защищена капчей). Воспользуйтесь вкладкой «Вставка Cookie».",
		}, nil
	}

	// 2. Clean login input (remove spaces and dashes from phone numbers)
	cleanLogin := strings.TrimSpace(login)
	if strings.ContainsAny(cleanLogin, " ()-") {
		cleanLogin = strings.ReplaceAll(cleanLogin, " ", "")
		cleanLogin = strings.ReplaceAll(cleanLogin, "-", "")
		cleanLogin = strings.ReplaceAll(cleanLogin, "(", "")
		cleanLogin = strings.ReplaceAll(cleanLogin, ")", "")
	}

	// 3. POST login form
	formData := url.Values{}
	formData.Set("csrfToken", csrfToken)
	formData.Set("radio", "sign")
	formData.Set("sign", cleanLogin)
	formData.Set("password", strings.TrimSpace(password))
	formData.Set("login_by_password", "1")
	if code != "" {
		formData.Set("code", strings.TrimSpace(code))
	}

	req2, err := http.NewRequestWithContext(ctx, "POST", "https://my.drom.ru/sign?login_by_password=1", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	req2.Header.Set("User-Agent", DromUserAgent)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Origin", "https://my.drom.ru")
	req2.Header.Set("Referer", "https://my.drom.ru/sign")
	req2.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req2.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	resp2, err := client.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки данных авторизации: %w", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	// 4. Inspect cookies across all relevant Drom domains
	sessionCookies, hasAuthCookie, hasRing := inspectSessionCookies(jar)

	finalURL := resp2.Request.URL.String()
	isLoggedIn := (hasAuthCookie || hasRing) && !strings.Contains(finalURL, "/sign")

	if isLoggedIn {
		sessionData := SessionData{Cookies: sessionCookies}
		data, err := json.MarshalIndent(sessionData, "", "  ")
		if err == nil {
			_ = os.WriteFile(DromSessionFile, data, 0644)
			_ = os.WriteFile("../drom-test/drom_session.json", data, 0644)
		}

		return map[string]interface{}{
			"success": true,
			"message": "Успешная авторизация на Drom",
		}, nil
	}

	// 5. Check if Drom returned specific errors or 2FA challenge
	// Correctly decode body directly from raw bytes using the response content-type
	var bodyStr string
	if docReader, err := charset.NewReader(bytes.NewReader(body2), resp2.Header.Get("Content-Type")); err == nil {
		if decoded, err := io.ReadAll(docReader); err == nil {
			bodyStr = string(decoded)
		}
	}
	if bodyStr == "" {
		bodyStr = string(body2)
	}

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
	if doc != nil {
		// Field-level error messages (e.g. wrong password or login)
		if errMsg := strings.TrimSpace(doc.Find("#sign_errors, #password_errors, .formError.controlErrors").First().Text()); errMsg != "" {
			return map[string]interface{}{
				"success": false,
				"error":   errMsg,
			}, nil
		}
		// Banner alert (e.g. "Пользователь не найден")
		if banner := strings.TrimSpace(doc.Find(".notification-banner__title-l, .notification-banner_alert, .sign-form-banner").First().Text()); banner != "" {
			desc := strings.TrimSpace(doc.Find(".notification-banner__content").First().Text())
			if desc != "" && desc != banner {
				banner = banner + ". " + desc
			}
			return map[string]interface{}{
				"success": false,
				"error":   banner,
			}, nil
		}
		// Two-factor authentication (SMS or app code)
		if doc.Find("input[name='code'], #code, .code-input").Length() > 0 ||
			strings.Contains(bodyStr, "подтвержд") ||
			strings.Contains(bodyStr, "код подтверждения") ||
			strings.Contains(bodyStr, "SMS") ||
			strings.Contains(bodyStr, "Telegram") {
			return map[string]interface{}{
				"success":   false,
				"need_code": true,
				"message":   "Требуется подтверждение входа (SMS или код из приложения)",
			}, nil
		}
	}

	return map[string]interface{}{
		"success": false,
		"error":   "Не удалось войти на Drom. Проверьте правильность введенных данных или воспользуйтесь авторизацией по Cookie.",
	}, nil
}
