package main

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
