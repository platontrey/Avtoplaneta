package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"messaging-service/db/sqlc"
)

type MessagingRepository interface {
	GetConversations(ctx context.Context, userID int64) ([]Conversation, error)
	GetConversation(ctx context.Context, id int64) (*Conversation, error)
	CreateConversation(ctx context.Context, conv *Conversation) error
	UpdateConversation(ctx context.Context, id int64, title string, lastMessageAt time.Time, lastMessage string, participants []int64) (*Conversation, error)
	UpdateConversationLastMessage(ctx context.Context, id int64, lastMessageAt time.Time, lastMessage string) error
	DeleteConversation(ctx context.Context, id int64) error

	GetMessages(ctx context.Context, conversationID int64) ([]Message, error)
	CreateMessage(ctx context.Context, msg *Message) error
	DeleteMessage(ctx context.Context, id int64) error
	DeleteMessagesByConversationID(ctx context.Context, conversationID int64) error
	AddUserToReadBy(ctx context.Context, messageID int64, userID int64) error

	GetReactionByID(ctx context.Context, id int64) (*Reaction, error)
	CreateReaction(ctx context.Context, reaction *Reaction) error
	DeleteReaction(ctx context.Context, id int64) error
	DeleteReactionByFields(ctx context.Context, messageID int64, userID int64, emoji string) error

	CreateMention(ctx context.Context, mention *Mention) error

	GetNotifications(ctx context.Context, userID int64) ([]Notification, error)
	CreateNotification(ctx context.Context, notification *Notification) error
	MarkNotificationRead(ctx context.Context, id int64) error
	MarkAllNotificationsRead(ctx context.Context, userID int64) error

	GetUserStatuses(ctx context.Context, userIDs []int64) ([]UserStatus, error)
	GetUserStatusByUserID(ctx context.Context, userID int64) (*UserStatus, error)
	UpsertUserStatus(ctx context.Context, status *UserStatus) error

	SearchMessages(ctx context.Context, userID int64, query string, limit, offset int) ([]Message, error)
}

type messagingRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewMessagingRepository(pool *pgxpool.Pool) MessagingRepository {
	return &messagingRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Helpers for time conversion
func toTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()}
}

// Conversations mappings
func sqlcConversationToDomain(c sqlc.Conversation) Conversation {
	return Conversation{
		ID:            c.ID,
		Participants:  c.Participants,
		Title:         c.Title,
		CreatedAt:     toTime(c.CreatedAt),
		LastMessageAt: toTime(c.LastMessageAt),
		LastMessage:   c.LastMessage,
	}
}

func (r *messagingRepository) GetConversations(ctx context.Context, userID int64) ([]Conversation, error) {
	rows, err := r.queries.GetConversations(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]Conversation, len(rows))
	for i, row := range rows {
		result[i] = Conversation{
			ID:            row.ID,
			Participants:  row.Participants,
			Title:         row.Title,
			CreatedAt:     toTime(row.CreatedAt),
			LastMessageAt: toTime(row.LastMessageAt),
			LastMessage:   row.LastMessage,
			UnreadCount:   row.UnreadCount,
		}
	}
	return result, nil
}

func (r *messagingRepository) GetConversation(ctx context.Context, id int64) (*Conversation, error) {
	c, err := r.queries.GetConversation(ctx, id)
	if err != nil {
		return nil, err
	}
	domain := sqlcConversationToDomain(c)
	return &domain, nil
}

func (r *messagingRepository) CreateConversation(ctx context.Context, conv *Conversation) error {
	params := sqlc.CreateConversationParams{
		Participants:  conv.Participants,
		Title:         conv.Title,
		CreatedAt:     toTimestamptz(conv.CreatedAt),
		LastMessageAt: toTimestamptz(conv.LastMessageAt),
		LastMessage:   conv.LastMessage,
	}

	c, err := r.queries.CreateConversation(ctx, params)
	if err != nil {
		return err
	}
	conv.ID = c.ID
	return nil
}

func (r *messagingRepository) UpdateConversation(ctx context.Context, id int64, title string, lastMessageAt time.Time, lastMessage string, participants []int64) (*Conversation, error) {
	params := sqlc.UpdateConversationParams{
		ID:            id,
		Title:         title,
		LastMessageAt: toTimestamptz(lastMessageAt),
		LastMessage:   lastMessage,
		Participants:  participants,
	}

	c, err := r.queries.UpdateConversation(ctx, params)
	if err != nil {
		return nil, err
	}
	domain := sqlcConversationToDomain(c)
	return &domain, nil
}

func (r *messagingRepository) UpdateConversationLastMessage(ctx context.Context, id int64, lastMessageAt time.Time, lastMessage string) error {
	params := sqlc.UpdateConversationLastMessageParams{
		ID:            id,
		LastMessageAt: toTimestamptz(lastMessageAt),
		LastMessage:   lastMessage,
	}
	return r.queries.UpdateConversationLastMessage(ctx, params)
}

func (r *messagingRepository) DeleteConversation(ctx context.Context, id int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	if err := qtx.DeleteMessagesByConversationID(ctx, id); err != nil {
		return err
	}
	if err := qtx.DeleteConversation(ctx, id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Messages mappings
func sqlcMessageToDomain(m sqlc.Message) Message {
	return Message{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Content:        m.Content,
		MessageType:    m.MessageType,
		VoiceURL:       m.VoiceUrl,
		CreatedAt:      toTime(m.CreatedAt),
		ReadBy:         m.ReadBy,
		Attachments:    m.Attachments,
	}
}

func (r *messagingRepository) populateMessagesRelations(ctx context.Context, dbMsgs []sqlc.Message) ([]Message, error) {
	if len(dbMsgs) == 0 {
		return []Message{}, nil
	}

	msgIDs := make([]int64, len(dbMsgs))
	for i, m := range dbMsgs {
		msgIDs[i] = m.ID
	}

	reactions, err := r.queries.GetReactionsByMessageIDs(ctx, msgIDs)
	if err != nil {
		return nil, err
	}

	mentions, err := r.queries.GetMentionsByMessageIDs(ctx, msgIDs)
	if err != nil {
		return nil, err
	}

	reactionsMap := make(map[int64][]Reaction)
	for _, re := range reactions {
		reactionsMap[re.MessageID] = append(reactionsMap[re.MessageID], Reaction{
			ID:        re.ID,
			MessageID: re.MessageID,
			UserID:    re.UserID,
			Emoji:     re.Emoji,
		})
	}

	mentionsMap := make(map[int64][]Mention)
	for _, me := range mentions {
		mentionsMap[me.MessageID] = append(mentionsMap[me.MessageID], Mention{
			ID:         me.ID,
			MessageID:  me.MessageID,
			Type:       me.Type,
			ResourceID: me.ResourceID,
			Text:       me.Text,
		})
	}

	result := make([]Message, len(dbMsgs))
	for i, m := range dbMsgs {
		domain := sqlcMessageToDomain(m)
		domain.Reactions = reactionsMap[m.ID]
		domain.Mentions = mentionsMap[m.ID]
		if domain.Reactions == nil {
			domain.Reactions = []Reaction{}
		}
		if domain.Mentions == nil {
			domain.Mentions = []Mention{}
		}
		result[i] = domain
	}

	return result, nil
}

func (r *messagingRepository) GetMessages(ctx context.Context, conversationID int64) ([]Message, error) {
	rows, err := r.queries.GetMessagesByConversationID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	return r.populateMessagesRelations(ctx, rows)
}

func (r *messagingRepository) CreateMessage(ctx context.Context, msg *Message) error {
	params := sqlc.CreateMessageParams{
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		MessageType:    msg.MessageType,
		VoiceUrl:       msg.VoiceURL,
		CreatedAt:      toTimestamptz(msg.CreatedAt),
		ReadBy:         msg.ReadBy,
		Attachments:    msg.Attachments,
	}

	m, err := r.queries.CreateMessage(ctx, params)
	if err != nil {
		return err
	}
	msg.ID = m.ID
	return nil
}

func (r *messagingRepository) DeleteMessage(ctx context.Context, id int64) error {
	return r.queries.DeleteMessage(ctx, id)
}

func (r *messagingRepository) DeleteMessagesByConversationID(ctx context.Context, conversationID int64) error {
	return r.queries.DeleteMessagesByConversationID(ctx, conversationID)
}

func (r *messagingRepository) AddUserToReadBy(ctx context.Context, messageID int64, userID int64) error {
	params := sqlc.AddUserToReadByParams{
		MessageID: messageID,
		UserID:    userID,
	}
	return r.queries.AddUserToReadBy(ctx, params)
}

// Reactions
func (r *messagingRepository) GetReactionByID(ctx context.Context, id int64) (*Reaction, error) {
	re, err := r.queries.GetReactionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Reaction{
		ID:        re.ID,
		MessageID: re.MessageID,
		UserID:    re.UserID,
		Emoji:     re.Emoji,
	}, nil
}

func (r *messagingRepository) CreateReaction(ctx context.Context, reaction *Reaction) error {
	params := sqlc.CreateReactionParams{
		MessageID: reaction.MessageID,
		UserID:    reaction.UserID,
		Emoji:     reaction.Emoji,
	}

	re, err := r.queries.CreateReaction(ctx, params)
	if err != nil {
		return err
	}
	reaction.ID = re.ID
	return nil
}

func (r *messagingRepository) DeleteReaction(ctx context.Context, id int64) error {
	return r.queries.DeleteReaction(ctx, id)
}

func (r *messagingRepository) DeleteReactionByFields(ctx context.Context, messageID int64, userID int64, emoji string) error {
	params := sqlc.DeleteReactionByFieldsParams{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}
	return r.queries.DeleteReactionByFields(ctx, params)
}

// Mentions
func (r *messagingRepository) CreateMention(ctx context.Context, mention *Mention) error {
	params := sqlc.CreateMentionParams{
		MessageID:  mention.MessageID,
		Type:       mention.Type,
		ResourceID: mention.ResourceID,
		Text:       mention.Text,
	}

	m, err := r.queries.CreateMention(ctx, params)
	if err != nil {
		return err
	}
	mention.ID = m.ID
	return nil
}

// Notifications
func sqlcNotificationToDomain(n sqlc.Notification) Notification {
	return Notification{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      n.Type,
		Message:   n.Message,
		Read:      n.Read,
		CreatedAt: toTime(n.CreatedAt),
		RelatedID: n.RelatedID,
	}
}

func (r *messagingRepository) GetNotifications(ctx context.Context, userID int64) ([]Notification, error) {
	rows, err := r.queries.GetNotificationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]Notification, len(rows))
	for i, row := range rows {
		result[i] = sqlcNotificationToDomain(row)
	}
	return result, nil
}

func (r *messagingRepository) CreateNotification(ctx context.Context, notification *Notification) error {
	params := sqlc.CreateNotificationParams{
		UserID:    notification.UserID,
		Type:      notification.Type,
		Message:   notification.Message,
		Read:      notification.Read,
		CreatedAt: toTimestamptz(notification.CreatedAt),
		RelatedID: notification.RelatedID,
	}

	n, err := r.queries.CreateNotification(ctx, params)
	if err != nil {
		return err
	}
	notification.ID = n.ID
	return nil
}

func (r *messagingRepository) MarkNotificationRead(ctx context.Context, id int64) error {
	return r.queries.MarkNotificationRead(ctx, id)
}

func (r *messagingRepository) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	return r.queries.MarkAllNotificationsRead(ctx, userID)
}

// UserStatus
func sqlcUserStatusToDomain(us sqlc.UserStatus) UserStatus {
	return UserStatus{
		ID:       us.ID,
		UserID:   us.UserID,
		IsOnline: us.IsOnline,
		LastSeen: toTime(us.LastSeen),
	}
}

func (r *messagingRepository) GetUserStatuses(ctx context.Context, userIDs []int64) ([]UserStatus, error) {
	var rows pgx.Rows
	var err error
	if len(userIDs) > 0 {
		rows, err = r.pool.Query(ctx, "SELECT id, user_id, is_online, last_seen FROM user_status WHERE user_id = ANY($1)", userIDs)
	} else {
		rows, err = r.pool.Query(ctx, "SELECT id, user_id, is_online, last_seen FROM user_status")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []UserStatus
	for rows.Next() {
		var us sqlc.UserStatus
		if err := rows.Scan(&us.ID, &us.UserID, &us.IsOnline, &us.LastSeen); err != nil {
			return nil, err
		}
		result = append(result, sqlcUserStatusToDomain(us))
	}
	return result, nil
}

func (r *messagingRepository) GetUserStatusByUserID(ctx context.Context, userID int64) (*UserStatus, error) {
	us, err := r.queries.GetUserStatusByUserID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Return nil, nil when status doesn't exist yet
		}
		return nil, err
	}
	domain := sqlcUserStatusToDomain(us)
	return &domain, nil
}

func (r *messagingRepository) UpsertUserStatus(ctx context.Context, status *UserStatus) error {
	params := sqlc.UpsertUserStatusParams{
		UserID:   status.UserID,
		IsOnline: status.IsOnline,
		LastSeen: toTimestamptz(status.LastSeen),
	}

	us, err := r.queries.UpsertUserStatus(ctx, params)
	if err != nil {
		return err
	}
	status.ID = us.ID
	return nil
}

// Search
func (r *messagingRepository) SearchMessages(ctx context.Context, userID int64, query string, limit, offset int) ([]Message, error) {
	var rows []sqlc.Message
	var err error
	if userID > 0 {
		params := sqlc.SearchMessagesWithUserParams{
			UserID: userID,
			Query:  query,
			Limit:  int32(limit),
			Offset: int32(offset),
		}
		rows, err = r.queries.SearchMessagesWithUser(ctx, params)
	} else {
		params := sqlc.SearchMessagesAllParams{
			Query:  query,
			Limit:  int32(limit),
			Offset: int32(offset),
		}
		rows, err = r.queries.SearchMessagesAll(ctx, params)
	}

	if err != nil {
		return nil, err
	}

	return r.populateMessagesRelations(ctx, rows)
}
