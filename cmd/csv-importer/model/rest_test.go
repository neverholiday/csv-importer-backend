package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		response BaseResponse
		expected string
	}{
		{
			name: "Response with data and message",
			response: BaseResponse{
				Data:    map[string]interface{}{"id": "123", "name": "test"},
				Message: "success",
			},
			expected: `{"data":{"id":"123","name":"test"},"message":"success"}`,
		},
		{
			name: "Response with nil data",
			response: BaseResponse{
				Data:    nil,
				Message: "error occurred",
			},
			expected: `{"message":"error occurred"}`,
		},
		{
			name: "Response with empty message",
			response: BaseResponse{
				Data:    "test data",
				Message: "",
			},
			expected: `{"data":"test data","message":""}`,
		},
		{
			name: "Response with slice data",
			response: BaseResponse{
				Data:    []string{"item1", "item2"},
				Message: "success",
			},
			expected: `{"data":["item1","item2"],"message":"success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.response)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test unmarshaling back
			var unmarshaled BaseResponse
			err = json.Unmarshal(jsonData, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.response.Message, unmarshaled.Message)
		})
	}
}

func TestBaseResponse_OmitEmptyData(t *testing.T) {
	// Test that nil data is omitted from JSON
	response := BaseResponse{
		Data:    nil,
		Message: "test message",
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.NotContains(t, string(jsonData), "data")
	assert.Contains(t, string(jsonData), `"message":"test message"`)
}

func TestEventCreateRequest_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		request  EventCreateRequest
		expected string
	}{
		{
			name:     "Normal event name",
			request:  EventCreateRequest{Name: "My Event"},
			expected: `{"name":"My Event"}`,
		},
		{
			name:     "Empty name",
			request:  EventCreateRequest{Name: ""},
			expected: `{"name":""}`,
		},
		{
			name:     "Unicode event name",
			request:  EventCreateRequest{Name: "活动名称"},
			expected: `{"name":"活动名称"}`,
		},
		{
			name:     "Special characters",
			request:  EventCreateRequest{Name: "Event with \"quotes\" & symbols"},
			expected: `{"name":"Event with \"quotes\" & symbols"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test unmarshaling back
			var unmarshaled EventCreateRequest
			err = json.Unmarshal(jsonData, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.request.Name, unmarshaled.Name)
		})
	}
}

func TestEventCreateRequest_Validation(t *testing.T) {
	// Basic structural test - the model doesn't have built-in validation
	// but we can test that it handles various input scenarios correctly
	
	tests := []struct {
		name        string
		jsonInput   string
		expectError bool
		expectedName string
	}{
		{
			name:        "Valid JSON",
			jsonInput:   `{"name":"Test Event"}`,
			expectError: false,
			expectedName: "Test Event",
		},
		{
			name:        "Missing name field",
			jsonInput:   `{}`,
			expectError: false,
			expectedName: "",
		},
		{
			name:        "Null name field",
			jsonInput:   `{"name":null}`,
			expectError: false,
			expectedName: "",
		},
		{
			name:        "Invalid JSON",
			jsonInput:   `{"name":"Test Event"`,
			expectError: true,
		},
		{
			name:        "Extra fields ignored",
			jsonInput:   `{"name":"Test Event","extra":"ignored"}`,
			expectError: false,
			expectedName: "Test Event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request EventCreateRequest
			err := json.Unmarshal([]byte(tt.jsonInput), &request)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, request.Name)
			}
		})
	}
}

func TestBaseResponse_WithComplexData(t *testing.T) {
	// Test BaseResponse with complex nested data
	complexData := map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"id":     "1",
				"name":   "Event 1",
				"status": "draft",
				"todos":  []string{"todo1", "todo2"},
			},
			{
				"id":     "2",
				"name":   "Event 2",
				"status": "start",
				"todos":  nil,
			},
		},
		"total": 2,
		"metadata": map[string]string{
			"version": "1.0",
		},
	}

	response := BaseResponse{
		Data:    complexData,
		Message: "success",
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	
	var unmarshaled BaseResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "success", unmarshaled.Message)
	assert.NotNil(t, unmarshaled.Data)

	// Verify complex data structure is preserved
	dataMap, ok := unmarshaled.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, dataMap, "events")
	assert.Contains(t, dataMap, "total")
	assert.Contains(t, dataMap, "metadata")
}