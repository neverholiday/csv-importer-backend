package repository

import (
	"context"
	"csv-importer-backend/cmd/csv-importer/model"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("Failed to create GORM instance: %v", err)
	}

	return gormDB, mock
}

func TestEventRepo_ListEvents_Success(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewEventRepo(gormDB)

	expectedTime := time.Now()
	expectedEvents := []model.Event{
		{
			ID:         "event-1",
			Name:       "Test Event 1",
			Status:     model.Created,
			CreateDate: expectedTime,
			UpdateDate: expectedTime,
		},
		{
			ID:         "event-2",
			Name:       "Test Event 2",
			Status:     model.Start,
			CreateDate: expectedTime,
			UpdateDate: expectedTime,
		},
	}

	rows := sqlmock.NewRows([]string{"id", "name", "status", "create_date", "update_date", "delete_date"}).
		AddRow("event-1", "Test Event 1", "draft", expectedTime, expectedTime, nil).
		AddRow("event-2", "Test Event 2", "start", expectedTime, expectedTime, nil)

	mock.ExpectQuery(`SELECT \* FROM "events"`).
		WillReturnRows(rows)

	ctx := context.Background()
	events, err := repo.ListEvents(ctx)

	assert.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, expectedEvents[0].ID, events[0].ID)
	assert.Equal(t, expectedEvents[0].Name, events[0].Name)
	assert.Equal(t, expectedEvents[0].Status, events[0].Status)
	assert.Equal(t, expectedEvents[1].ID, events[1].ID)
	assert.Equal(t, expectedEvents[1].Name, events[1].Name)
	assert.Equal(t, expectedEvents[1].Status, events[1].Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepo_ListEvents_DatabaseError(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewEventRepo(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "events"`).
		WillReturnError(errors.New("database connection failed"))

	ctx := context.Background()
	events, err := repo.ListEvents(ctx)

	assert.Error(t, err)
	assert.Nil(t, events)
	assert.Contains(t, err.Error(), "database connection failed")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepo_ListEvents_EmptyResult(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewEventRepo(gormDB)

	rows := sqlmock.NewRows([]string{"id", "name", "status", "create_date", "update_date", "delete_date"})

	mock.ExpectQuery(`SELECT \* FROM "events"`).
		WillReturnRows(rows)

	ctx := context.Background()
	events, err := repo.ListEvents(ctx)

	assert.NoError(t, err)
	assert.Empty(t, events)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepo_CreateEvent_Success(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewEventRepo(gormDB)

	event := model.Event{
		ID:         "event-123",
		Name:       "New Test Event",
		Status:     model.Created,
		CreateDate: time.Now(),
		UpdateDate: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "events"`).
		WithArgs(event.ID, event.Name, event.Status, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	ctx := context.Background()
	err := repo.CreateEvent(ctx, event)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepo_CreateEvent_DatabaseError(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewEventRepo(gormDB)

	event := model.Event{
		ID:         "event-123",
		Name:       "New Test Event",
		Status:     model.Created,
		CreateDate: time.Now(),
		UpdateDate: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "events"`).
		WithArgs(event.ID, event.Name, event.Status, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnError(errors.New("database insert failed"))
	mock.ExpectRollback()

	ctx := context.Background()
	err := repo.CreateEvent(ctx, event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database insert failed")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepo_CreateEvent_DuplicateID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}()

	repo := NewEventRepo(gormDB)

	event := model.Event{
		ID:         "duplicate-id",
		Name:       "Duplicate Event",
		Status:     model.Created,
		CreateDate: time.Now(),
		UpdateDate: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "events"`).
		WithArgs(event.ID, event.Name, event.Status, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	ctx := context.Background()
	err := repo.CreateEvent(ctx, event)

	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}