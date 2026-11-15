package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func parseBriefHTML(htmlSnippet string, dialog *DromDialog) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlSnippet))
	if err != nil {
		return
	}

	if imgSel := doc.Find(".dialog-brief__image"); imgSel.Length() > 0 {
		style, _ := imgSel.Attr("style")
		if m := regexp.MustCompile(`url\(([^)]+)\)`).FindStringSubmatch(style); len(m) > 1 {
			dialog.PartPhotoURL = strings.Trim(m[1], `"' `)
		}
	}

	if msg := strings.TrimSpace(doc.Find(".dialog-brief__latest_msg").Text()); msg != "" {
		dialog.LastMessage = &msg
	}

	if dt := strings.TrimSpace(doc.Find(".bzr-dialog__message-dt").Text()); dt != "" && dialog.LastMessageAt == "" {
		dialog.LastMessageAt = dt
	}
}

func parseDromPartPrice(fullText, partTitle string) string {
	if idx := strings.Index(fullText, "•"); idx != -1 {
		return strings.TrimSpace(fullText[:idx])
	}
	if idx := strings.Index(fullText, "р."); idx != -1 {
		return strings.TrimSpace(fullText[:idx+len("р.")])
	}
	if partTitle != "" {
		pricePart := strings.TrimSpace(strings.ReplaceAll(fullText, partTitle, ""))
		return strings.Trim(pricePart, " •\t\r\n")
	}
	return ""
}

func parseDromSubtitle(subTitle *goquery.Selection, dialog *DromDialog) {
	if subTitle.Length() == 0 {
		return
	}
	if partLink := subTitle.Find("a"); partLink.Length() > 0 {
		dialog.PartTitle = strings.TrimSpace(partLink.Text())
		dialog.PartURL = strings.TrimSpace(partLink.AttrOr("href", ""))
	}

	fullText := strings.ReplaceAll(subTitle.Text(), "\u00a0", " ")
	if price := parseDromPartPrice(fullText, dialog.PartTitle); price != "" {
		dialog.PartPrice = price
	}
}

func parseDromHeaderHTML(headerHTML string, dialog *DromDialog) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(headerHTML))
	if err != nil {
		return
	}

	if nick := strings.TrimSpace(doc.Find(".userNick a").Text()); nick != "" {
		dialog.Interlocutor = nick
	}

	if rating := strings.TrimSpace(doc.Find(".userRating a, .ratingPositive, .ratingNegative, .userRating").First().Text()); rating != "" {
		dialog.UserRating = rating
	}

	if status := strings.TrimSpace(doc.Find("[data-role='last-seen'], .bzr-dialog__user-meta").Text()); status != "" {
		dialog.UserStatus = status
	}

	parseDromSubtitle(doc.Find(".bzr-dialog-header__sub-title"), dialog)

	if dialog.PartURL != "" {
		if m := regexp.MustCompile(`-(\d+)\.html`).FindStringSubmatch(dialog.PartURL); len(m) > 1 {
			dialog.BulletinID = m[1]
		}
	}
}

func extractDromMessagesFromHTML(htmlContent, interlocutorName string) []DromMessage {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("Failed to parse Drom HTML: %v", err)
		return nil
	}

	var messages []DromMessage

	// Modern Drom / Baza (.bzr-dialog__msg-container)
	bzrContainers := doc.Find(".bzr-dialog__msg-container")
	if bzrContainers.Length() > 0 {
		bzrContainers.Each(func(i int, s *goquery.Selection) {
			messageID, _ := s.Attr("data-message-id")
			block := s.Find(".bzr-dialog__message")
			if block.Length() == 0 {
				block = s
			}

			text := strings.TrimSpace(block.Find(".bzr-dialog__text").Text())
			if text == "" {
				text = strings.TrimSpace(block.Find(".message-text").Text())
			}
			if text == "" {
				text = "[Вложение]"
			}

			timeStr := strings.TrimSpace(block.Find(".bzr-dialog__message-dt").Text())
			if timeStr == "" {
				timeStr = strings.TrimSpace(block.Find(".message-time").Text())
			}

			isOutgoing := block.HasClass("bzr-dialog__message_out") || block.HasClass("message-outgoing")
			author := interlocutorName
			direction := "incoming"
			if isOutgoing {
				author = "Я"
				direction = "outgoing"
			}

			isRead := false
			if val, ok := block.Find(".bzr-dialog__message-check").Attr("data-state"); ok && val == "read" {
				isRead = true
			} else if isOutgoing {
				isRead = true
			}

			messages = append(messages, DromMessage{
				ID:        int64(i + 1),
				MessageID: messageID,
				Author:    author,
				Direction: direction,
				Time:      timeStr,
				Text:      text,
				IsRead:    isRead,
				CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
			})
		})
		return messages
	}

	// Legacy Drom fallback (.message-item)
	doc.Find(".message-item").Each(func(i int, s *goquery.Selection) {
		messageID, _ := s.Attr("data-message-id")
		text := strings.TrimSpace(s.Find(".message-text").Text())
		timeStr := strings.TrimSpace(s.Find(".message-time").Text())
		isOutgoing := s.HasClass("message-outgoing")

		author := interlocutorName
		direction := "incoming"
		if isOutgoing {
			author = "Я"
			direction = "outgoing"
		}

		messages = append(messages, DromMessage{
			ID:        int64(i + 1),
			MessageID: messageID,
			Author:    author,
			Direction: direction,
			Time:      timeStr,
			Text:      text,
			IsRead:    true,
			CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
		})
	})

	return messages
}

func extractCSRFTokenFromBody(body []byte, doc *goquery.Document) string {
	if doc != nil {
		if val, exists := doc.Find("input[name='csrfToken']").Attr("value"); exists && val != "" {
			return strings.TrimSpace(val)
		}
		if val, exists := doc.Find("#csrfToken").Attr("value"); exists && val != "" {
			return strings.TrimSpace(val)
		}
	}
	patterns := []string{
		`<input[^>]*name=["']csrfToken["'][^>]*value=["']([^"']+)["']`,
		`<input[^>]*value=["']([^"']+)["'][^>]*name=["']csrfToken["']`,
		`name=["']csrfToken["']\s+type=["']hidden["']\s+value=["']([^"']+)["']`,
	}
	for _, p := range patterns {
		if m := regexp.MustCompile(p).FindSubmatch(body); len(m) > 1 {
			return string(m[1])
		}
	}
	return ""
}

func inspectSessionCookies(jar http.CookieJar) ([]SessionCookie, bool, bool) {
	seenCookies := make(map[string]bool)
	var sessionCookies []SessionCookie
	hasRing := false
	hasAuthCookie := false

	for _, hostURL := range []string{"https://my.drom.ru", "https://baza.drom.ru", "https://drom.ru"} {
		u, _ := url.Parse(hostURL)
		for _, c := range jar.Cookies(u) {
			if !seenCookies[c.Name] {
				seenCookies[c.Name] = true
				sessionCookies = append(sessionCookies, SessionCookie{
					Name:  c.Name,
					Value: c.Value,
				})
			}
			if c.Name == "ring" || c.Name == "ring_session" {
				hasRing = true
			}
			if c.Name == "uid" || c.Name == "login" || c.Name == "boobs" {
				hasAuthCookie = true
			}
		}
	}
	return sessionCookies, hasAuthCookie, hasRing
}

func parseDromCookies(raw []byte) []SessionCookie {
	// 1. Попытка распарсить как SessionData {"cookies": [...]}
	var session SessionData
	if err := json.Unmarshal(raw, &session); err == nil && len(session.Cookies) > 0 {
		return session.Cookies
	}

	// 2. Попытка распарсить как массив cookies [...]
	var cookies []SessionCookie
	if err := json.Unmarshal(raw, &cookies); err == nil && len(cookies) > 0 {
		return cookies
	}

	// 3. Попытка распарсить как строку заголовка "ring=...; ring_session=..."
	return parseRawCookieHeader(string(raw))
}

func parseRawCookieHeader(cookieStr string) []SessionCookie {
	cookieStr = strings.TrimSpace(cookieStr)
	cookieStr = strings.Trim(cookieStr, "\"'")
	if strings.HasPrefix(strings.ToLower(cookieStr), "cookie:") {
		cookieStr = strings.TrimSpace(cookieStr[7:])
	}

	var cookies []SessionCookie
	for _, part := range strings.Split(cookieStr, ";") {
		cookie, ok := parseCookiePair(part)
		if ok {
			cookies = append(cookies, cookie)
		}
	}
	return cookies
}

func parseCookiePair(part string) (SessionCookie, bool) {
	part = strings.TrimSpace(part)
	if strings.HasPrefix(strings.ToLower(part), "cookie:") {
		part = strings.TrimSpace(part[7:])
	}

	idx := strings.Index(part, "=")
	if idx <= 0 {
		return SessionCookie{}, false
	}

	name := strings.TrimSpace(part[:idx])
	val := strings.Trim(strings.TrimSpace(part[idx+1:]), "\"'")
	if name == "" || val == "" {
		return SessionCookie{}, false
	}

	return SessionCookie{Name: name, Value: val}, true
}
