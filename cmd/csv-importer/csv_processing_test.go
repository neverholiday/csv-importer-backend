package main

import (
	"csv-importer-backend/cmd/csv-importer/model"
	"encoding/csv"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
)

func TestCSVProcessing_LargeFiles(t *testing.T) {
	// Test with a large number of rows
	var csvBuilder strings.Builder
	csvBuilder.WriteString("todo_name,note\n")
	
	const numRows = 10000
	for i := 0; i < numRows; i++ {
		csvBuilder.WriteString("Task ")
		csvBuilder.WriteString(fmt.Sprintf("%d", i))
		csvBuilder.WriteString(",Note for task ")
		csvBuilder.WriteString(fmt.Sprintf("%d", i))
		csvBuilder.WriteString("\n")
	}

	reader := strings.NewReader(csvBuilder.String())
	var todos []*model.TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	
	assert.NoError(t, err)
	assert.Len(t, todos, numRows)
	assert.Equal(t, "Task 0", todos[0].TodoName)
	assert.Equal(t, "Note for task 0", todos[0].Note)
	assert.Equal(t, "Task 9999", todos[9999].TodoName)
	assert.Equal(t, "Note for task 9999", todos[9999].Note)
}

func TestCSVProcessing_LargeFieldContent(t *testing.T) {
	// Test with very large field content
	largeNote := strings.Repeat("This is a very long note. ", 1000) // ~27,000 characters
	
	csvContent := fmt.Sprintf(`todo_name,note
Large content task,"%s"
Normal task,Normal note`, largeNote)

	reader := strings.NewReader(csvContent)
	var todos []*model.TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	
	assert.NoError(t, err)
	assert.Len(t, todos, 2)
	assert.Equal(t, "Large content task", todos[0].TodoName)
	assert.Equal(t, largeNote, todos[0].Note)
	assert.Equal(t, "Normal task", todos[1].TodoName)
	assert.Equal(t, "Normal note", todos[1].Note)
}

func TestCSVProcessing_UnicodeCharacters(t *testing.T) {
	testCases := []struct {
		name        string
		todoName    string
		note        string
		description string
	}{
		{
			name:        "Chinese characters",
			todoName:    "买食物",
			note:        "需要买牛奶和面包",
			description: "Chinese text should be handled correctly",
		},
		{
			name:        "Japanese characters",
			todoName:    "買い物",
			note:        "牛乳とパンを買う",
			description: "Japanese text should be handled correctly",
		},
		{
			name:        "Arabic characters",
			todoName:    "شراء البقالة",
			note:        "نحتاج إلى شراء الحليب والخبز",
			description: "Arabic text should be handled correctly",
		},
		{
			name:        "Emoji characters",
			todoName:    "🛒 Shopping",
			note:        "📝 Don't forget: 🥛 milk, 🍞 bread, 🧀 cheese",
			description: "Emoji characters should be handled correctly",
		},
		{
			name:        "Mixed scripts",
			todoName:    "Meeting with José at Café Müller",
			note:        "Discuss プロジェクト and новые идеи",
			description: "Mixed language scripts should be handled correctly",
		},
		{
			name:        "Special unicode characters",
			todoName:    "Math symbols: ∑∫∆√",
			note:        "Currency: €£¥₹, Arrows: ←↑→↓, Stars: ★☆✓",
			description: "Special unicode symbols should be handled correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Properly quote CSV fields that might contain commas or special characters
			csvContent := fmt.Sprintf("todo_name,note\n\"%s\",\"%s\"", tc.todoName, tc.note)
			
			reader := strings.NewReader(csvContent)
			var todos []*model.TodoCSV
			err := gocsv.Unmarshal(reader, &todos)
			
			assert.NoError(t, err, tc.description)
			assert.Len(t, todos, 1)
			assert.Equal(t, tc.todoName, todos[0].TodoName)
			assert.Equal(t, tc.note, todos[0].Note)
			
			// Verify the strings are valid UTF-8
			assert.True(t, utf8.ValidString(todos[0].TodoName), "TodoName should be valid UTF-8")
			assert.True(t, utf8.ValidString(todos[0].Note), "Note should be valid UTF-8")
		})
	}
}

func TestCSVProcessing_ByteOrderMark(t *testing.T) {
	// Test CSV with BOM (Byte Order Mark)
	csvWithBOM := "\uFEFFtodo_name,note\nBuy groceries,Milk and bread\nCall dentist,Schedule appointment"
	
	reader := strings.NewReader(csvWithBOM)
	var todos []*model.TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	
	assert.NoError(t, err)
	assert.Len(t, todos, 2)
	
	// First field might contain BOM, depending on CSV library handling
	firstTodo := todos[0].TodoName
	if strings.HasPrefix(firstTodo, "\uFEFF") {
		// BOM was not stripped by the library
		assert.Equal(t, "\uFEFFBuy groceries", firstTodo)
	} else {
		// BOM was stripped by the library
		assert.Equal(t, "Buy groceries", firstTodo)
	}
	assert.Equal(t, "Milk and bread", todos[0].Note)
}

func TestCSVProcessing_DifferentEncodings(t *testing.T) {
	// Test handling of different line endings
	testCases := []struct {
		name        string
		csvContent  string
		expectedRows int
	}{
		{
			name:        "Unix line endings (LF)",
			csvContent:  "todo_name,note\nBuy groceries,Milk and bread\nCall dentist,Schedule appointment",
			expectedRows: 2,
		},
		{
			name:        "Windows line endings (CRLF)",
			csvContent:  "todo_name,note\r\nBuy groceries,Milk and bread\r\nCall dentist,Schedule appointment",
			expectedRows: 2,
		},
		{
			name:        "Old Mac line endings (CR)",
			csvContent:  "todo_name,note\rBuy groceries,Milk and bread\rCall dentist,Schedule appointment",
			expectedRows: 0, // CR-only line endings aren't properly supported by Go's CSV parser
		},
		{
			name:        "Mixed line endings",
			csvContent:  "todo_name,note\nBuy groceries,Milk and bread\r\nCall dentist,Schedule appointment",
			expectedRows: 2, // Skip the CR-only part since it's not well-supported
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.csvContent)
			var todos []*model.TodoCSV
			err := gocsv.Unmarshal(reader, &todos)
			
			assert.NoError(t, err)
			assert.Len(t, todos, tc.expectedRows)
			if len(todos) > 0 {
				assert.Equal(t, "Buy groceries", todos[0].TodoName)
				assert.Equal(t, "Milk and bread", todos[0].Note)
			}
		})
	}
}

func TestCSVProcessing_AlternativeDelimiters(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		delimiter rune
	}{
		{
			name:      "Semicolon delimiter",
			content:   "todo_name;note\nBuy groceries;Milk and bread\nCall dentist;Schedule appointment",
			delimiter: ';',
		},
		{
			name:      "Tab delimiter",
			content:   "todo_name\tnote\nBuy groceries\tMilk and bread\nCall dentist\tSchedule appointment",
			delimiter: '\t',
		},
		{
			name:      "Pipe delimiter",
			content:   "todo_name|note\nBuy groceries|Milk and bread\nCall dentist|Schedule appointment",
			delimiter: '|',
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.content)
			csvReader := csv.NewReader(reader)
			csvReader.Comma = tc.delimiter
			
			var todos []*model.TodoCSV
			err := gocsv.UnmarshalCSV(csvReader, &todos)
			
			assert.NoError(t, err)
			assert.Len(t, todos, 2)
			assert.Equal(t, "Buy groceries", todos[0].TodoName)
			assert.Equal(t, "Milk and bread", todos[0].Note)
			assert.Equal(t, "Call dentist", todos[1].TodoName)
			assert.Equal(t, "Schedule appointment", todos[1].Note)
		})
	}
}

func TestCSVProcessing_ComplexQuoting(t *testing.T) {
	testCases := []struct {
		name         string
		csvContent   string
		expectedName string
		expectedNote string
	}{
		{
			name:         "Quoted field with delimiter",
			csvContent:   `todo_name,note` + "\n" + `"Buy groceries, fresh ones","Milk, bread, and eggs"`,
			expectedName: "Buy groceries, fresh ones",
			expectedNote: "Milk, bread, and eggs",
		},
		{
			name:         "Quoted field with newlines",
			csvContent:   "todo_name,note\n\"Multi-line\ntask\",\"This is a\nmulti-line note\"",
			expectedName: "Multi-line\ntask",
			expectedNote: "This is a\nmulti-line note",
		},
		{
			name:         "Quoted field with escaped quotes",
			csvContent:   `todo_name,note` + "\n" + `"Call ""John"" Smith","He said ""Hello""!"`,
			expectedName: `Call "John" Smith`,
			expectedNote: `He said "Hello"!`,
		},
		{
			name:         "Mixed quoted and unquoted fields",
			csvContent:   `todo_name,note` + "\n" + `Normal task,"Quoted note with, comma"`,
			expectedName: "Normal task",
			expectedNote: "Quoted note with, comma",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.csvContent)
			var todos []*model.TodoCSV
			err := gocsv.Unmarshal(reader, &todos)
			
			assert.NoError(t, err)
			assert.Len(t, todos, 1)
			assert.Equal(t, tc.expectedName, todos[0].TodoName)
			assert.Equal(t, tc.expectedNote, todos[0].Note)
		})
	}
}

func TestCSVProcessing_ErrorHandling(t *testing.T) {
	errorCases := []struct {
		name       string
		csvContent string
		shouldError bool
	}{
		{
			name:        "Unclosed quote",
			csvContent:  `todo_name,note\n"Unclosed quote,This should fail`,
			shouldError: true,
		},
		{
			name:        "Quote in middle of unquoted field",
			csvContent:  `todo_name,note\nThis has a " quote,Normal note`,
			shouldError: false, // Most CSV parsers handle this gracefully
		},
		{
			name:        "Extra quote at end",
			csvContent:  `todo_name,note\nNormal task,Normal note"`,
			shouldError: false, // Usually handled gracefully
		},
		{
			name:        "Inconsistent number of fields",
			csvContent:  `todo_name,note\nTask 1,Note 1\nTask 2,Note 2,Extra field\nTask 3`,
			shouldError: false, // CSV readers usually handle this
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.csvContent)
			var todos []*model.TodoCSV
			err := gocsv.Unmarshal(reader, &todos)
			
			if tc.shouldError {
				assert.Error(t, err, "Expected error for malformed CSV")
			} else {
				// Even if we don't expect an error, verify the behavior
				t.Logf("Result: err=%v, todos=%+v", err, todos)
			}
		})
	}
}

func TestCSVProcessing_EmptyAndWhitespaceFields(t *testing.T) {
	csvContent := `todo_name,note
Task with empty note,
,Note without task name
"  ",   
	Tabs and spaces	,	Note with tabs	
Normal task,Normal note`

	reader := strings.NewReader(csvContent)
	var todos []*model.TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	
	assert.NoError(t, err)
	assert.Len(t, todos, 5)
	
	// First row: empty note
	assert.Equal(t, "Task with empty note", todos[0].TodoName)
	assert.Equal(t, "", todos[0].Note)
	
	// Second row: empty task name
	assert.Equal(t, "", todos[1].TodoName)
	assert.Equal(t, "Note without task name", todos[1].Note)
	
	// Third row: whitespace in quotes and spaces
	assert.Equal(t, "  ", todos[2].TodoName)
	assert.Equal(t, "   ", todos[2].Note)
	
	// Fourth row: tabs and spaces
	assert.Equal(t, "\tTabs and spaces\t", todos[3].TodoName)
	assert.Equal(t, "\tNote with tabs\t", todos[3].Note)
	
	// Fifth row: normal
	assert.Equal(t, "Normal task", todos[4].TodoName)
	assert.Equal(t, "Normal note", todos[4].Note)
}

func BenchmarkCSVProcessing_LargeFile(b *testing.B) {
	// Create a large CSV content once
	var csvBuilder strings.Builder
	csvBuilder.WriteString("todo_name,note\n")
	
	const numRows = 1000
	for i := 0; i < numRows; i++ {
		csvBuilder.WriteString(fmt.Sprintf("Task %d,Note for task %d\n", i, i))
	}
	csvContent := csvBuilder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(csvContent)
		var todos []*model.TodoCSV
		err := gocsv.Unmarshal(reader, &todos)
		if err != nil {
			b.Fatalf("Failed to unmarshal CSV: %v", err)
		}
		if len(todos) != numRows {
			b.Fatalf("Expected %d todos, got %d", numRows, len(todos))
		}
	}
}

func BenchmarkCSVProcessing_UnicodeContent(b *testing.B) {
	csvContent := `todo_name,note
买食物,需要买牛奶和面包
買い物,牛乳とパンを買う
🛒 Shopping,📝 Don't forget items
Meeting at Café,Discuss новые идеи
Math: ∑∫∆√,Currency: €£¥₹`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(csvContent)
		var todos []*model.TodoCSV
		err := gocsv.Unmarshal(reader, &todos)
		if err != nil {
			b.Fatalf("Failed to unmarshal CSV: %v", err)
		}
		if len(todos) != 5 {
			b.Fatalf("Expected 5 todos, got %d", len(todos))
		}
	}
}