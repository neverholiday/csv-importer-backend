package main

import (
	"context"
	"csv-importer-backend/cmd/csv-importer/apis"
	"csv-importer-backend/cmd/csv-importer/model"
	"csv-importer-backend/cmd/csv-importer/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	testDBHost     = "localhost"
	testDBPort     = 5432
	testDBUser     = "postgres"
	testDBPassword = "mypassword"
	testDBName     = "postgres" // Use existing database instead of separate test DB
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Skip integration tests if not in integration test environment
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	// Ensure tables exist (auto-migrate if needed)
	err = db.AutoMigrate(&model.Event{}, &model.TodoEvent{})
	require.NoError(t, err, "Failed to migrate test database")
	
	// Clean up existing test data after tables are ensured to exist
	db.Exec("TRUNCATE TABLE events CASCADE")
	db.Exec("TRUNCATE TABLE todo_events CASCADE")

	return db
}

func teardownTestDB(t *testing.T, db *gorm.DB) {
	// Clean up test data (ignore errors since tables might not exist yet)
	db.Exec("TRUNCATE TABLE events CASCADE")
	db.Exec("TRUNCATE TABLE todo_events CASCADE")
	
	// Close database connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func TestIntegration_EventAPI_CreateAndList(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	eventRepo := repository.NewEventRepo(db)
	_ = apis.NewEventAPI(eventRepo)

	// Test creating an event directly through repository
	// Since the API methods are private, we'll test the integration at the repository level
	testEvent := model.Event{
		ID:         "integration-test-event",
		Name:       "Integration Test Event",
		Status:     model.Created,
		CreateDate: time.Now(),
		UpdateDate: time.Now(),
	}

	err := eventRepo.CreateEvent(context.Background(), testEvent)
	assert.NoError(t, err)

	// Test listing events to verify creation
	events, err := eventRepo.ListEvents(context.Background())
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, testEvent.ID, events[0].ID)
	assert.Equal(t, testEvent.Name, events[0].Name)
	
	// Verify database state
	var eventCount int64
	db.Model(&model.Event{}).Count(&eventCount)
	assert.Equal(t, int64(1), eventCount)
	
	t.Log("Integration test completed successfully - event created and retrieved")
}

func TestIntegration_DatabaseOperations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewEventRepo(db)

	// Test creating an event
	testEvent := model.Event{
		ID:         "integration-test-1",
		Name:       "Integration Test Event",
		Status:     model.Created,
		CreateDate: time.Now(),
		UpdateDate: time.Now(),
	}

	err := repo.CreateEvent(db.Statement.Context, testEvent)
	assert.NoError(t, err)

	// Test listing events
	events, err := repo.ListEvents(db.Statement.Context)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, testEvent.ID, events[0].ID)
	assert.Equal(t, testEvent.Name, events[0].Name)
	assert.Equal(t, testEvent.Status, events[0].Status)

	// Test creating multiple events
	testEvent2 := model.Event{
		ID:         "integration-test-2",
		Name:       "Second Test Event",
		Status:     model.Start,
		CreateDate: time.Now(),
		UpdateDate: time.Now(),
	}

	err = repo.CreateEvent(db.Statement.Context, testEvent2)
	assert.NoError(t, err)

	events, err = repo.ListEvents(db.Statement.Context)
	assert.NoError(t, err)
	assert.Len(t, events, 2)

	// Verify both events exist
	var foundFirst, foundSecond bool
	for _, event := range events {
		if event.ID == testEvent.ID {
			foundFirst = true
			assert.Equal(t, testEvent.Name, event.Name)
		}
		if event.ID == testEvent2.ID {
			foundSecond = true
			assert.Equal(t, testEvent2.Name, event.Name)
		}
	}
	assert.True(t, foundFirst, "First event not found")
	assert.True(t, foundSecond, "Second event not found")
}

func TestIntegration_DatabaseConstraints(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewEventRepo(db)

	// Test unique constraint on ID
	testEvent := model.Event{
		ID:         "duplicate-test",
		Name:       "First Event",
		Status:     model.Created,
		CreateDate: time.Now(),
		UpdateDate: time.Now(),
	}

	err := repo.CreateEvent(db.Statement.Context, testEvent)
	assert.NoError(t, err)

	// Try to create another event with the same ID
	duplicateEvent := model.Event{
		ID:         "duplicate-test", // Same ID
		Name:       "Second Event",
		Status:     model.Start,
		CreateDate: time.Now(),
		UpdateDate: time.Now(),
	}

	err = repo.CreateEvent(db.Statement.Context, duplicateEvent)
	assert.Error(t, err, "Should fail due to duplicate ID")

	// Verify only one event exists
	events, err := repo.ListEvents(db.Statement.Context)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "First Event", events[0].Name)
}

func TestIntegration_DatabaseTransactions(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Test transaction rollback on error
	err := db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewEventRepo(tx)

		// Create first event
		testEvent1 := model.Event{
			ID:         "tx-test-1",
			Name:       "Transaction Test 1",
			Status:     model.Created,
			CreateDate: time.Now(),
			UpdateDate: time.Now(),
		}

		err := repo.CreateEvent(tx.Statement.Context, testEvent1)
		if err != nil {
			return err
		}

		// Create second event that will cause the transaction to rollback
		testEvent2 := model.Event{
			ID:         "tx-test-1", // Duplicate ID to force error
			Name:       "Transaction Test 2",
			Status:     model.Start,
			CreateDate: time.Now(),
			UpdateDate: time.Now(),
		}

		return repo.CreateEvent(tx.Statement.Context, testEvent2)
	})

	assert.Error(t, err, "Transaction should fail due to duplicate ID")

	// Verify no events were created due to rollback
	repo := repository.NewEventRepo(db)
	events, err := repo.ListEvents(db.Statement.Context)
	assert.NoError(t, err)
	assert.Len(t, events, 0, "No events should exist after transaction rollback")
}

func TestIntegration_LargeDataset(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewEventRepo(db)

	// Create 100 events to test performance and stability
	const numEvents = 100
	for i := 0; i < numEvents; i++ {
		testEvent := model.Event{
			ID:         fmt.Sprintf("large-test-%d", i),
			Name:       fmt.Sprintf("Large Dataset Test Event %d", i),
			Status:     model.Created,
			CreateDate: time.Now(),
			UpdateDate: time.Now(),
		}

		err := repo.CreateEvent(db.Statement.Context, testEvent)
		assert.NoError(t, err, "Failed to create event %d", i)
	}

	// Verify all events were created
	events, err := repo.ListEvents(db.Statement.Context)
	assert.NoError(t, err)
	assert.Len(t, events, numEvents)

	// Verify events are returned in some order (database default)
	eventIDs := make(map[string]bool)
	for _, event := range events {
		eventIDs[event.ID] = true
		assert.Contains(t, event.Name, "Large Dataset Test Event")
	}
	assert.Len(t, eventIDs, numEvents, "All events should have unique IDs")
}

// Helper function to create a test server with real database
func createTestServer(t *testing.T) (*echo.Echo, *gorm.DB) {
	db := setupTestDB(t)
	
	e := echo.New()
	rootg := e.Group("")
	v1g := rootg.Group("/api/v1")

	// Setup health check
	apis.NewHealthCheckAPI(db).Setup(rootg)

	// Setup event API
	eventRepo := repository.NewEventRepo(db)
	apis.NewEventAPI(eventRepo).Setup(v1g)

	return e, db
}

func TestIntegration_HealthCheckEndpoint(t *testing.T) {
	server, db := createTestServer(t)
	defer teardownTestDB(t, db)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// Health check should return some indication of system status
	t.Logf("Health check response: %v", response)
}

// Benchmark for database operations
func BenchmarkIntegration_CreateEvent(b *testing.B) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		b.Skip("Skipping integration benchmark. Set INTEGRATION_TEST=1 to run.")
	}

	db := setupTestDB(&testing.T{})
	defer teardownTestDB(&testing.T{}, db)

	repo := repository.NewEventRepo(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testEvent := model.Event{
			ID:         fmt.Sprintf("bench-test-%d", i),
			Name:       fmt.Sprintf("Benchmark Test Event %d", i),
			Status:     model.Created,
			CreateDate: time.Now(),
			UpdateDate: time.Now(),
		}

		err := repo.CreateEvent(db.Statement.Context, testEvent)
		if err != nil {
			b.Fatalf("Failed to create event: %v", err)
		}
	}
}