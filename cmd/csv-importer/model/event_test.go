package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvent_TableName(t *testing.T) {
	event := Event{}
	assert.Equal(t, "events", event.TableName())
}

func TestEvent_JSONSerialization(t *testing.T) {
	now := time.Now()
	event := Event{
		ID:         "test-id",
		Name:       "Test Event",
		Status:     Created,
		CreateDate: now,
		UpdateDate: now,
		DeleteDate: nil,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"id":"test-id"`)
	assert.Contains(t, string(jsonData), `"name":"Test Event"`)
	assert.Contains(t, string(jsonData), `"status":"draft"`)

	// Test JSON unmarshaling
	var unmarshaledEvent Event
	err = json.Unmarshal(jsonData, &unmarshaledEvent)
	assert.NoError(t, err)
	assert.Equal(t, event.ID, unmarshaledEvent.ID)
	assert.Equal(t, event.Name, unmarshaledEvent.Name)
	assert.Equal(t, event.Status, unmarshaledEvent.Status)
}

func TestEvent_JSONSerializationWithDeleteDate(t *testing.T) {
	now := time.Now()
	deleteTime := now.Add(24 * time.Hour)
	event := Event{
		ID:         "test-id",
		Name:       "Deleted Event",
		Status:     End,
		CreateDate: now,
		UpdateDate: now,
		DeleteDate: &deleteTime,
	}

	jsonData, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"delete_date"`)

	var unmarshaledEvent Event
	err = json.Unmarshal(jsonData, &unmarshaledEvent)
	assert.NoError(t, err)
	assert.NotNil(t, unmarshaledEvent.DeleteDate)
	assert.Equal(t, deleteTime.Unix(), unmarshaledEvent.DeleteDate.Unix())
}

func TestEventStatus_Constants(t *testing.T) {
	assert.Equal(t, EventStatus("draft"), Created)
	assert.Equal(t, EventStatus("start"), Start)
	assert.Equal(t, EventStatus("end"), End)
}

func TestEventStatus_String(t *testing.T) {
	assert.Equal(t, "draft", string(Created))
	assert.Equal(t, "start", string(Start))
	assert.Equal(t, "end", string(End))
}

func TestTodoEvent_JSONSerialization(t *testing.T) {
	now := time.Now()
	todoEvent := TodoEvent{
		ID:         "todo-1",
		EventID:    "event-1",
		CreateDate: now,
		UpdateDate: now,
		DeleteDate: nil,
	}

	jsonData, err := json.Marshal(todoEvent)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"id":"todo-1"`)
	assert.Contains(t, string(jsonData), `"event_id":"event-1"`)

	var unmarshaled TodoEvent
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, todoEvent.ID, unmarshaled.ID)
	assert.Equal(t, todoEvent.EventID, unmarshaled.EventID)
}

func TestTodoEvent_JSONSerializationWithDeleteDate(t *testing.T) {
	now := time.Now()
	deleteTime := now.Add(time.Hour)
	todoEvent := TodoEvent{
		ID:         "todo-1",
		EventID:    "event-1",
		CreateDate: now,
		UpdateDate: now,
		DeleteDate: &deleteTime,
	}

	jsonData, err := json.Marshal(todoEvent)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"delete_date"`)

	var unmarshaled TodoEvent
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.NotNil(t, unmarshaled.DeleteDate)
	assert.Equal(t, deleteTime.Unix(), unmarshaled.DeleteDate.Unix())
}