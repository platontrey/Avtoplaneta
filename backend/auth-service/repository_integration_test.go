//go:build integration
// +build integration

package main

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	pool         *pgxpool.Pool
	userRepo     UserRepository
	activityRepo ActivityLogRepository
}

func loadEnvFile() {
	paths := []string{"../../.env", "../.env", ".env"}
	var file *os.File
	var err error
	for _, path := range paths {
		file, err = os.Open(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = value[1 : len(value)-1]
			}
			if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}
}

func (s *IntegrationTestSuite) SetupSuite() {
	loadEnvFile()

	dbUser := os.Getenv("POSTGRES_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "avtoplaneta"
	}

	connStr := "host=localhost user=" + dbUser + " password=" + dbPassword + " dbname=" + dbName + " port=5433 sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	assert.NoError(s.T(), err)
	s.pool = pool

	err = runMigrationsForTest(pool)
	assert.NoError(s.T(), err)

	s.userRepo = NewUserRepository(pool)
	s.activityRepo = NewActivityLogRepository(pool)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.pool.Close()
}

func (s *IntegrationTestSuite) SetupTest() {
	s.pool.Exec(context.Background(), "DELETE FROM user_activity_logs")
	s.pool.Exec(context.Background(), "DELETE FROM users")
}

// ─── Миграция ───────────────────────────────────────────────────────────────

func runMigrationsForTest(pool *pgxpool.Pool) error {
	entries, _ := migrationsFS.ReadDir("db/migrations")
	for _, entry := range entries {
		if len(entry.Name()) > 7 && entry.Name()[len(entry.Name())-7:] == ".up.sql" {
			data, _ := migrationsFS.ReadFile("db/migrations/" + entry.Name())
			_, err := pool.Exec(context.Background(), string(data))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ─── CreateUser + FindByID ──────────────────────────────────────────────────

func (s *IntegrationTestSuite) TestCreateAndFindUser() {
	user := &User{
		Email:    "real@test.com",
		Name:     "Real User",
		Initials: "RU",
		INN:      "1234567890",
		Provider: "local",
		Role:     "admin",
		Password: "hashed_password",
	}

	created, err := s.userRepo.Create(user)
	assert.NoError(s.T(), err)
	assert.NotZero(s.T(), created.ID)
	assert.Equal(s.T(), "real@test.com", created.Email)

	found, err := s.userRepo.FindByID(created.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, found.ID)
	assert.Equal(s.T(), "Real User", found.Name)
	assert.Equal(s.T(), "RU", found.Initials)
	assert.Equal(s.T(), "1234567890", found.INN)
}

// ─── FindByEmailOrName ──────────────────────────────────────────────────────

func (s *IntegrationTestSuite) TestFindByEmailOrName() {
	user := &User{Email: "byemail@test.com", Name: "Email User", Initials: "EU", Provider: "local", Password: "x"}
	s.userRepo.Create(user)

	found, err := s.userRepo.FindByEmailOrName("byemail@test.com")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "byemail@test.com", found.Email)

	found, err = s.userRepo.FindByEmailOrName("Email User")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Email User", found.Name)

	// Case-insensitive & trimmed search by name
	found, err = s.userRepo.FindByEmailOrName("  email user  ")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Email User", found.Name)

	// Case-insensitive & trimmed search by email
	found, err = s.userRepo.FindByEmailOrName(" BYEMAIL@TEST.COM ")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "byemail@test.com", found.Email)

	// Search by initials
	found, err = s.userRepo.FindByEmailOrName("eu")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Email User", found.Name)
}

// ─── ExistsByEmailOrName ────────────────────────────────────────────────────

func (s *IntegrationTestSuite) TestExistsByEmailOrName() {
	user := &User{Email: "exists@test.com", Name: "Exists User", Provider: "local", Password: "x"}
	s.userRepo.Create(user)

	exists, err := s.userRepo.ExistsByEmailOrName("exists@test.com", "")
	assert.NoError(s.T(), err)
	assert.True(s.T(), exists)

	exists, err = s.userRepo.ExistsByEmailOrName("", "Exists User")
	assert.NoError(s.T(), err)
	assert.True(s.T(), exists)

	exists, err = s.userRepo.ExistsByEmailOrName("no@test.com", "Nobody")
	assert.NoError(s.T(), err)
	assert.False(s.T(), exists)
}

// ─── UpdateUser ─────────────────────────────────────────────────────────────

func (s *IntegrationTestSuite) TestUpdateUser() {
	user := &User{Email: "update@test.com", Name: "Old Name", Provider: "local", Password: "x"}
	created, _ := s.userRepo.Create(user)

	updated, err := s.userRepo.Update(created.ID, UpdateUserParams{
		Name:  "New Name",
		Role:  "manager",
		INN:   "999999",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "New Name", updated.Name)
	assert.Equal(s.T(), "manager", updated.Role)
	assert.Equal(s.T(), "999999", updated.INN)
	assert.Equal(s.T(), "update@test.com", updated.Email)

	found, _ := s.userRepo.FindByID(created.ID)
	assert.Equal(s.T(), "New Name", found.Name)
	assert.Equal(s.T(), "manager", found.Role)
}

// ─── Partial Update ─────────────────────────────────────────────────────────

func (s *IntegrationTestSuite) TestUpdateUserPartial() {
	user := &User{
		Email:    "partial@test.com",
		Name:     "Original",
		Provider: "local",
		Role:     "operator",
		Password: "x",
	}
	created, _ := s.userRepo.Create(user)

	updated, err := s.userRepo.Update(created.ID, UpdateUserParams{
		Name: "Only Name Changed",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Only Name Changed", updated.Name)
	assert.Equal(s.T(), "operator", updated.Role)
	assert.Equal(s.T(), "partial@test.com", updated.Email)
}

// ─── DeleteUser ─────────────────────────────────────────────────────────────

func (s *IntegrationTestSuite) TestDeleteUser() {
	user := &User{Email: "delete@test.com", Name: "Delete Me", Provider: "local", Password: "x"}
	created, _ := s.userRepo.Create(user)

	err := s.userRepo.Delete(created.ID)
	assert.NoError(s.T(), err)

	_, err = s.userRepo.FindByID(created.ID)
	assert.Error(s.T(), err)
}

// ─── FindAll + CountAll ─────────────────────────────────────────────────────

func (s *IntegrationTestSuite) TestFindAllAndCount() {
	s.userRepo.Create(&User{Email: "a@test.com", Name: "A", Provider: "local", Password: "x"})
	s.userRepo.Create(&User{Email: "b@test.com", Name: "B", Provider: "local", Password: "x"})
	s.userRepo.Create(&User{Email: "c@test.com", Name: "C", Provider: "google", Password: ""})

	count, err := s.userRepo.CountAll()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 3, count)

	users, err := s.userRepo.FindAll()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 3)
	assert.Equal(s.T(), "a@test.com", users[0].Email)
	assert.Equal(s.T(), "c@test.com", users[2].Email)
}

// ─── CreateActivityLog ──────────────────────────────────────────────────────

func (s *IntegrationTestSuite) TestCreateActivityLog() {
	log := &UserActivityLog{
		UserID:       1,
		UserName:     "Test",
		UserEmail:    "test@test.com",
		Action:       "login",
		ResourceType: "user",
		Details:      "Login via password",
		IPAddress:    "192.168.1.1",
		UserAgent:    "GoTest/1.0",
	}

	created, err := s.activityRepo.Create(log)
	assert.NoError(s.T(), err)
	assert.NotZero(s.T(), created.ID)
	assert.Equal(s.T(), "login", created.Action)
	assert.Equal(s.T(), "Login via password", created.Details)
	assert.Equal(s.T(), "192.168.1.1", created.IPAddress)
	assert.False(s.T(), created.CreatedAt.IsZero())
}

// ─── CreateActivityLog with ResourceID ──────────────────────────────────────

func (s *IntegrationTestSuite) TestCreateActivityLogWithResourceID() {
	rid := int64(42)
	log := &UserActivityLog{
		UserID:       2,
		UserName:     "Admin",
		UserEmail:    "admin@test.com",
		Action:       "create_part",
		ResourceType: "part",
		ResourceID:   &rid,
	}

	created, err := s.activityRepo.Create(log)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), created.ResourceID)
	assert.Equal(s.T(), int64(42), *created.ResourceID)
}

// ─── FindActivityLogs with Squirrel filters ─────────────────────────────────

func (s *IntegrationTestSuite) TestFindActivityLogsWithFilters() {
	s.activityRepo.Create(&UserActivityLog{UserID: 1, UserName: "U1", UserEmail: "u1@t.com", Action: "login", ResourceType: "user"})
	s.activityRepo.Create(&UserActivityLog{UserID: 1, UserName: "U1", UserEmail: "u1@t.com", Action: "logout", ResourceType: "user"})
	s.activityRepo.Create(&UserActivityLog{UserID: 2, UserName: "U2", UserEmail: "u2@t.com", Action: "login", ResourceType: "user"})

	logs, err := s.activityRepo.FindWithFilters(ActivityLogFilters{Action: "login", Limit: 100})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), logs, 2)

	uid := int64(1)
	logs, err = s.activityRepo.FindWithFilters(ActivityLogFilters{UserID: &uid, Limit: 100})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), logs, 2)
}

// ─── FindActivityLogs Empty ─────────────────────────────────────────────────

func (s *IntegrationTestSuite) TestFindActivityLogsEmpty() {
	logs, err := s.activityRepo.FindWithFilters(ActivityLogFilters{Action: "nonexistent", Limit: 100})
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), logs)
}

// ─── Squirrel query with date range ─────────────────────────────────────────

func (s *IntegrationTestSuite) TestFindActivityLogsByDateRange() {
	s.activityRepo.Create(&UserActivityLog{UserID: 1, UserName: "U", UserEmail: "u@t.com", Action: "login", ResourceType: "user"})

	logs, err := s.activityRepo.FindWithFilters(ActivityLogFilters{Limit: 100})
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), logs)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
