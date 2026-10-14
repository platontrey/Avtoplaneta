package main

import (
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
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/html/charset"
)

type MessagingService interface {
	GetConversations(ctx context.Context, userID int64) ([]Conversation, error)
	GetConversation(ctx context.Context, id int64, userID int64) (*Conversation, error)
	CreateConversation(ctx context.Context, title string, participants []int64) (*Conversation, error)
	UpdateConversation(ctx context.Context, id int64, title string, participants []int64) (*Conversation, error)
	DeleteConversation(ctx context.Context, id int64) error

	GetMessages(ctx context.Context, conversationID int64) ([]Message, error)
	SendMessage(ctx context.Context, conversationID int64, senderID int64, content string, attachments []string) (*Message, error)
	SendVoiceMessage(ctx context.Context, conversationID int64, senderID int64, voiceURL string) (*Message, error)
	DeleteMessage(ctx context.Context, id int64, userID int64) error
	MarkMessageRead(ctx context.Context, messageID int64, userID int64) error

	AddReaction(ctx context.Context, messageID int64, userID int64, emoji string) (*Reaction, error)
	RemoveReaction(ctx context.Context, reactionID int64, userID int64) error

	GetNotifications(ctx context.Context, userID int64) ([]Notification, error)
	MarkNotificationRead(ctx context.Context, id int64) error
	MarkAllNotificationsRead(ctx context.Context, userID int64) error

	GetUserStatuses(ctx context.Context, userIDs []int64) ([]UserStatus, error)
	UpdateUserStatus(ctx context.Context, userID int64, isOnline bool) error
	GetUserStatus(ctx context.Context, userID int64) (*UserStatus, error)

	SearchMessages(ctx context.Context, userID int64, query string, limit, offset int) ([]Message, error)

	// Drom integration methods
	GetDromBriefs(ctx context.Context) ([]DromDialog, error)
	GetDromMessages(ctx context.Context, dialogID string) ([]DromMessage, *DromDialog, error)
	SendDromMessage(ctx context.Context, dialogID string, content string) (*DromMessage, error)
	GetDromAuthStatus(ctx context.Context) (map[string]interface{}, error)
	LoginDrom(ctx context.Context, login, password, code string) (map[string]interface{}, error)
	SaveDromSession(ctx context.Context, rawSession []byte) error
}

type messagingService struct {
	repo  MessagingRepository
	redis *redis.Client
}

func NewMessagingService(repo MessagingRepository, redisClient *redis.Client) MessagingService {
	return &messagingService{
		repo:  repo,
		redis: redisClient,
	}
}

// Cache helpers
func (s *messagingService) getMessagesFromCache(ctx context.Context, conversationID int64) ([]Message, error) {
	if s.redis == nil {
		return nil, errors.New("redis not initialized")
	}
	key := fmt.Sprintf("messages:%d", conversationID)
	val, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var messages []Message
	if err := json.Unmarshal([]byte(val), &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *messagingService) cacheMessages(ctx context.Context, conversationID int64, messages []Message) {
	if s.redis == nil {
		return
	}
	key := fmt.Sprintf("messages:%d", conversationID)
	data, err := json.Marshal(messages)
	if err != nil {
		return
	}
	s.redis.Set(ctx, key, string(data), 10*time.Minute)
}

func (s *messagingService) invalidateMessagesCache(ctx context.Context, conversationID int64) {
	if s.redis == nil {
		return
	}
	key := fmt.Sprintf("messages:%d", conversationID)
	s.redis.Del(ctx, key)
}

// Conversations
func (s *messagingService) GetConversations(ctx context.Context, userID int64) ([]Conversation, error) {
	return s.repo.GetConversations(ctx, userID)
}

func (s *messagingService) GetConversation(ctx context.Context, id int64, userID int64) (*Conversation, error) {
	conv, err := s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate access
	if userID > 0 {
		isParticipant := false
		for _, p := range conv.Participants {
			if p == userID {
				isParticipant = true
				break
			}
		}
		if !isParticipant {
			return nil, errors.New("access denied")
		}
	}

	return conv, nil
}

func (s *messagingService) CreateConversation(ctx context.Context, title string, participants []int64) (*Conversation, error) {
	conv := Conversation{
		Participants:  participants,
		Title:         title,
		CreatedAt:     time.Now(),
		LastMessageAt: time.Now(),
	}
	err := s.repo.CreateConversation(ctx, &conv)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (s *messagingService) UpdateConversation(ctx context.Context, id int64, title string, participants []int64) (*Conversation, error) {
	// GORM structure was update title or participants
	return s.repo.UpdateConversation(ctx, id, title, time.Now(), "", participants)
}

func (s *messagingService) DeleteConversation(ctx context.Context, id int64) error {
	// First invalidate cache
	s.invalidateMessagesCache(ctx, id)
	return s.repo.DeleteConversation(ctx, id)
}

// Messages
func (s *messagingService) GetMessages(ctx context.Context, conversationID int64) ([]Message, error) {
	// Try cache first
	cached, err := s.getMessagesFromCache(ctx, conversationID)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	// Database query
	messages, err := s.repo.GetMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Cache messages in background
	go s.cacheMessages(context.Background(), conversationID, messages)

	return messages, nil
}

func (s *messagingService) SendMessage(ctx context.Context, conversationID int64, senderID int64, content string, attachments []string) (*Message, error) {
	// 1. Verify conversation exists and sender is participant
	conv, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	isParticipant := false
	for _, p := range conv.Participants {
		if p == senderID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, errors.New("access denied")
	}

	// 2. Create message
	msg := Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		MessageType:    "text",
		CreatedAt:      time.Now(),
		ReadBy:         []int64{senderID},
		Attachments:    attachments,
	}

	if err := s.repo.CreateMessage(ctx, &msg); err != nil {
		return nil, err
	}

	// 3. Update last message at in conversation (fixed duplicate Save() bug here)
	if err := s.repo.UpdateConversationLastMessage(ctx, conversationID, msg.CreatedAt, msg.Content); err != nil {
		log.Printf("Warning: failed to update conversation last message: %v", err)
	}

	// 4. Invalidate cache
	go s.invalidateMessagesCache(context.Background(), conversationID)

	return &msg, nil
}

func (s *messagingService) SendVoiceMessage(ctx context.Context, conversationID int64, senderID int64, voiceURL string) (*Message, error) {
	// 1. Verify conversation exists
	conv, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	isParticipant := false
	for _, p := range conv.Participants {
		if p == senderID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, errors.New("access denied")
	}

	// 2. Create message
	msg := Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        "[Голосовое сообщение]",
		MessageType:    "voice",
		VoiceURL:       voiceURL,
		CreatedAt:      time.Now(),
		ReadBy:         []int64{senderID},
		Attachments:    []string{},
	}

	if err := s.repo.CreateMessage(ctx, &msg); err != nil {
		return nil, err
	}

	// 3. Update last message at
	if err := s.repo.UpdateConversationLastMessage(ctx, conversationID, msg.CreatedAt, msg.Content); err != nil {
		log.Printf("Warning: failed to update conversation last message: %v", err)
	}

	// 4. Invalidate cache
	go s.invalidateMessagesCache(context.Background(), conversationID)

	return &msg, nil
}

func (s *messagingService) DeleteMessage(ctx context.Context, id int64, userID int64) error {
	// First fetch the message to find the conversation ID
	// GORM code loaded first. We need to do raw query or sqlc or we can fetch conversation.
	// Actually we should write a query to delete message only if sender matches or check sender ID.
	// Since we delete by message ID in REST:
	// Let's get conversation ID to invalidate cache
	// We didn't define a GetMessageByID in sqlc, but we can easily add it or query directly,
	// or we can just use the squirrel builder / raw DB. pool.QueryRow().
	var conversationID int64
	var senderID int64
	err := s.repo.(*messagingRepository).pool.QueryRow(ctx, "SELECT conversation_id, sender_id FROM messages WHERE id = $1", id).Scan(&conversationID, &senderID)
	if err != nil {
		return err
	}

	if senderID != userID {
		return errors.New("access denied")
	}

	if err := s.repo.DeleteMessage(ctx, id); err != nil {
		return err
	}

	go s.invalidateMessagesCache(context.Background(), conversationID)
	return nil
}

func (s *messagingService) MarkMessageRead(ctx context.Context, messageID int64, userID int64) error {
	// Get message conversation ID to invalidate cache
	var conversationID int64
	err := s.repo.(*messagingRepository).pool.QueryRow(ctx, "SELECT conversation_id FROM messages WHERE id = $1", messageID).Scan(&conversationID)
	if err != nil {
		return err
	}

	if err := s.repo.AddUserToReadBy(ctx, messageID, userID); err != nil {
		return err
	}

	go s.invalidateMessagesCache(context.Background(), conversationID)
	return nil
}

// Reactions
func (s *messagingService) AddReaction(ctx context.Context, messageID int64, userID int64, emoji string) (*Reaction, error) {
	// Get conversation ID to invalidate cache
	var conversationID int64
	err := s.repo.(*messagingRepository).pool.QueryRow(ctx, "SELECT conversation_id FROM messages WHERE id = $1", messageID).Scan(&conversationID)
	if err != nil {
		return nil, err
	}

	reaction := Reaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}

	if err := s.repo.CreateReaction(ctx, &reaction); err != nil {
		return nil, err
	}

	go s.invalidateMessagesCache(context.Background(), conversationID)
	return &reaction, nil
}

func (s *messagingService) RemoveReaction(ctx context.Context, reactionID int64, userID int64) error {
	reaction, err := s.repo.GetReactionByID(ctx, reactionID)
	if err != nil {
		return err
	}

	if userID > 0 && reaction.UserID != userID {
		return errors.New("access denied")
	}

	var conversationID int64
	err = s.repo.(*messagingRepository).pool.QueryRow(ctx, "SELECT conversation_id FROM messages WHERE id = $1", reaction.MessageID).Scan(&conversationID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteReaction(ctx, reactionID); err != nil {
		return err
	}

	go s.invalidateMessagesCache(context.Background(), conversationID)
	return nil
}

// Notifications
func (s *messagingService) GetNotifications(ctx context.Context, userID int64) ([]Notification, error) {
	return s.repo.GetNotifications(ctx, userID)
}

func (s *messagingService) MarkNotificationRead(ctx context.Context, id int64) error {
	return s.repo.MarkNotificationRead(ctx, id)
}

func (s *messagingService) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	return s.repo.MarkAllNotificationsRead(ctx, userID)
}

// UserStatuses
func (s *messagingService) GetUserStatuses(ctx context.Context, userIDs []int64) ([]UserStatus, error) {
	return s.repo.GetUserStatuses(ctx, userIDs)
}

func (s *messagingService) UpdateUserStatus(ctx context.Context, userID int64, isOnline bool) error {
	status := UserStatus{
		UserID:   userID,
		IsOnline: isOnline,
		LastSeen: time.Now(),
	}
	return s.repo.UpsertUserStatus(ctx, &status)
}

func (s *messagingService) GetUserStatus(ctx context.Context, userID int64) (*UserStatus, error) {
	return s.repo.GetUserStatusByUserID(ctx, userID)
}

// Search
func (s *messagingService) SearchMessages(ctx context.Context, userID int64, query string, limit, offset int) ([]Message, error) {
	return s.repo.SearchMessages(ctx, userID, query, limit, offset)
}

// === Drom Scraper Integration ===
// === Drom Scraper Integration ===
type DromDialog struct {
	ID            int64   `json:"id"`
	DialogID      string  `json:"dialog_id"`
	Interlocutor  string  `json:"interlocutor"`
	CreatedAt     string  `json:"created_at"`
	LastMessageAt string  `json:"last_message_at"`
	LastMessage   *string `json:"last_message,omitempty"`
	IsUnread      bool    `json:"is_unread"`
	PartTitle     string  `json:"part_title,omitempty"`
	PartPrice     string  `json:"part_price,omitempty"`
	PartURL       string  `json:"part_url,omitempty"`
	PartPhotoURL  string  `json:"part_photo_url,omitempty"`
	BulletinID    string  `json:"bulletin_id,omitempty"`
	UserRating    string  `json:"user_rating,omitempty"`
	UserStatus    string  `json:"user_status,omitempty"`
}

type DromMessage struct {
	ID        int64  `json:"id"`
	DialogID  string `json:"dialog_id"`
	MessageID string `json:"message_id"`
	Author    string `json:"author"`
	Direction string `json:"direction"`
	Time      string `json:"time"`
	Text      string `json:"text"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

type InboxBrief struct {
	DialogID        int    `json:"dialogId"`
	Interlocutor    string `json:"interlocutor"`
	LastMessageDate int64  `json:"lastMessageDate"`
	IsUnread        bool   `json:"isUnread"`
	URL             string `json:"url"`
	HTML            string `json:"html"`
}

type InboxListResponse struct {
	Briefs []InboxBrief `json:"briefs"`
}

type DromApiResponse struct {
	Interlocutor        string `json:"interlocutor"`
	RelatedDialogsCount int    `json:"relatedDialogsCount"`
	HeaderHTML          string `json:"header"`
	DialogHTML          string `json:"dialog"`
}

type SessionCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SessionData struct {
	Cookies []SessionCookie `json:"cookies"`
}

const (
	DromSessionFile = "drom_session.json"
	DromListURL     = "https://my.drom.ru/personal/messaging/inbox-list?ajax=1&fromIndex=0&count=50&list=personal"
	DromViewURL     = "https://my.drom.ru/personal/messaging/view?dialogId=%s&json=true&flat-layout=false&ajax=1"
	DromPostURL     = "https://my.drom.ru/personal/messaging/view"
	DromUserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
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

	subTitle := doc.Find(".bzr-dialog-header__sub-title")
	if subTitle.Length() > 0 {
		partLink := subTitle.Find("a")
		if partLink.Length() > 0 {
			dialog.PartTitle = strings.TrimSpace(partLink.Text())
			dialog.PartURL = strings.TrimSpace(partLink.AttrOr("href", ""))
		}

		fullText := strings.ReplaceAll(subTitle.Text(), "\u00a0", " ")
		if idx := strings.Index(fullText, "•"); idx != -1 {
			dialog.PartPrice = strings.TrimSpace(fullText[:idx])
		} else if idx := strings.Index(fullText, "р."); idx != -1 {
			dialog.PartPrice = strings.TrimSpace(fullText[:idx+len("р.")])
		} else if dialog.PartTitle != "" {
			pricePart := strings.TrimSpace(strings.ReplaceAll(fullText, dialog.PartTitle, ""))
			dialog.PartPrice = strings.Trim(pricePart, " •\t\r\n")
		}
	}

	if dialog.PartURL != "" {
		if m := regexp.MustCompile(`-(\d+)\.html`).FindStringSubmatch(dialog.PartURL); len(m) > 1 {
			dialog.BulletinID = m[1]
		}
	}
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
	if err := sendDromMessageToAPI(dialogID, content); err != nil {
		return nil, err
	}

	message := DromMessage{
		ID:        time.Now().Unix(),
		DialogID:  dialogID,
		MessageID: strconv.FormatInt(time.Now().Unix(), 10),
		Author:    "Я",
		Direction: "outgoing",
		Time:      time.Now().Format("2006-01-02 15:04:05"),
		Text:      content,
		IsRead:    true,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
	}

	return &message, nil
}

func sendDromMessageToAPI(dialogID, text string) error {
	formData := url.Values{}
	formData.Set("dialogId", dialogID)
	formData.Set("text", text)
	formData.Set("ajax", "1")
	formData.Set("action", "send")

	// Read session file
	data, err := os.ReadFile(DromSessionFile)
	if err != nil {
		return fmt.Errorf("failed to read drom session: %v", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to parse drom session: %v", err)
	}

	req, err := http.NewRequest("POST", DromPostURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", DromUserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range session.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("drom post message returned status: %d", resp.StatusCode)
	}

	return nil
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
	var session SessionData

	// Attempt JSON parsing (SessionData format or array of cookies)
	if err := json.Unmarshal(rawSession, &session); err != nil || len(session.Cookies) == 0 {
		var cookies []SessionCookie
		if errArray := json.Unmarshal(rawSession, &cookies); errArray == nil && len(cookies) > 0 {
			session.Cookies = cookies
		} else {
			// Attempt parsing raw cookie header: "ring=abc; ring_session=xyz"
			cookieStr := strings.TrimSpace(string(rawSession))
			parts := strings.Split(cookieStr, ";")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if idx := strings.Index(p, "="); idx > 0 {
					session.Cookies = append(session.Cookies, SessionCookie{
						Name:  strings.TrimSpace(p[:idx]),
						Value: strings.TrimSpace(p[idx+1:]),
					})
				}
			}
		}
	}

	if len(session.Cookies) == 0 {
		return errors.New("не удалось распознать cookies в переданных данных")
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(DromSessionFile, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи %s: %w", DromSessionFile, err)
	}
	_ = os.WriteFile("../drom-test/drom_session.json", data, 0644)

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

	resp1, err := client.Do(req1)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса страницы входа: %w", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()

	// Extract CSRF token
	csrfToken := ""
	if m := regexp.MustCompile(`name=["']csrfToken["']\s+type=["']hidden["']\s+value=["']([^"']+)["']`).FindSubmatch(body1); len(m) > 1 {
		csrfToken = string(m[1])
	} else if m := regexp.MustCompile(`id=["']csrfToken["'][^>]*value=["']([^"']+)["']`).FindSubmatch(body1); len(m) > 1 {
		csrfToken = string(m[1])
	}

	if csrfToken == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "Не удалось получить CSRF токен со страницы входа Drom",
		}, nil
	}

	// 2. POST login form
	formData := url.Values{}
	formData.Set("csrfToken", csrfToken)
	formData.Set("radio", "sign")
	formData.Set("sign", strings.TrimSpace(login))
	formData.Set("password", strings.TrimSpace(password))
	if code != "" {
		formData.Set("code", strings.TrimSpace(code))
	}

	req2, err := http.NewRequestWithContext(ctx, "POST", "https://my.drom.ru/sign", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	req2.Header.Set("User-Agent", DromUserAgent)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Origin", "https://my.drom.ru")
	req2.Header.Set("Referer", "https://my.drom.ru/sign")

	resp2, err := client.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки данных авторизации: %w", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	// 3. Inspect cookies in cookiejar for drom.ru
	dromURL, _ := url.Parse("https://my.drom.ru")
	cookies := jar.Cookies(dromURL)

	hasRing := false
	var sessionCookies []SessionCookie
	for _, c := range cookies {
		sessionCookies = append(sessionCookies, SessionCookie{
			Name:  c.Name,
			Value: c.Value,
		})
		if c.Name == "ring" || c.Name == "ring_session" {
			hasRing = true
		}
	}

	finalURL := resp2.Request.URL.String()
	isLoggedIn := hasRing && !strings.Contains(finalURL, "/sign")

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

	// Check if Drom returned specific errors
	bodyStr := string(body2)
	if docReader, err := charset.NewReader(strings.NewReader(bodyStr), resp2.Header.Get("Content-Type")); err == nil {
		if decoded, err := io.ReadAll(docReader); err == nil {
			bodyStr = string(decoded)
		}
	}

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
	if doc != nil {
		if errMsg := strings.TrimSpace(doc.Find("#sign_errors, #password_errors, .formError.controlErrors").First().Text()); errMsg != "" {
			return map[string]interface{}{
				"success": false,
				"error":   errMsg,
			}, nil
		}
		if banner := strings.TrimSpace(doc.Find(".notification-banner_alert, .sign-form-banner").First().Text()); banner != "" {
			return map[string]interface{}{
				"success": false,
				"error":   banner,
			}, nil
		}
	}

	if strings.Contains(bodyStr, "code") || strings.Contains(bodyStr, "подтвержд") || strings.Contains(bodyStr, "SMS") {
		return map[string]interface{}{
			"success":   false,
			"need_code": true,
			"message":   "Требуется подтверждение входа (SMS/код)",
		}, nil
	}

	return map[string]interface{}{
		"success": false,
		"error":   "Не удалось войти на Drom. Проверьте правильность введенных данных.",
	}, nil
}
