package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	msgv1 "avtoplaneta/gen/messaging/v1"
	"avtoplaneta/pkg/grpcserver"
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

func (s *messagingGRPCServer) GetConversations(ctx context.Context, req *msgv1.GetConversationsRequest) (*msgv1.GetConversationsResponse, error) {
	conversations, err := svc.GetConversations(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить диалоги: %v", err)
	}

	protoConvs := make([]*msgv1.Conversation, len(conversations))
	for i := range conversations {
		protoConvs[i] = conversations[i].ToProto()
	}

	return &msgv1.GetConversationsResponse{Conversations: protoConvs}, nil
}

func (s *messagingGRPCServer) CreateConversation(ctx context.Context, req *msgv1.CreateConversationRequest) (*msgv1.CreateConversationResponse, error) {
	conv, err := svc.CreateConversation(ctx, req.GetTitle(), req.GetParticipants())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось создать диалог: %v", err)
	}

	return &msgv1.CreateConversationResponse{Conversation: conv.ToProto()}, nil
}

func (s *messagingGRPCServer) GetConversation(ctx context.Context, req *msgv1.GetConversationRequest) (*msgv1.GetConversationResponse, error) {
	conv, err := svc.GetConversation(ctx, req.GetId(), 0)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "диалог не найден: %v", err)
	}
	return &msgv1.GetConversationResponse{Conversation: conv.ToProto()}, nil
}

func (s *messagingGRPCServer) UpdateConversation(ctx context.Context, req *msgv1.UpdateConversationRequest) (*msgv1.UpdateConversationResponse, error) {
	conv, err := svc.GetConversation(ctx, req.GetId(), 0)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "диалог не найден: %v", err)
	}

	title := conv.Title
	if req.GetTitle() != "" {
		title = req.GetTitle()
	}

	updated, err := svc.UpdateConversation(ctx, req.GetId(), title, conv.Participants)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось обновить диалог: %v", err)
	}

	return &msgv1.UpdateConversationResponse{Conversation: updated.ToProto()}, nil
}

func (s *messagingGRPCServer) DeleteConversation(ctx context.Context, req *msgv1.DeleteConversationRequest) (*msgv1.DeleteConversationResponse, error) {
	if err := svc.DeleteConversation(ctx, req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить диалог: %v", err)
	}
	return &msgv1.DeleteConversationResponse{}, nil
}

func (s *messagingGRPCServer) RemoveParticipant(ctx context.Context, req *msgv1.RemoveParticipantRequest) (*msgv1.RemoveParticipantResponse, error) {
	conv, err := svc.GetConversation(ctx, req.GetConversationId(), 0)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "диалог не найден: %v", err)
	}

	newParticipants := []int64{}
	for _, p := range conv.Participants {
		if p != req.GetUserId() {
			newParticipants = append(newParticipants, p)
		}
	}

	_, err = svc.UpdateConversation(ctx, req.GetConversationId(), conv.Title, newParticipants)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить участника: %v", err)
	}

	return &msgv1.RemoveParticipantResponse{}, nil
}

// ─── Messages ───────────────────────────────────────────────────────────────

func (s *messagingGRPCServer) GetMessages(ctx context.Context, req *msgv1.GetMessagesRequest) (*msgv1.GetMessagesResponse, error) {
	messages, err := svc.GetMessages(ctx, req.GetConversationId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить сообщения: %v", err)
	}

	protoMsgs := make([]*msgv1.Message, len(messages))
	for i := range messages {
		protoMsgs[i] = messages[i].ToProto()
	}

	return &msgv1.GetMessagesResponse{Messages: protoMsgs}, nil
}

func (s *messagingGRPCServer) SendMessage(ctx context.Context, req *msgv1.SendMessageRequest) (*msgv1.SendMessageResponse, error) {
	msg, err := svc.SendMessage(ctx, req.GetConversationId(), req.GetSenderId(), req.GetContent(), nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отправить сообщение: %v", err)
	}

	return &msgv1.SendMessageResponse{Message: msg.ToProto()}, nil
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
		conversationID = chunk.GetConversationId()
		senderID = chunk.GetSenderId()
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

	return stream.SendAndClose(&msgv1.SendVoiceMessageResponse{Message: msg.ToProto()})
}

func (s *messagingGRPCServer) DeleteMessage(ctx context.Context, req *msgv1.DeleteMessageRequest) (*msgv1.DeleteMessageResponse, error) {
	if err := svc.DeleteMessage(ctx, req.GetId(), 0); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить сообщение: %v", err)
	}
	return &msgv1.DeleteMessageResponse{}, nil
}

func (s *messagingGRPCServer) MarkMessageRead(ctx context.Context, req *msgv1.MarkMessageReadRequest) (*msgv1.MarkMessageReadResponse, error) {
	if err := svc.MarkMessageRead(ctx, req.GetMessageId(), req.GetUserId()); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отметить сообщение прочитанным: %v", err)
	}
	return &msgv1.MarkMessageReadResponse{}, nil
}

// ─── Reactions ──────────────────────────────────────────────────────────────

func (s *messagingGRPCServer) AddReaction(ctx context.Context, req *msgv1.AddReactionRequest) (*msgv1.AddReactionResponse, error) {
	reaction, err := svc.AddReaction(ctx, req.GetMessageId(), req.GetUserId(), req.GetEmoji())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось добавить реакцию: %v", err)
	}

	return &msgv1.AddReactionResponse{Reaction: reaction.ToProto()}, nil
}

func (s *messagingGRPCServer) RemoveReaction(ctx context.Context, req *msgv1.RemoveReactionRequest) (*msgv1.RemoveReactionResponse, error) {
	if err := svc.RemoveReaction(ctx, req.GetReactionId(), 0); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить реакцию: %v", err)
	}
	return &msgv1.RemoveReactionResponse{}, nil
}

// ─── Notifications ──────────────────────────────────────────────────────────

func (s *messagingGRPCServer) GetNotifications(ctx context.Context, req *msgv1.GetNotificationsRequest) (*msgv1.GetNotificationsResponse, error) {
	notifications, err := svc.GetNotifications(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить уведомления: %v", err)
	}

	protoNotifs := make([]*msgv1.Notification, len(notifications))
	for i := range notifications {
		protoNotifs[i] = notifications[i].ToProto()
	}

	return &msgv1.GetNotificationsResponse{Notifications: protoNotifs}, nil
}

func (s *messagingGRPCServer) MarkNotificationRead(ctx context.Context, req *msgv1.MarkNotificationReadRequest) (*msgv1.MarkNotificationReadResponse, error) {
	if err := svc.MarkNotificationRead(ctx, req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отметить уведомление: %v", err)
	}
	return &msgv1.MarkNotificationReadResponse{}, nil
}

func (s *messagingGRPCServer) MarkAllNotificationsRead(ctx context.Context, req *msgv1.MarkAllNotificationsReadRequest) (*msgv1.MarkAllNotificationsReadResponse, error) {
	if err := svc.MarkAllNotificationsRead(ctx, req.GetUserId()); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отметить все уведомления: %v", err)
	}
	return &msgv1.MarkAllNotificationsReadResponse{}, nil
}

// ─── User Status ────────────────────────────────────────────────────────────

func (s *messagingGRPCServer) GetUserStatuses(ctx context.Context, req *msgv1.GetUserStatusesRequest) (*msgv1.GetUserStatusesResponse, error) {
	statuses, err := svc.GetUserStatuses(ctx, req.GetUserIds())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить статусы: %v", err)
	}

	protoStatuses := make([]*msgv1.UserStatus, len(statuses))
	for i := range statuses {
		protoStatuses[i] = statuses[i].ToProto()
	}

	return &msgv1.GetUserStatusesResponse{Statuses: protoStatuses}, nil
}

func (s *messagingGRPCServer) UpdateUserStatus(ctx context.Context, req *msgv1.UpdateUserStatusRequest) (*msgv1.UpdateUserStatusResponse, error) {
	if err := svc.UpdateUserStatus(ctx, req.GetUserId(), req.GetIsOnline()); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось обновить статус: %v", err)
	}

	us, err := svc.GetUserStatus(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить статус: %v", err)
	}

	return &msgv1.UpdateUserStatusResponse{Status: us.ToProto()}, nil
}

// ─── Search ─────────────────────────────────────────────────────────────────

func (s *messagingGRPCServer) SearchMessages(ctx context.Context, req *msgv1.SearchMessagesRequest) (*msgv1.SearchMessagesResponse, error) {
	limit := int(req.GetLimit())
	if limit == 0 {
		limit = 50
	}

	messages, err := svc.SearchMessages(ctx, req.GetUserId(), req.GetQuery(), limit, int(req.GetOffset()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось выполнить поиск: %v", err)
	}

	protoMsgs := make([]*msgv1.Message, len(messages))
	for i := range messages {
		protoMsgs[i] = messages[i].ToProto()
	}

	return &msgv1.SearchMessagesResponse{Messages: protoMsgs}, nil
}

// ─── Drom Integration ───────────────────────────────────────────────────────

func (s *messagingGRPCServer) GetDromDialogs(ctx context.Context, req *msgv1.GetDromDialogsRequest) (*msgv1.GetDromDialogsResponse, error) {
	dialogs, err := svc.GetDromBriefs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить диалоги Drom: %v", err)
	}

	protoDialogs := make([]*msgv1.DromDialog, len(dialogs))
	for i, d := range dialogs {
		lastMsg := ""
		if d.LastMessage != nil {
			lastMsg = *d.LastMessage
		}
		protoDialogs[i] = &msgv1.DromDialog{
			Id:            d.ID,
			DialogId:      d.DialogID,
			Interlocutor:  d.Interlocutor,
			LastMessageAt: d.LastMessageAt,
			LastMessage:   lastMsg,
			IsUnread:      d.IsUnread,
			PartTitle:     d.PartTitle,
			PartPrice:     d.PartPrice,
			PartUrl:       d.PartURL,
			PartPhotoUrl:  d.PartPhotoURL,
			BulletinId:    d.BulletinID,
			UserRating:    d.UserRating,
			UserStatus:    d.UserStatus,
		}
	}

	return &msgv1.GetDromDialogsResponse{Dialogs: protoDialogs}, nil
}

func (s *messagingGRPCServer) GetDromMessages(ctx context.Context, req *msgv1.GetDromMessagesRequest) (*msgv1.GetDromMessagesResponse, error) {
	messages, dialogMeta, err := svc.GetDromMessages(ctx, req.GetDialogId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить сообщения Drom: %v", err)
	}

	protoMsgs := make([]*msgv1.DromMessage, len(messages))
	for i, m := range messages {
		protoMsgs[i] = &msgv1.DromMessage{
			Id:        m.ID,
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

	var protoDialog *msgv1.DromDialog
	if dialogMeta != nil {
		lastMsg := ""
		if dialogMeta.LastMessage != nil {
			lastMsg = *dialogMeta.LastMessage
		}
		protoDialog = &msgv1.DromDialog{
			Id:            dialogMeta.ID,
			DialogId:      dialogMeta.DialogID,
			Interlocutor:  dialogMeta.Interlocutor,
			LastMessageAt: dialogMeta.LastMessageAt,
			LastMessage:   lastMsg,
			IsUnread:      dialogMeta.IsUnread,
			PartTitle:     dialogMeta.PartTitle,
			PartPrice:     dialogMeta.PartPrice,
			PartUrl:       dialogMeta.PartURL,
			PartPhotoUrl:  dialogMeta.PartPhotoURL,
			BulletinId:    dialogMeta.BulletinID,
			UserRating:    dialogMeta.UserRating,
			UserStatus:    dialogMeta.UserStatus,
		}
	}

	return &msgv1.GetDromMessagesResponse{
		Messages: protoMsgs,
		Dialog:   protoDialog,
	}, nil
}

func (s *messagingGRPCServer) SendDromMessage(ctx context.Context, req *msgv1.SendDromMessageRequest) (*msgv1.SendDromMessageResponse, error) {
	_, err := svc.SendDromMessage(ctx, req.GetDialogId(), req.GetContent())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отправить сообщение в Drom: %v", err)
	}

	return &msgv1.SendDromMessageResponse{Success: true}, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func conversationToProto(c *Conversation) *msgv1.Conversation {
	return c.ToProto()
}

func messageToProto(m *Message) *msgv1.Message {
	return m.ToProto()
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
			grpcserver.ValidationUnaryInterceptor(),
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
