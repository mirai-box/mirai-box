//go:build integration
// +build integration

package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/config"
	"github.com/mirai-box/mirai-box/internal/database"
	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
	"github.com/mirai-box/mirai-box/internal/service"
)

const (
	testUsername = "testuser"
	testPassword = "testpass"
)

func setupTestEnvironment(t *testing.T) (*gorm.DB, *config.Config, func()) {
	// Set up test environment variables
	os.Setenv("APP_ENV", "test")
	os.Setenv("SESSION_KEY", "test-session-key")
	os.Setenv("SECRET_KEY", "123f4ad10c4d8fd1678f89a3b586b4153d518ad59f3c9b08e5967ab50175cb82")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_POSTGRES_PASSWORD", "postgrespass")

	// Start PostgreSQL container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Update DB_HOST and DB_PORT environment variables
	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")
	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port.Port())

	// Load configuration
	conf, err := config.GetApplicationConfig()
	conf.StorageRoot = os.TempDir()
	require.NoError(t, err)

	// Connect to the database
	dbConnectionString := conf.Database.ConnectionString()
	db, err := gorm.Open(postgres.Open(dbConnectionString), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = database.RunMigrations(db)
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		pgContainer.Terminate(ctx)
	}

	return db, conf, cleanup
}

func loginTestUser(t *testing.T, router http.Handler, user *models.User) *http.Cookie {
	t.Helper()
	loginReqBody := map[string]interface{}{
		"username":     user.Username,
		"password":     testPassword,
		"keepSignedIn": true,
	}
	jsonBody, err := json.Marshal(loginReqBody)
	assert.NoError(t, err)

	loginReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)

	require.Equal(t, http.StatusOK, loginResp.Code, "login status is OK")

	var userResp models.UserResponse
	err = json.Unmarshal(loginResp.Body.Bytes(), &userResp)
	require.NoError(t, err)

	assert.NotEmpty(t, userResp.ID)
	assert.Equal(t, user.Username, userResp.Username)

	var sessionCookie *http.Cookie
	for _, cookie := range loginResp.Result().Cookies() {
		if cookie.Name == models.SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.True(t, sessionCookie.MaxAge > 0)

	return sessionCookie
}

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	t.Helper()

	username := fmt.Sprintf("test_%d", rand.Intn(1000))
	userRepo := repos.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	user, err := userService.CreateUser(
		context.Background(),
		username,
		testPassword,
		"user",
	)
	require.NoError(t, err)

	return user
}

func createTestUserRest(t *testing.T, router http.Handler, db *gorm.DB) *models.User {
	t.Helper()

	username := fmt.Sprintf("test_%d", rand.Intn(1000))
	userRepo := repos.NewUserRepository(db)
	stashRepo := repos.NewStashRepository(db)

	reqBody := map[string]interface{}{
		"username": username,
		"password": testPassword,
		"role":     "user",
	}
	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	var userResp models.UserResponse
	err = json.Unmarshal(resp.Body.Bytes(), &userResp)
	require.NoError(t, err)

	assert.NotEmpty(t, userResp.ID)
	assert.Equal(t, username, userResp.Username)
	assert.Equal(t, "user", userResp.Role)

	// Check database using user repo
	dbUser, err := userRepo.FindByUsername(context.Background(), username)
	assert.NoError(t, err)
	assert.Equal(t, userResp.ID, dbUser.ID)
	assert.Equal(t, userResp.Username, dbUser.Username)
	assert.Equal(t, userResp.Role, dbUser.Role)

	// Check stash creation
	stash, err := stashRepo.FindByUserID(context.Background(), dbUser.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, dbUser.ID.String(), stash.UserID.String())

	return dbUser
}
