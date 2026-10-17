package domain

import (
	authv1 "avtoplaneta/gen/auth/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto конвертирует доменную модель User в Protobuf User.
func (u *User) ToProto() *authv1.User {
	if u == nil {
		return nil
	}
	return &authv1.User{
		Id:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		Initials:  u.Initials,
		Inn:       u.INN,
		Provider:  u.Provider,
		AvatarUrl: u.AvatarURL,
	}
}

// ToProto конвертирует доменную модель UserActivityLog в Protobuf ActivityLog.
func (l *UserActivityLog) ToProto() *authv1.ActivityLog {
	if l == nil {
		return nil
	}
	protoLog := &authv1.ActivityLog{
		Id:           l.ID,
		UserId:       l.UserID,
		UserName:     l.UserName,
		UserEmail:    l.UserEmail,
		Action:       l.Action,
		ResourceType: l.ResourceType,
		Details:      l.Details,
		IpAddress:    l.IPAddress,
		UserAgent:    l.UserAgent,
		CreatedAt:    timestamppb.New(l.CreatedAt),
	}
	if l.ResourceID != nil {
		protoLog.ResourceId = *l.ResourceID
	}
	return protoLog
}

// UserFromCreateProto создает DTO CreateUserRequest из Protobuf запроса, используя безопасные геттеры.
func UserFromCreateProto(req *authv1.CreateUserRequest) CreateUserRequest {
	return CreateUserRequest{
		Username:  req.GetUsername(),
		Email:     req.GetEmail(),
		Name:      req.GetName(),
		Password:  req.GetPassword(),
		Role:      req.GetRole(),
		Initials:  req.GetInitials(),
		INN:       req.GetInn(),
		AvatarURL: req.GetAvatarUrl(),
	}
}

// UserFromUpdateProto создает DTO UpdateUserRequest из Protobuf запроса, используя безопасные геттеры.
func UserFromUpdateProto(req *authv1.UpdateUserRequest) UpdateUserRequest {
	return UpdateUserRequest{
		Username:  req.GetUsername(),
		Name:      req.GetName(),
		Email:     req.GetEmail(),
		Initials:  req.GetInitials(),
		INN:       req.GetInn(),
		Role:      req.GetRole(),
		AvatarURL: req.GetAvatarUrl(),
	}
}
