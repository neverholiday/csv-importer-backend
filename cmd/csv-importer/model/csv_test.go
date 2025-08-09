package model

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
)

func TestTodoCSV_CSVTags(t *testing.T) {
	todo := TodoCSV{
		TodoName: "Buy groceries",
		Note:     "Milk and bread",
	}

	// Test CSV marshaling
	var buf bytes.Buffer
	err := gocsv.Marshal([]*TodoCSV{&todo}, &buf)
	assert.NoError(t, err)
	
	csvContent := buf.String()
	assert.Contains(t, csvContent, "todo_name,note")
	assert.Contains(t, csvContent, "Buy groceries,Milk and bread")
}

func TestTodoCSV_CSVUnmarshaling(t *testing.T) {
	csvContent := `todo_name,note
Buy groceries,Milk and bread
Call dentist,Schedule appointment`

	reader := strings.NewReader(csvContent)
	var todos []*TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	assert.NoError(t, err)
	assert.Len(t, todos, 2)
	assert.Equal(t, "Buy groceries", todos[0].TodoName)
	assert.Equal(t, "Milk and bread", todos[0].Note)
	assert.Equal(t, "Call dentist", todos[1].TodoName)
	assert.Equal(t, "Schedule appointment", todos[1].Note)
}

func TestTodoCSV_EmptyFields(t *testing.T) {
	csvContent := `todo_name,note
Buy groceries,
,Important note
,`

	reader := strings.NewReader(csvContent)
	var todos []*TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	assert.NoError(t, err)
	assert.Len(t, todos, 3)
	assert.Equal(t, "Buy groceries", todos[0].TodoName)
	assert.Equal(t, "", todos[0].Note)
	assert.Equal(t, "", todos[1].TodoName)
	assert.Equal(t, "Important note", todos[1].Note)
	assert.Equal(t, "", todos[2].TodoName)
	assert.Equal(t, "", todos[2].Note)
}

func TestTodoCSV_UnicodeContent(t *testing.T) {
	csvContent := `todo_name,note
买食物,牛奶和面包
Café,Rendez-vous à 15h
🛒 Shopping,📝 Don't forget items`

	reader := strings.NewReader(csvContent)
	var todos []*TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	assert.NoError(t, err)
	assert.Len(t, todos, 3)
	assert.Equal(t, "买食物", todos[0].TodoName)
	assert.Equal(t, "牛奶和面包", todos[0].Note)
	assert.Equal(t, "Café", todos[1].TodoName)
	assert.Equal(t, "Rendez-vous à 15h", todos[1].Note)
	assert.Equal(t, "🛒 Shopping", todos[2].TodoName)
	assert.Equal(t, "📝 Don't forget items", todos[2].Note)
}

func TestTodoCSV_QuotedFields(t *testing.T) {
	csvContent := `todo_name,note
"Buy groceries","Milk, bread, and eggs"
"Call ""John"" Smith","Important meeting"
"Multi-line
note","This spans
multiple lines"`

	reader := strings.NewReader(csvContent)
	var todos []*TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	assert.NoError(t, err)
	assert.Len(t, todos, 3)
	assert.Equal(t, "Buy groceries", todos[0].TodoName)
	assert.Equal(t, "Milk, bread, and eggs", todos[0].Note)
	assert.Equal(t, `Call "John" Smith`, todos[1].TodoName)
	assert.Equal(t, "Important meeting", todos[1].Note)
	assert.Equal(t, "Multi-line\nnote", todos[2].TodoName)
	assert.Equal(t, "This spans\nmultiple lines", todos[2].Note)
}

func TestTodoCSV_InvalidCSV(t *testing.T) {
	// Test with unclosed quote
	csvContent := `todo_name,note
"Unclosed quote,This should fail`

	reader := strings.NewReader(csvContent)
	var todos []*TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	assert.Error(t, err)
}

func TestTodoCSV_WrongHeaders(t *testing.T) {
	csvContent := `wrong_header,another_wrong
Task 1,Note 1
Task 2,Note 2`

	reader := strings.NewReader(csvContent)
	var todos []*TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	assert.NoError(t, err)
	assert.Len(t, todos, 2)
	// Fields should be empty since headers don't match
	assert.Equal(t, "", todos[0].TodoName)
	assert.Equal(t, "", todos[0].Note)
}

func TestTodoCSV_HeadersOnly(t *testing.T) {
	csvContent := `todo_name,note`

	reader := strings.NewReader(csvContent)
	var todos []*TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	assert.NoError(t, err)
	assert.Len(t, todos, 0)
}

func TestTodoCSV_ExtraColumns(t *testing.T) {
	csvContent := `todo_name,note,extra_column
Buy groceries,Milk and bread,ignored
Call dentist,Schedule appointment,also ignored`

	reader := strings.NewReader(csvContent)
	var todos []*TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	assert.NoError(t, err)
	assert.Len(t, todos, 2)
	assert.Equal(t, "Buy groceries", todos[0].TodoName)
	assert.Equal(t, "Milk and bread", todos[0].Note)
}

func TestTodoCSV_LargeContent(t *testing.T) {
	var csvBuilder strings.Builder
	csvBuilder.WriteString("todo_name,note\n")
	
	// Generate 1000 rows of test data
	for i := 0; i < 1000; i++ {
		csvBuilder.WriteString("Task ")
		csvBuilder.WriteString(string(rune('0' + (i % 10))))
		csvBuilder.WriteString(",Note for task ")
		csvBuilder.WriteString(string(rune('0' + (i % 10))))
		csvBuilder.WriteString("\n")
	}

	reader := strings.NewReader(csvBuilder.String())
	var todos []*TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	assert.NoError(t, err)
	assert.Len(t, todos, 1000)
	assert.Equal(t, "Task 0", todos[0].TodoName)
	assert.Equal(t, "Note for task 0", todos[0].Note)
	assert.Equal(t, "Task 9", todos[999].TodoName)
	assert.Equal(t, "Note for task 9", todos[999].Note)
}

func TestTodoCSV_DifferentDelimiters(t *testing.T) {
	// Test with semicolon delimiter
	csvContent := `todo_name;note
Buy groceries;Milk and bread
Call dentist;Schedule appointment`

	reader := strings.NewReader(csvContent)
	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';'
	
	var todos []*TodoCSV
	err := gocsv.UnmarshalCSV(csvReader, &todos)
	assert.NoError(t, err)
	assert.Len(t, todos, 2)
	assert.Equal(t, "Buy groceries", todos[0].TodoName)
	assert.Equal(t, "Milk and bread", todos[0].Note)
}

func TestTodoCSV_TabDelimited(t *testing.T) {
	// Test with tab delimiter
	csvContent := "todo_name\tnote\nBuy groceries\tMilk and bread\nCall dentist\tSchedule appointment"

	reader := strings.NewReader(csvContent)
	csvReader := csv.NewReader(reader)
	csvReader.Comma = '\t'
	
	var todos []*TodoCSV
	err := gocsv.UnmarshalCSV(csvReader, &todos)
	assert.NoError(t, err)
	assert.Len(t, todos, 2)
	assert.Equal(t, "Buy groceries", todos[0].TodoName)
	assert.Equal(t, "Milk and bread", todos[0].Note)
}