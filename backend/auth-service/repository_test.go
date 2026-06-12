package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ─── UserRepository Mock ────────────────────────────────────────────────────

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *User) (*User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id int64) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByEmailOrName(identifier string) (*User, error) {
	args := m.Called(identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Update(id int64, params UpdateUserParams) (*User, error) {
	args := m.Called(id, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) FindAll() ([]User, error) {
	args := m.Called()
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmailOrName(email, name string) (bool, error) {
	args := m.Called(email, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CountAll() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// ─── ActivityLogRepository Mock ─────────────────────────────────────────────

type MockActivityLogRepository struct {
	mock.Mock
}

func (m *MockActivityLogRepository) Create(log *UserActivityLog) (*UserActivityLog, error) {
	args := m.Called(log)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserActivityLog), args.Error(1)
}

func (m *MockActivityLogRepository) FindWithFilters(filters ActivityLogFilters) ([]UserActivityLog, error) {
	args := m.Called(filters)
	return args.Get(0).([]UserActivityLog), args.Error(1)
}

// ─── Repository Test Suite ──────────────────────────────────────────────────

type RepositoryTestSuite struct {
	suite.Suite
	mockUserRepo     *MockUserRepository
	mockActivityRepo *MockActivityLogRepository
	userRepo         UserRepository
	activityRepo     ActivityLogRepository
}

func (suite *RepositoryTestSuite) SetupTest() {
	suite.mockUserRepo = new(MockUserRepository)
	suite.mockActivityRepo = new(MockActivityLogRepository)
	suite.userRepo = suite.mockUserRepo
	suite.activityRepo = suite.mockActivityRepo
}

func (suite *RepositoryTestSuite) TearDownTest() {
	suite.mockUserRepo.AssertExpectations(suite.T())
	suite.mockActivityRepo.AssertExpectations(suite.T())
}

// TestUserCreate
func (suite *RepositoryTestSuite) TestUserCreate() {
	user := &User{
		Email:    "test@example.com",
		Name:     "Test User",
		Provider: "local",
		Role:     "operator",
		Password: "hashed",
	}

	expected := &User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Provider: "local",
		Role:     "operator",
		Password: "hashed",
	}

	suite.mockUserRepo.On("Create", user).Return(expected, nil)

	result, err := suite.userRepo.Create(user)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(1), result.ID)
	assert.Equal(suite.T(), "test@example.com", result.Email)
}

// TestUserCreate_Error
func (suite *RepositoryTestSuite) TestUserCreate_Error() {
	user := &User{Email: "", Name: ""}
	suite.mockUserRepo.On("Create", user).Return(nil, assert.AnError)

	result, err := suite.userRepo.Create(user)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// TestUserFindByID
func (suite *RepositoryTestSuite) TestUserFindByID() {
	expected := &User{ID: 1, Email: "test@example.com", Name: "Test User"}
	suite.mockUserRepo.On("FindByID", int64(1)).Return(expected, nil)

	result, err := suite.userRepo.FindByID(1)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Test User", result.Name)
}

// TestUserFindByID_NotFound
func (suite *RepositoryTestSuite) TestUserFindByID_NotFound() {
	suite.mockUserRepo.On("FindByID", int64(999)).Return(nil, assert.AnError)

	result, err := suite.userRepo.FindByID(999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// TestUserFindByEmailOrName
func (suite *RepositoryTestSuite) TestUserFindByEmailOrName() {
	expected := &User{ID: 1, Email: "test@example.com", Name: "Test User"}
	suite.mockUserRepo.On("FindByEmailOrName", "test@example.com").Return(expected, nil)

	result, err := suite.userRepo.FindByEmailOrName("test@example.com")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "test@example.com", result.Email)
}

// TestUserUpdate
func (suite *RepositoryTestSuite) TestUserUpdate() {
	params := UpdateUserParams{Name: "Updated", Email: "updated@example.com", Role: "admin"}
	expected := &User{ID: 1, Name: "Updated", Email: "updated@example.com", Role: "admin"}
	suite.mockUserRepo.On("Update", int64(1), params).Return(expected, nil)

	result, err := suite.userRepo.Update(1, params)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Updated", result.Name)
	assert.Equal(suite.T(), "admin", result.Role)
}

// TestUserDelete
func (suite *RepositoryTestSuite) TestUserDelete() {
	suite.mockUserRepo.On("Delete", int64(1)).Return(nil)

	err := suite.userRepo.Delete(1)
	assert.NoError(suite.T(), err)
}

// TestUserFindAll
func (suite *RepositoryTestSuite) TestUserFindAll() {
	expected := []User{
		{ID: 1, Email: "user1@example.com", Name: "User 1"},
		{ID: 2, Email: "user2@example.com", Name: "User 2"},
		{ID: 3, Email: "user3@example.com", Name: "User 3"},
	}
	suite.mockUserRepo.On("FindAll").Return(expected, nil)

	users, err := suite.userRepo.FindAll()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), users, 3)
	assert.Equal(suite.T(), "User 1", users[0].Name)
	assert.Equal(suite.T(), "User 3", users[2].Name)
}

// TestUserFindAll_Empty
func (suite *RepositoryTestSuite) TestUserFindAll_Empty() {
	suite.mockUserRepo.On("FindAll").Return([]User{}, nil)

	users, err := suite.userRepo.FindAll()
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), users)
}

// TestUserExistsByEmailOrName_Exists
func (suite *RepositoryTestSuite) TestUserExistsByEmailOrName_Exists() {
	suite.mockUserRepo.On("ExistsByEmailOrName", "test@example.com", "Test User").Return(true, nil)

	exists, err := suite.userRepo.ExistsByEmailOrName("test@example.com", "Test User")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

// TestUserExistsByEmailOrName_NotExists
func (suite *RepositoryTestSuite) TestUserExistsByEmailOrName_NotExists() {
	suite.mockUserRepo.On("ExistsByEmailOrName", "nonexistent@example.com", "Nobody").Return(false, nil)

	exists, err := suite.userRepo.ExistsByEmailOrName("nonexistent@example.com", "Nobody")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

// TestUserCountAll
func (suite *RepositoryTestSuite) TestUserCountAll() {
	suite.mockUserRepo.On("CountAll").Return(42, nil)

	count, err := suite.userRepo.CountAll()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 42, count)
}

// TestActivityLogCreate
func (suite *RepositoryTestSuite) TestActivityLogCreate() {
	logEntry := &UserActivityLog{
		UserID:       1,
		UserName:     "Test",
		UserEmail:    "test@example.com",
		Action:       "login",
		ResourceType: "user",
	}

	expected := &UserActivityLog{
		ID:           100,
		UserID:       1,
		UserName:     "Test",
		UserEmail:    "test@example.com",
		Action:       "login",
		ResourceType: "user",
		CreatedAt:    time.Now(),
	}

	suite.mockActivityRepo.On("Create", logEntry).Return(expected, nil)

	result, err := suite.activityRepo.Create(logEntry)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(100), result.ID)
	assert.Equal(suite.T(), "login", result.Action)
}

// TestActivityLogFindWithFilters
func (suite *RepositoryTestSuite) TestActivityLogFindWithFilters() {
	now := time.Now()
	userID := int64(1)
	filters := ActivityLogFilters{
		UserID: &userID,
		Action: "login",
		Limit:  50,
		Offset: 0,
	}

	expected := []UserActivityLog{
		{ID: 1, UserID: 1, Action: "login", ResourceType: "user", CreatedAt: now},
		{ID: 2, UserID: 1, Action: "login", ResourceType: "part", CreatedAt: now.Add(-time.Hour)},
	}

	suite.mockActivityRepo.On("FindWithFilters", filters).Return(expected, nil)

	logs, err := suite.activityRepo.FindWithFilters(filters)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), logs, 2)
	assert.Equal(suite.T(), "login", logs[0].Action)
}

// TestActivityLogFindWithFilters_Empty
func (suite *RepositoryTestSuite) TestActivityLogFindWithFilters_Empty() {
	filters := ActivityLogFilters{
		Action: "nonexistent_action",
		Limit:  100,
	}

	suite.mockActivityRepo.On("FindWithFilters", filters).Return([]UserActivityLog{}, nil)

	logs, err := suite.activityRepo.FindWithFilters(filters)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), logs)
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
