package main

import (
	msgv1 "avtoplaneta/gen/messaging/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto конвертирует Conversation в Protobuf Conversation
func (c *Conversation) ToProto() *msgv1.Conversation {
	if c == nil {
		return nil
	}
	return &msgv1.Conversation{
		Id:            c.ID,
		Participants:  c.Participants,
		Title:         c.Title,
		CreatedAt:     timestamppb.New(c.CreatedAt),
		LastMessageAt: timestamppb.New(c.LastMessageAt),
		LastMessage:   c.LastMessage,
		UnreadCount:   int32(c.UnreadCount),
	}
}

// ToProto конвертирует Message в Protobuf Message
func (m *Message) ToProto() *msgv1.Message {
	if m == nil {
		return nil
	}
	protoMsg := &msgv1.Message{
		Id:             m.ID,
		ConversationId: m.ConversationID,
		SenderId:       m.SenderID,
		Content:        m.Content,
		MessageType:    m.MessageType,
		VoiceUrl:       m.VoiceURL,
		CreatedAt:      timestamppb.New(m.CreatedAt),
		ReadBy:         m.ReadBy,
		Attachments:    m.Attachments,
	}

	for _, mention := range m.Mentions {
		protoMsg.Mentions = append(protoMsg.Mentions, mention.ToProto())
	}

	for _, reaction := range m.Reactions {
		protoMsg.Reactions = append(protoMsg.Reactions, reaction.ToProto())
	}

	return protoMsg
}

// ToProto конвертирует Mention в Protobuf Mention
func (m *Mention) ToProto() *msgv1.Mention {
	if m == nil {
		return nil
	}
	return &msgv1.Mention{
		Id:         m.ID,
		MessageId:  m.MessageID,
		Type:       m.Type,
		ResourceId: m.ResourceID,
		Text:       m.Text,
	}
}

// ToProto конвертирует Reaction в Protobuf Reaction
func (r *Reaction) ToProto() *msgv1.Reaction {
	if r == nil {
		return nil
	}
	return &msgv1.Reaction{
		Id:        r.ID,
		MessageId: r.MessageID,
		UserId:    r.UserID,
		Emoji:     r.Emoji,
	}
}

// ToProto конвертирует Notification в Protobuf Notification
func (n *Notification) ToProto() *msgv1.Notification {
	if n == nil {
		return nil
	}
	return &msgv1.Notification{
		Id:        n.ID,
		UserId:    n.UserID,
		Type:      n.Type,
		Message:   n.Message,
		Read:      n.Read,
		CreatedAt: timestamppb.New(n.CreatedAt),
		RelatedId: n.RelatedID,
	}
}

// ToProto конвертирует UserStatus в Protobuf UserStatus
func (us *UserStatus) ToProto() *msgv1.UserStatus {
	if us == nil {
		return nil
	}
	return &msgv1.UserStatus{
		Id:       us.ID,
		UserId:   us.UserID,
		IsOnline: us.IsOnline,
		LastSeen: timestamppb.New(us.LastSeen),
	}
}
