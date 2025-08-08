package apis

import (
	"bytes"
	"context"
	"csv-importer-backend/cmd/csv-importer/model"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventRepo implements IEventRepo interface for testing
type MockEventRepo struct {
	mock.Mock
}

func (m *MockEventRepo) ListEvents(ctx context.Context) ([]model.Event, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Event), args.Error(1)
}

func (m *MockEventRepo) CreateEvent(ctx context.Context, event model.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestEventAPI_ListEvents_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockEventRepo)
	api := NewEventAPI(mockRepo)

	expectedEvents := []model.Event{
		{
			ID:         "event-1",
			Name:       "Test Event 1",
			Status:     model.Created,
			CreateDate: time.Now(),
			UpdateDate: time.Now(),
		},
		{
			ID:         "event-2",
			Name:       "Test Event 2",
			Status:     model.Start,
			CreateDate: time.Now(),
			UpdateDate: time.Now(),
		},
	}

	mockRepo.On("ListEvents", mock.Anything).Return(expectedEvents, nil)

	err := api.listEvents(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.BaseResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Message)
	
	// Convert response.Data to events slice for assertion
	eventsData, err := json.Marshal(response.Data)
	assert.NoError(t, err)
	
	var actualEvents []model.Event
	err = json.Unmarshal(eventsData, &actualEvents)
	assert.NoError(t, err)
	assert.Len(t, actualEvents, 2)
	assert.Equal(t, expectedEvents[0].ID, actualEvents[0].ID)
	assert.Equal(t, expectedEvents[1].ID, actualEvents[1].ID)

	mockRepo.AssertExpectations(t)
}

func TestEventAPI_ListEvents_RepositoryError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockEventRepo)
	api := NewEventAPI(mockRepo)

	mockRepo.On("ListEvents", mock.Anything).Return([]model.Event{}, errors.New("database connection failed"))

	err := api.listEvents(c)

	assert.NoError(t, err) // Echo doesn't return error for JSON responses
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response model.BaseResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "database connection failed")

	mockRepo.AssertExpectations(t)
}

func TestEventAPI_CreateEvent_ValidCSV(t *testing.T) {
	e := echo.New()

	// Create multipart form data with CSV file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add name field
	nameField, err := writer.CreateFormField("name")
	assert.NoError(t, err)
	_, err = nameField.Write([]byte("Test Event"))
	assert.NoError(t, err)

	// Add CSV file
	csvField, err := writer.CreateFormFile("csvfile", "test.csv")
	assert.NoError(t, err)
	csvContent := "todo_name,note\nBuy groceries,Get milk and bread\nCall dentist,Schedule appointment"
	_, err = csvField.Write([]byte(csvContent))
	assert.NoError(t, err)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/event", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockEventRepo)
	api := NewEventAPI(mockRepo)

	mockRepo.On("CreateEvent", mock.Anything, mock.Anything).Return(nil)

	err = api.createEvent(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.BaseResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Message)

	mockRepo.AssertExpectations(t)
}

func TestEventAPI_CreateEvent_MissingFile(t *testing.T) {
	e := echo.New()

	// Create multipart form data without CSV file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	nameField, err := writer.CreateFormField("name")
	assert.NoError(t, err)
	_, err = nameField.Write([]byte("Test Event"))
	assert.NoError(t, err)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/event", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockEventRepo)
	api := NewEventAPI(mockRepo)

	err = api.createEvent(c)

	assert.NoError(t, err) // Echo doesn't return error for JSON responses
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response model.BaseResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "no such file")

	// Don't assert expectations as repo shouldn't be called
}

func TestEventAPI_CreateEvent_InvalidCSV(t *testing.T) {
	e := echo.New()

	// Create multipart form data with invalid CSV
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	nameField, err := writer.CreateFormField("name")
	assert.NoError(t, err)
	_, err = nameField.Write([]byte("Test Event"))
	assert.NoError(t, err)

	csvField, err := writer.CreateFormFile("csvfile", "invalid.csv")
	assert.NoError(t, err)
	csvContent := "wrong_column,another_wrong\nTask 1,Note 1"
	_, err = csvField.Write([]byte(csvContent))
	assert.NoError(t, err)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/event", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockEventRepo)
	api := NewEventAPI(mockRepo)

	// Even with invalid CSV structure, the API currently processes it
	// This test shows current behavior - you might want to add validation
	mockRepo.On("CreateEvent", mock.Anything, mock.Anything).Return(nil)

	err = api.createEvent(c)

	assert.NoError(t, err)
	// Current implementation doesn't validate CSV structure, so it succeeds
	assert.Equal(t, http.StatusOK, rec.Code)

	mockRepo.AssertExpectations(t)
}

func TestEventAPI_CreateEvent_MalformedCSV(t *testing.T) {
	e := echo.New()

	// Create multipart form data with malformed CSV
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	nameField, err := writer.CreateFormField("name")
	assert.NoError(t, err)
	_, err = nameField.Write([]byte("Test Event"))
	assert.NoError(t, err)

	csvField, err := writer.CreateFormFile("csvfile", "malformed.csv")
	assert.NoError(t, err)
	csvContent := "todo_name,note\n\"Unclosed quote,This is bad\nAnother row,Good row"
	_, err = csvField.Write([]byte(csvContent))
	assert.NoError(t, err)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/event", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockEventRepo)
	api := NewEventAPI(mockRepo)

	err = api.createEvent(c)

	assert.NoError(t, err) // Echo doesn't return error for JSON responses
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response model.BaseResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	// Should contain CSV parsing error
	assert.NotEqual(t, "success", response.Message)

	// Don't assert expectations as repo shouldn't be called due to CSV error
}

func TestEventAPI_CreateEvent_RepositoryError(t *testing.T) {
	e := echo.New()

	// Create multipart form data with valid CSV
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	nameField, err := writer.CreateFormField("name")
	assert.NoError(t, err)
	_, err = nameField.Write([]byte("Test Event"))
	assert.NoError(t, err)

	csvField, err := writer.CreateFormFile("csvfile", "test.csv")
	assert.NoError(t, err)
	csvContent := "todo_name,note\nBuy groceries,Get milk and bread"
	_, err = csvField.Write([]byte(csvContent))
	assert.NoError(t, err)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/event", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockEventRepo)
	api := NewEventAPI(mockRepo)

	mockRepo.On("CreateEvent", mock.Anything, mock.Anything).Return(errors.New("database connection failed"))

	err = api.createEvent(c)

	assert.NoError(t, err) // Echo doesn't return error for JSON responses
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response model.BaseResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "database connection failed")

	mockRepo.AssertExpectations(t)
}

func TestEventAPI_CreateEvent_EmptyCSV(t *testing.T) {
	e := echo.New()

	// Create multipart form data with empty CSV (only headers)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	nameField, err := writer.CreateFormField("name")
	assert.NoError(t, err)
	_, err = nameField.Write([]byte("Test Event"))
	assert.NoError(t, err)

	csvField, err := writer.CreateFormFile("csvfile", "empty.csv")
	assert.NoError(t, err)
	csvContent := "todo_name,note" // Only headers, no data
	_, err = csvField.Write([]byte(csvContent))
	assert.NoError(t, err)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/event", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockEventRepo)
	api := NewEventAPI(mockRepo)

	mockRepo.On("CreateEvent", mock.Anything, mock.Anything).Return(nil)

	err = api.createEvent(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.BaseResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Message)

	mockRepo.AssertExpectations(t)
}

// Integration test using actual test data files
func TestEventAPI_CreateEvent_WithTestDataFiles(t *testing.T) {
	e := echo.New()

	testCases := []struct {
		name           string
		fileName       string
		expectedStatus int
		shouldCallRepo bool
	}{
		{
			name:           "Valid CSV file",
			fileName:       "valid.csv",
			expectedStatus: http.StatusOK,
			shouldCallRepo: true,
		},
		{
			name:           "Empty CSV file",
			fileName:       "empty.csv",
			expectedStatus: http.StatusOK,
			shouldCallRepo: true,
		},
		{
			name:           "Malformed CSV file",
			fileName:       "malformed.csv",
			expectedStatus: http.StatusInternalServerError,
			shouldCallRepo: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read test file
			filePath := filepath.Join("..", "..", "..", "testdata", tc.fileName)
			fileContent, err := os.ReadFile(filePath)
			assert.NoError(t, err)

			// Create multipart form data
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			nameField, err := writer.CreateFormField("name")
			assert.NoError(t, err)
			nameField.Write([]byte("Test Event"))

			csvField, err := writer.CreateFormFile("csvfile", tc.fileName)
			assert.NoError(t, err)
			_, err = csvField.Write(fileContent)
			assert.NoError(t, err)

			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/event", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockRepo := new(MockEventRepo)
			api := NewEventAPI(mockRepo)

			if tc.shouldCallRepo {
				mockRepo.On("CreateEvent", mock.Anything, mock.Anything).Return(nil)
			}

			err = api.createEvent(c)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}