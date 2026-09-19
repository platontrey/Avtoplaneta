package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type mockUserEventPublisher struct {
	called  bool
	userID  int64
	newName string
}

func (m *mockUserEventPublisher) PublishUserRenamed(ctx context.Context, userID int64, newName string) error {
	m.called = true
	m.userID = userID
	m.newName = newName
	return nil
}

type ServiceTestSuite struct {
	suite.Suite
	mockUserRepo     *MockUserRepository
	mockActivityRepo *MockActivityLogRepository
	service          AuthService
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.mockUserRepo = new(MockUserRepository)
	suite.mockActivityRepo = new(MockActivityLogRepository)
	suite.service = NewAuthService(suite.mockUserRepo, suite.mockActivityRepo, nil, nil)
}

func (suite *ServiceTestSuite) TearDownTest() {
	suite.mockUserRepo.AssertExpectations(suite.T())
	suite.mockActivityRepo.AssertExpectations(suite.T())
}

func testPasswordHash(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed)
}

// ─── AuthenticateUser ───────────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestAuthenticateUser_Success() {
	password := "correct_password"
	hashed := testPasswordHash(password)

	user := &User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Provider: "local",
		Role:     "operator",
		Password: hashed,
	}

	suite.mockUserRepo.On("FindByEmailOrName", "test@example.com").Return(user, nil)

	result, err := suite.service.AuthenticateUser("test@example.com", password)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(1), result.ID)
}

func (suite *ServiceTestSuite) TestAuthenticateUser_WrongPassword() {
	hashed := testPasswordHash("correct_password")

	user := &User{
		ID:       1,
		Email:    "test@example.com",
		Provider: "local",
		Password: hashed,
	}

	suite.mockUserRepo.On("FindByEmailOrName", "test@example.com").Return(user, nil)

	_, err := suite.service.AuthenticateUser("test@example.com", "wrong_password")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "неверные учетные данные")
}

func (suite *ServiceTestSuite) TestAuthenticateUser_NonLocal() {
	user := &User{
		ID:       1,
		Email:    "test@gmail.com",
		Provider: "google",
		Password: "",
	}

	suite.mockUserRepo.On("FindByEmailOrName", "test@gmail.com").Return(user, nil)

	_, err := suite.service.AuthenticateUser("test@gmail.com", "anything")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "неверные учетные данные")
}

func (suite *ServiceTestSuite) TestAuthenticateUser_NotFound() {
	suite.mockUserRepo.On("FindByEmailOrName", "nonexistent@example.com").Return(nil, assert.AnError)

	_, err := suite.service.AuthenticateUser("nonexistent@example.com", "password")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "неверные учетные данные")
}

// ─── CreateUserFromGoogle ───────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestCreateUserFromGoogle_NewUser() {
	suite.mockUserRepo.On("FindByEmailOrName", "new@gmail.com").Return(nil, assert.AnError)

	expected := &User{
		ID:       2,
		Email:    "new@gmail.com",
		Name:     "New User",
		Provider: "google",
		Role:     "operator",
	}

	call := suite.mockUserRepo.On("Create", mock.MatchedBy(func(u *User) bool {
		return u.Email == "new@gmail.com" && u.Provider == "google" && u.Role == "operator"
	})).Return(expected, nil).Run(func(args mock.Arguments) {
		user := args.Get(0).(*User)
		user.ID = 2
	})

	result, err := suite.service.CreateUserFromGoogle("new@gmail.com", "New User")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(2), result.ID)
	assert.Equal(suite.T(), "google", result.Provider)
	suite.mockUserRepo.AssertCalled(suite.T(), "Create", mock.AnythingOfType("*main.User"))
	_ = call
}

func (suite *ServiceTestSuite) TestCreateUserFromGoogle_Existing() {
	existing := &User{
		ID:       1,
		Email:    "existing@gmail.com",
		Name:     "Existing",
		Provider: "google",
		Role:     "operator",
	}

	suite.mockUserRepo.On("FindByEmailOrName", "existing@gmail.com").Return(existing, nil)

	result, err := suite.service.CreateUserFromGoogle("existing@gmail.com", "Existing")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(1), result.ID)
	suite.mockUserRepo.AssertNotCalled(suite.T(), "Create", mock.Anything)
}

// ─── CreateUser ─────────────────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestCreateUser_Success() {
	suite.mockUserRepo.On("ExistsByEmailOrName", "new@example.com", "New User").Return(false, nil)

	expected := &User{
		ID:       3,
		Email:    "new@example.com",
		Name:     "New User",
		Provider: "local",
		Role:     "operator",
		Password: "",
	}

	suite.mockUserRepo.On("Create", mock.MatchedBy(func(u *User) bool {
		return u.Email == "new@example.com" && u.Provider == "local"
	})).Return(expected, nil).Run(func(args mock.Arguments) {
		user := args.Get(0).(*User)
		user.ID = 3
	})

	req := CreateUserRequest{
		Email:    "new@example.com",
		Name:     "New User",
		Password: "password123",
		Role:     "operator",
	}

	result, err := suite.service.CreateUser(req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(3), result.ID)
	assert.Equal(suite.T(), "", result.Password)
}

func (suite *ServiceTestSuite) TestCreateUser_Duplicate() {
	suite.mockUserRepo.On("ExistsByEmailOrName", "exists@example.com", "Exists").Return(true, nil)

	req := CreateUserRequest{
		Email:    "exists@example.com",
		Name:     "Exists",
		Password: "password",
		Role:     "operator",
	}

	result, err := suite.service.CreateUser(req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "уже существует")
}

// ─── UpdateUser ─────────────────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestUpdateUser_Success() {
	existing := &User{ID: 1, Email: "old@example.com", Name: "Old Name", Role: "operator"}

	updated := &User{ID: 1, Email: "new@example.com", Name: "New Name", Role: "admin"}

	suite.mockUserRepo.On("FindByID", int64(1)).Return(existing, nil)
	suite.mockUserRepo.On("ExistsByEmailOrName", "new@example.com", "").Return(false, nil)

	updateParams := UpdateUserParams{
		Name:  "New Name",
		Email: "new@example.com",
		Role:  "admin",
	}
	suite.mockUserRepo.On("Update", int64(1), updateParams).Return(updated, nil)

	req := UpdateUserRequest{
		Name:  "New Name",
		Email: "new@example.com",
		Role:  "admin",
	}

	result, err := suite.service.UpdateUser(1, req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "New Name", result.Name)
	assert.Equal(suite.T(), "", result.Password)
}

func (suite *ServiceTestSuite) TestUpdateUser_PublishesRenamedEvent() {
	existing := &User{ID: 5, Email: "seller@example.com", Name: "Old Seller", Role: "operator"}
	updated := &User{ID: 5, Email: "seller@example.com", Name: "Renamed Seller", Role: "operator"}

	suite.mockUserRepo.On("FindByID", int64(5)).Return(existing, nil)
	suite.mockUserRepo.On("Update", int64(5), UpdateUserParams{Name: "Renamed Seller"}).Return(updated, nil)

	mockPub := &mockUserEventPublisher{}
	svc := NewAuthService(suite.mockUserRepo, suite.mockActivityRepo, nil, nil, mockPub)

	result, err := svc.UpdateUser(5, UpdateUserRequest{Name: "Renamed Seller"})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Renamed Seller", result.Name)
	assert.True(suite.T(), mockPub.called)
	assert.Equal(suite.T(), int64(5), mockPub.userID)
	assert.Equal(suite.T(), "Renamed Seller", mockPub.newName)
}

func (suite *ServiceTestSuite) TestUpdateUser_NotFound() {
	suite.mockUserRepo.On("FindByID", int64(999)).Return(nil, assert.AnError)

	req := UpdateUserRequest{Name: "Test"}
	result, err := suite.service.UpdateUser(999, req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "не найден")
}

func (suite *ServiceTestSuite) TestUpdateUser_DuplicateEmail() {
	existing := &User{ID: 1, Email: "old@example.com", Name: "Old"}
	suite.mockUserRepo.On("FindByID", int64(1)).Return(existing, nil)
	suite.mockUserRepo.On("ExistsByEmailOrName", "taken@example.com", "").Return(true, nil)

	req := UpdateUserRequest{Email: "taken@example.com"}
	result, err := suite.service.UpdateUser(1, req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "уже существует")
}

func (suite *ServiceTestSuite) TestUpdateUser_NoFields() {
	existing := &User{ID: 1, Email: "test@example.com"}
	suite.mockUserRepo.On("FindByID", int64(1)).Return(existing, nil)

	req := UpdateUserRequest{}
	result, err := suite.service.UpdateUser(1, req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "нет полей для обновления")
}

// ─── DeleteUser ─────────────────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestDeleteUser_Success() {
	existing := &User{ID: 1, Email: "test@example.com"}
	suite.mockUserRepo.On("FindByID", int64(1)).Return(existing, nil)
	suite.mockUserRepo.On("Delete", int64(1)).Return(nil)

	err := suite.service.DeleteUser(1)
	assert.NoError(suite.T(), err)
}

func (suite *ServiceTestSuite) TestDeleteUser_NotFound() {
	suite.mockUserRepo.On("FindByID", int64(999)).Return(nil, assert.AnError)

	err := suite.service.DeleteUser(999)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "не найден")
}

// ─── GetUsers ───────────────────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestGetUsers() {
	users := []User{
		{ID: 1, Email: "user1@example.com", Name: "User 1", Password: "secret1"},
		{ID: 2, Email: "user2@example.com", Name: "User 2", Password: "secret2"},
	}
	suite.mockUserRepo.On("FindAll").Return(users, nil)

	result, err := suite.service.GetUsers()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "", result[0].Password)
	assert.Equal(suite.T(), "", result[1].Password)
}

func (suite *ServiceTestSuite) TestGetUsers_Empty() {
	suite.mockUserRepo.On("FindAll").Return([]User{}, nil)

	result, err := suite.service.GetUsers()
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), result)
}

// ─── GetCurrentUser ─────────────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestGetCurrentUser_Success() {
	user := &User{ID: 1, Email: "test@example.com", Password: "secret"}
	suite.mockUserRepo.On("FindByID", int64(1)).Return(user, nil)

	result, err := suite.service.GetCurrentUser(1)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "", result.Password)
}

func (suite *ServiceTestSuite) TestGetCurrentUser_NotFound() {
	suite.mockUserRepo.On("FindByID", int64(999)).Return(nil, assert.AnError)

	result, err := suite.service.GetCurrentUser(999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// ─── GetTotalUsersCount ─────────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestGetTotalUsersCount() {
	suite.mockUserRepo.On("CountAll").Return(10, nil)

	count, err := suite.service.GetTotalUsersCount()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 10, count)
}

// ─── LogUserActivity ────────────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestLogUserActivity() {
	user := &User{ID: 1, Name: "Test", Email: "test@example.com"}

	suite.mockActivityRepo.On("Create", mock.MatchedBy(func(log *UserActivityLog) bool {
		return log.UserID == 1 && log.Action == "login" && log.ResourceType == "user"
	})).Return(&UserActivityLog{ID: 1, UserID: 1, Action: "login"}, nil)

	err := suite.service.LogUserActivity(user, "login", "user", nil, "details", "127.0.0.1", "Mozilla")
	assert.NoError(suite.T(), err)
}

func (suite *ServiceTestSuite) TestLogUserActivity_WithResourceID() {
	user := &User{ID: 2, Name: "Admin", Email: "admin@example.com"}
	resourceID := int64(42)

	suite.mockActivityRepo.On("Create", mock.MatchedBy(func(log *UserActivityLog) bool {
		return log.ResourceID != nil && *log.ResourceID == 42
	})).Return(&UserActivityLog{ID: 2, UserID: 2, Action: "update_part"}, nil)

	err := suite.service.LogUserActivity(user, "update_part", "part", &resourceID, "Updated", "10.0.0.1", "curl")
	assert.NoError(suite.T(), err)
}

// ─── GetUserActivityLogs ────────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestGetUserActivityLogs() {
	now := time.Now()
	userID := int64(1)
	filters := ActivityLogFilters{
		UserID: &userID,
		Action: "login",
		Limit:  50,
	}

	expected := []UserActivityLog{
		{ID: 1, UserID: 1, Action: "login", CreatedAt: now},
	}

	suite.mockActivityRepo.On("FindWithFilters", filters).Return(expected, nil)

	result, err := suite.service.GetUserActivityLogs(filters)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), "login", result[0].Action)
}

// ─── Logout ─────────────────────────────────────────────────────────────────

func (suite *ServiceTestSuite) TestLogout() {
	err := suite.service.Logout(1)
	assert.NoError(suite.T(), err)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
