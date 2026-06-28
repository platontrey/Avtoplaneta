package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	msgv1 "avtoplaneta/gen/messaging/v1"
)

// messagingGRPCServer реализует gRPC-сервер для MessagingService
type messagingGRPCServer struct {
	msgv1.UnimplementedMessagingServiceServer
}

// NewMessagingGRPCServer создаёт новый gRPC-сервер
func NewMessagingGRPCServer() *messagingGRPCServer {
	return &messagingGRPCServer{}
}

// ─── Conversations ──────────────────────────────────────────────────────────

func (s *messagingGRPCServer) GetConversations(ctx context.Context, req *msgv1.GetConversationsRequest) (*msgv1.ConversationList, error) {
	conversations, err := svc.GetConversations(ctx, int64(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить диалоги: %v", err)
	}

	protoConvs := make([]*msgv1.Conversation, len(conversations))
	for i, c := range conversations {
		protoConvs[i] = conversationToProto(&c)
	}

	return &msgv1.ConversationList{Conversations: protoConvs}, nil
}

func (s *messagingGRPCServer) CreateConversation(ctx context.Context, req *msgv1.CreateConversationRequest) (*msgv1.Conversation, error) {
	participants := make([]int64, len(req.Participants))
	for i, p := range req.Participants {
		participants[i] = int64(p)
	}

	conv, err := svc.CreateConversation(ctx, req.Title, participants)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось создать диалог: %v", err)
	}

	return conversationToProto(conv), nil
}

func (s *messagingGRPCServer) GetConversation(ctx context.Context, req *msgv1.GetConversationRequest) (*msgv1.Conversation, error) {
	conv, err := svc.GetConversation(ctx, int64(req.Id), 0)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "диалог не найден: %v", err)
	}
	return conversationToProto(conv), nil
}

func (s *messagingGRPCServer) UpdateConversation(ctx context.Context, req *msgv1.UpdateConversationRequest) (*msgv1.Conversation, error) {
	conv, err := svc.GetConversation(ctx, int64(req.Id), 0)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "диалог не найден: %v", err)
	}

	title := conv.Title
	if req.Title != "" {
		title = req.Title
	}

	updated, err := svc.UpdateConversation(ctx, int64(req.Id), title, conv.Participants)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось обновить диалог: %v", err)
	}

	return conversationToProto(updated), nil
}

func (s *messagingGRPCServer) DeleteConversation(ctx context.Context, req *msgv1.DeleteConversationRequest) (*msgv1.DeleteConversationResponse, error) {
	if err := svc.DeleteConversation(ctx, int64(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить диалог: %v", err)
	}
	return &msgv1.DeleteConversationResponse{}, nil
}

func (s *messagingGRPCServer) RemoveParticipant(ctx context.Context, req *msgv1.RemoveParticipantRequest) (*msgv1.RemoveParticipantResponse, error) {
	conv, err := svc.GetConversation(ctx, int64(req.ConversationId), 0)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "диалог не найден: %v", err)
	}

	newParticipants := []int64{}
	for _, p := range conv.Participants {
		if p != int64(req.UserId) {
			newParticipants = append(newParticipants, p)
		}
	}

	_, err = svc.UpdateConversation(ctx, int64(req.ConversationId), conv.Title, newParticipants)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить участника: %v", err)
	}

	return &msgv1.RemoveParticipantResponse{}, nil
}

// ─── Messages ───────────────────────────────────────────────────────────────

func (s *messagingGRPCServer) GetMessages(ctx context.Context, req *msgv1.GetMessagesRequest) (*msgv1.MessageList, error) {
	messages, err := svc.GetMessages(ctx, int64(req.ConversationId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить сообщения: %v", err)
	}

	protoMsgs := make([]*msgv1.Message, len(messages))
	for i, m := range messages {
		protoMsgs[i] = messageToProto(&m)
	}

	return &msgv1.MessageList{Messages: protoMsgs}, nil
}

func (s *messagingGRPCServer) SendMessage(ctx context.Context, req *msgv1.SendMessageRequest) (*msgv1.Message, error) {
	msg, err := svc.SendMessage(ctx, int64(req.ConversationId), int64(req.SenderId), req.Content, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отправить сообщение: %v", err)
	}

	return messageToProto(msg), nil
}

func (s *messagingGRPCServer) SendVoiceMessage(stream msgv1.MessagingService_SendVoiceMessageServer) error {
	var conversationID int64
	var senderID int64
	var fileData []byte

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "ошибка чтения stream: %v", err)
		}
		conversationID = int64(chunk.ConversationId)
		senderID = int64(chunk.SenderId)
		fileData = append(fileData, chunk.Data...)
	}

	if conversationID == 0 || senderID == 0 {
		return status.Errorf(codes.InvalidArgument, "conversation_id и sender_id обязательны")
	}

	os.MkdirAll("./uploads/voice", 0755)
	filename := fmt.Sprintf("voice_%d_%d.ogg", conversationID, time.Now().Unix())
	filePath := filepath.Join("./uploads/voice", filename)
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return status.Errorf(codes.Internal, "не удалось сохранить голосовое сообщение: %v", err)
	}

	voiceURL := "/uploads/voice/" + filename

	msg, err := svc.SendVoiceMessage(stream.Context(), conversationID, senderID, voiceURL)
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось отправить голосовое сообщение: %v", err)
	}

	return stream.SendAndClose(messageToProto(msg))
}

func (s *messagingGRPCServer) DeleteMessage(ctx context.Context, req *msgv1.DeleteMessageRequest) (*msgv1.DeleteMessageResponse, error) {
	if err := svc.DeleteMessage(ctx, int64(req.Id), 0); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить сообщение: %v", err)
	}
	return &msgv1.DeleteMessageResponse{}, nil
}

func (s *messagingGRPCServer) MarkMessageRead(ctx context.Context, req *msgv1.MarkMessageReadRequest) (*msgv1.MarkMessageReadResponse, error) {
	if err := svc.MarkMessageRead(ctx, int64(req.MessageId), int64(req.UserId)); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отметить сообщение прочитанным: %v", err)
	}
	return &msgv1.MarkMessageReadResponse{}, nil
}

// ─── Reactions ──────────────────────────────────────────────────────────────

func (s *messagingGRPCServer) AddReaction(ctx context.Context, req *msgv1.AddReactionRequest) (*msgv1.Reaction, error) {
	reaction, err := svc.AddReaction(ctx, int64(req.MessageId), int64(req.UserId), req.Emoji)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось добавить реакцию: %v", err)
	}

	return &msgv1.Reaction{
		Id:        int32(reaction.ID),
		MessageId: int32(reaction.MessageID),
		UserId:    int32(reaction.UserID),
		Emoji:     reaction.Emoji,
	}, nil
}

func (s *messagingGRPCServer) RemoveReaction(ctx context.Context, req *msgv1.RemoveReactionRequest) (*msgv1.RemoveReactionResponse, error) {
	if err := svc.RemoveReaction(ctx, int64(req.ReactionId), 0); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить реакцию: %v", err)
	}
	return &msgv1.RemoveReactionResponse{}, nil
}

// ─── Notifications ──────────────────────────────────────────────────────────

func (s *messagingGRPCServer) GetNotifications(ctx context.Context, req *msgv1.GetNotificationsRequest) (*msgv1.NotificationList, error) {
	notifications, err := svc.GetNotifications(ctx, int64(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить уведомления: %v", err)
	}

	protoNotifs := make([]*msgv1.Notification, len(notifications))
	for i, n := range notifications {
		protoNotifs[i] = &msgv1.Notification{
			Id:        int32(n.ID),
			UserId:    int32(n.UserID),
			Type:      n.Type,
			Message:   n.Message,
			Read:      n.Read,
			CreatedAt: timestamppb.New(n.CreatedAt),
			RelatedId: int32(n.RelatedID),
		}
	}

	return &msgv1.NotificationList{Notifications: protoNotifs}, nil
}

func (s *messagingGRPCServer) MarkNotificationRead(ctx context.Context, req *msgv1.MarkNotificationReadRequest) (*msgv1.MarkNotificationReadResponse, error) {
	if err := svc.MarkNotificationRead(ctx, int64(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отметить уведомление: %v", err)
	}
	return &msgv1.MarkNotificationReadResponse{}, nil
}

func (s *messagingGRPCServer) MarkAllNotificationsRead(ctx context.Context, req *msgv1.MarkAllNotificationsReadRequest) (*msgv1.MarkAllNotificationsReadResponse, error) {
	if err := svc.MarkAllNotificationsRead(ctx, int64(req.UserId)); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отметить все уведомления: %v", err)
	}
	return &msgv1.MarkAllNotificationsReadResponse{}, nil
}

// ─── User Status ────────────────────────────────────────────────────────────

func (s *messagingGRPCServer) GetUserStatuses(ctx context.Context, req *msgv1.GetUserStatusesRequest) (*msgv1.UserStatusList, error) {
	userIDs := make([]int64, len(req.UserIds))
	for i, id := range req.UserIds {
		userIDs[i] = int64(id)
	}

	statuses, err := svc.GetUserStatuses(ctx, userIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить статусы: %v", err)
	}

	protoStatuses := make([]*msgv1.UserStatus, len(statuses))
	for i, us := range statuses {
		protoStatuses[i] = &msgv1.UserStatus{
			Id:       int32(us.ID),
			UserId:   int32(us.UserID),
			IsOnline: us.IsOnline,
			LastSeen: timestamppb.New(us.LastSeen),
		}
	}

	return &msgv1.UserStatusList{Statuses: protoStatuses}, nil
}

func (s *messagingGRPCServer) UpdateUserStatus(ctx context.Context, req *msgv1.UpdateUserStatusRequest) (*msgv1.UserStatus, error) {
	if err := svc.UpdateUserStatus(ctx, int64(req.UserId), req.IsOnline); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось обновить статус: %v", err)
	}

	us, err := svc.GetUserStatus(ctx, int64(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить статус: %v", err)
	}

	return &msgv1.UserStatus{
		Id:       int32(us.ID),
		UserId:   int32(us.UserID),
		IsOnline: us.IsOnline,
		LastSeen: timestamppb.New(us.LastSeen),
	}, nil
}

// ─── Search ─────────────────────────────────────────────────────────────────

func (s *messagingGRPCServer) SearchMessages(ctx context.Context, req *msgv1.SearchMessagesRequest) (*msgv1.MessageList, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}

	messages, err := svc.SearchMessages(ctx, int64(req.UserId), req.Query, limit, int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось выполнить поиск: %v", err)
	}

	protoMsgs := make([]*msgv1.Message, len(messages))
	for i, m := range messages {
		protoMsgs[i] = messageToProto(&m)
	}

	return &msgv1.MessageList{Messages: protoMsgs}, nil
}

// ─── Drom Integration ───────────────────────────────────────────────────────

func (s *messagingGRPCServer) GetDromDialogs(ctx context.Context, req *msgv1.GetDromDialogsRequest) (*msgv1.DromDialogList, error) {
	briefs, err := svc.GetDromBriefs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить диалоги Drom: %v", err)
	}

	protoDialogs := make([]*msgv1.DromDialog, len(briefs))
	for i, d := range briefs {
		protoDialogs[i] = &msgv1.DromDialog{
			Id:           int32(d.DialogID),
			DialogId:     strconv.Itoa(d.DialogID),
			Interlocutor: d.Interlocutor,
			IsUnread:     d.IsUnread,
		}
	}

	return &msgv1.DromDialogList{Dialogs: protoDialogs}, nil
}

func (s *messagingGRPCServer) GetDromMessages(ctx context.Context, req *msgv1.GetDromMessagesRequest) (*msgv1.DromMessageList, error) {
	messages, err := svc.GetDromMessages(ctx, req.DialogId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить сообщения Drom: %v", err)
	}

	protoMsgs := make([]*msgv1.DromMessage, len(messages))
	for i, m := range messages {
		protoMsgs[i] = &msgv1.DromMessage{
			Id:        int32(m.ID),
			DialogId:  m.DialogID,
			MessageId: m.MessageID,
			Author:    m.Author,
			Direction: m.Direction,
			Time:      m.Time,
			Text:      m.Text,
			IsRead:    m.IsRead,
			CreatedAt: m.CreatedAt,
		}
	}

	return &msgv1.DromMessageList{Messages: protoMsgs}, nil
}

func (s *messagingGRPCServer) SendDromMessage(ctx context.Context, req *msgv1.SendDromMessageRequest) (*msgv1.SendDromMessageResponse, error) {
	_, err := svc.SendDromMessage(ctx, req.DialogId, req.Content)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отправить сообщение в Drom: %v", err)
	}

	return &msgv1.SendDromMessageResponse{}, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func conversationToProto(c *Conversation) *msgv1.Conversation {
	participants := make([]int64, len(c.Participants))
	for i, p := range c.Participants {
		participants[i] = p
	}

	return &msgv1.Conversation{
		Id:            int32(c.ID),
		Participants:  participants,
		Title:         c.Title,
		CreatedAt:     timestamppb.New(c.CreatedAt),
		LastMessageAt: timestamppb.New(c.LastMessageAt),
		LastMessage:   c.LastMessage,
		UnreadCount:   int32(c.UnreadCount),
	}
}

func messageToProto(m *Message) *msgv1.Message {
	readBy := make([]int64, len(m.ReadBy))
	for i, r := range m.ReadBy {
		readBy[i] = r
	}

	protoMsg := &msgv1.Message{
		Id:             int32(m.ID),
		ConversationId: int32(m.ConversationID),
		SenderId:       int32(m.SenderID),
		Content:        m.Content,
		MessageType:    m.MessageType,
		VoiceUrl:       m.VoiceURL,
		CreatedAt:      timestamppb.New(m.CreatedAt),
		ReadBy:         readBy,
		Attachments:    m.Attachments,
	}

	for _, mention := range m.Mentions {
		protoMsg.Mentions = append(protoMsg.Mentions, &msgv1.Mention{
			Id:         int32(mention.ID),
			MessageId:  int32(mention.MessageID),
			Type:       mention.Type,
			ResourceId: int32(mention.ResourceID),
			Text:       mention.Text,
		})
	}

	for _, reaction := range m.Reactions {
		protoMsg.Reactions = append(protoMsg.Reactions, &msgv1.Reaction{
			Id:        int32(reaction.ID),
			MessageId: int32(reaction.MessageID),
			UserId:    int32(reaction.UserID),
			Emoji:     reaction.Emoji,
		})
	}

	return protoMsg
}

// ─── gRPC Server Startup ────────────────────────────────────────────────────

// StartGRPCServer запускает gRPC-сервер на указанном порту
func StartGRPCServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingUnaryInterceptor,
			recoveryUnaryInterceptor,
		),
	)

	msgv1.RegisterMessagingServiceServer(srv, NewMessagingGRPCServer())

	// Health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("messaging.v1.MessagingService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	logrus.WithField("port", port).Info("gRPC server listening")
	return srv.Serve(lis)
}

// ─── Interceptors ───────────────────────────────────────────────────────────

func loggingUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	fields := logrus.Fields{
		"method":   info.FullMethod,
		"duration": duration.String(),
	}
	if err != nil {
		fields["error"] = err.Error()
		logrus.WithFields(fields).Warn("gRPC call failed")
	} else if duration > 100*time.Millisecond {
		logrus.WithFields(fields).Info("gRPC call slow")
	}

	return resp, err
}

func recoveryUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).WithField("method", info.FullMethod).Error("gRPC panic recovered")
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()
	return handler(ctx, req)
}
