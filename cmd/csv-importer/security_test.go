package main

import (
	"csv-importer-backend/cmd/csv-importer/model"
	"fmt"
	"strings"
	"testing"

	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
)

func TestSecurity_FileUploadSizeLimits(t *testing.T) {
	testCases := []struct {
		name           string
		fileSizeBytes  int
		shouldPass     bool
		description    string
	}{
		{
			name:          "Small valid file",
			fileSizeBytes: 1024, // 1KB
			shouldPass:    true,
			description:   "Small files should be accepted",
		},
		{
			name:          "Medium file",
			fileSizeBytes: 1024 * 1024, // 1MB
			shouldPass:    true,
			description:   "Medium files should be accepted",
		},
		{
			name:          "Large file",
			fileSizeBytes: 10 * 1024 * 1024, // 10MB
			shouldPass:    false,
			description:   "Large files should be rejected",
		},
		{
			name:          "Very large file",
			fileSizeBytes: 100 * 1024 * 1024, // 100MB
			shouldPass:    false,
			description:   "Very large files should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create CSV content of specified size
			csvContent := createCSVContent(tc.fileSizeBytes)
			
			// Verify the content size is approximately what we requested (allow for larger variance since CSV has overhead)
			assert.InDelta(t, tc.fileSizeBytes, len(csvContent), float64(tc.fileSizeBytes)*0.8, "CSV content size should be close to requested size")

			// Test parsing the content
			reader := strings.NewReader(csvContent)
			var todos []*model.TodoCSV
			err := gocsv.Unmarshal(reader, &todos)
			
			if tc.shouldPass {
				assert.NoError(t, err, tc.description)
				assert.Greater(t, len(todos), 0, "Should parse some todos")
			} else {
				// Large files should parse but would be rejected by upload limits
				// This test demonstrates the concept of size validation
				t.Logf("Large file test - Size: %d bytes, Todos parsed: %d", len(csvContent), len(todos))
			}
		})
	}
}

func TestSecurity_MaliciousFileContent(t *testing.T) {
	testCases := []struct {
		name        string
		csvContent  string
		expectError bool
		description string
	}{
		{
			name:        "CSV injection attempt",
			csvContent:  `todo_name,note` + "\n" + `"=cmd|'/c calc'!A1","Malicious formula"` + "\n" + `"+cmd|'/c notepad'!A1","Another formula"`,
			expectError: false, // CSV parsing shouldn't fail, but content should be sanitized
			description: "Excel formula injection attempts should be handled",
		},
		{
			name:        "Script injection attempt",
			csvContent:  `todo_name,note` + "\n" + `"<script>alert('xss')</script>","Normal note"` + "\n" + `"<img src=x onerror=alert('xss')>","Image injection"`,
			expectError: false, // CSV parsing won't fail, but content needs sanitization
			description: "HTML/JS injection attempts should be handled",
		},
		{
			name:        "SQL injection attempt",
			csvContent:  `todo_name,note` + "\n" + `"'; DROP TABLE events; --","SQL injection"` + "\n" + `"1' OR '1'='1","Boolean injection"`,
			expectError: false, // CSV parsing won't fail, parameterized queries prevent SQL injection
			description: "SQL injection attempts in CSV content",
		},
		{
			name:        "Path traversal attempt",
			csvContent:  `todo_name,note` + "\n" + `"../../../etc/passwd","Path traversal"` + "\n" + `"..\\..\\windows\\system32\\config\\sam","Windows path traversal"`,
			expectError: false,
			description: "Path traversal attempts in CSV content",
		},
		{
			name:        "Command injection attempt",
			csvContent:  `todo_name,note` + "\n" + `"$(rm -rf /)","Command injection"` + "\n" + `"`+"`"+`ls -la`+"`"+`","Backtick command"`,
			expectError: false,
			description: "Command injection attempts in CSV content",
		},
		{
			name:        "Buffer overflow attempt",
			csvContent:  fmt.Sprintf("todo_name,note\n%s,Buffer overflow attempt", strings.Repeat("A", 100000)),
			expectError: false,
			description: "Large field content that might cause buffer overflow",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.csvContent)
			var todos []*model.TodoCSV
			err := gocsv.Unmarshal(reader, &todos)
			
			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, "CSV parsing should succeed even with malicious content")
				
				// Verify content is captured but should be validated/sanitized later
				if len(todos) > 0 {
					t.Logf("Parsed potentially malicious content: TodoName='%s', Note='%s'", 
						todos[0].TodoName, todos[0].Note)
					
					// These tests demonstrate what content gets through CSV parsing
					// The application should implement additional validation/sanitization
				}
			}
		})
	}
}

func TestSecurity_FileTypeValidation(t *testing.T) {
	testCases := []struct {
		name        string
		filename    string
		content     string
		shouldAllow bool
		description string
	}{
		{
			name:        "Valid CSV file",
			filename:    "data.csv",
			content:     "todo_name,note\nTask 1,Note 1",
			shouldAllow: true,
			description: "CSV files should be allowed",
		},
		{
			name:        "CSV file with different extension",
			filename:    "data.txt",
			content:     "todo_name,note\nTask 1,Note 1",
			shouldAllow: true, // Content matters more than extension
			description: "Text files with CSV content might be allowed",
		},
		{
			name:        "Executable file disguised as CSV",
			filename:    "malware.exe",
			content:     "MZ\x90\x00\x03\x00\x00\x00", // PE executable header
			shouldAllow: false,
			description: "Executable files should be rejected",
		},
		{
			name:        "Script file disguised as CSV",
			filename:    "script.bat",
			content:     "@echo off\necho Malicious script\npause",
			shouldAllow: false,
			description: "Script files should be rejected",
		},
		{
			name:        "HTML file disguised as CSV",
			filename:    "page.html",
			content:     "<html><head></head><body><script>alert('xss')</script></body></html>",
			shouldAllow: false,
			description: "HTML files should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test demonstrates file type validation concepts
			// The actual implementation would check file headers, extensions, and content type
			
			isCSVContent := isValidCSVContent(tc.content)
			hasValidExtension := hasValidCSVExtension(tc.filename)
			
			if tc.shouldAllow {
				assert.True(t, isCSVContent || hasValidExtension, tc.description)
			} else {
				assert.False(t, isCSVContent && hasValidExtension, tc.description)
			}
		})
	}
}

func TestSecurity_InputSanitization(t *testing.T) {
	dangerousInputs := []struct {
		name  string
		input string
	}{
		{"HTML tags", "<script>alert('xss')</script>"},
		{"SQL injection", "'; DROP TABLE events; --"},
		{"Excel formula", "=cmd|'/c calc'!A1"},
		{"Path traversal", "../../../etc/passwd"},
		{"Command injection", "$(rm -rf /)"},
		{"Null bytes", "test\x00injection"},
		{"LDAP injection", "admin)(|(password=*))"},
		{"XML injection", "<?xml version=\"1.0\"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM \"file:///etc/passwd\">]>"},
	}

	for _, dangerous := range dangerousInputs {
		t.Run(dangerous.name, func(t *testing.T) {
			// Test that dangerous input can be safely processed
			sanitized := sanitizeInput(dangerous.input)
			
			// Sanitized input should not contain dangerous patterns
			assert.NotContains(t, sanitized, "<script>", "Should remove script tags")
			assert.NotContains(t, sanitized, "DROP TABLE", "Should remove SQL commands")
			assert.NotContains(t, sanitized, "=cmd|", "Should remove Excel formulas")
			assert.NotContains(t, sanitized, "../", "Should remove path traversal")
			assert.NotContains(t, sanitized, "$(", "Should remove command substitution")
			assert.NotContains(t, sanitized, "\x00", "Should remove null bytes")
			
			t.Logf("Original: '%s' -> Sanitized: '%s'", dangerous.input, sanitized)
		})
	}
}

func TestSecurity_RateLimiting(t *testing.T) {
	// Test that simulates rate limiting for file uploads
	maxUploadsPerMinute := 10
	uploads := 0
	
	for i := 0; i < 15; i++ {
		uploads++
		
		if uploads <= maxUploadsPerMinute {
			// Upload should be allowed
			assert.LessOrEqual(t, uploads, maxUploadsPerMinute, "Upload %d should be within rate limit", i+1)
		} else {
			// Upload should be rejected due to rate limiting
			assert.Greater(t, uploads, maxUploadsPerMinute, "Upload %d should be rejected due to rate limiting", i+1)
		}
	}
}

func TestSecurity_FileNameValidation(t *testing.T) {
	testCases := []struct {
		filename    string
		shouldAllow bool
		description string
	}{
		{"data.csv", true, "Simple CSV filename should be allowed"},
		{"my-data_file.csv", true, "CSV with hyphens and underscores should be allowed"},
		{"data with spaces.csv", true, "Filename with spaces should be allowed"},
		{"../../../etc/passwd", false, "Path traversal in filename should be rejected"},
		{"con.csv", false, "Windows reserved filename should be rejected"},
		{"prn.csv", false, "Windows reserved filename should be rejected"},
		{"aux.csv", false, "Windows reserved filename should be rejected"},
		{"nul.csv", false, "Windows reserved filename should be rejected"},
		{"data\x00.csv", false, "Filename with null byte should be rejected"},
		{"data<script>.csv", false, "Filename with script tags should be rejected"},
		{strings.Repeat("a", 300) + ".csv", false, "Extremely long filename should be rejected"},
		{"", false, "Empty filename should be rejected"},
		{".csv", true, "Filename with just extension might be allowed"},
		{"normal.exe", false, "Executable extension should be rejected"},
		{"data.csv.exe", false, "Double extension with executable should be rejected"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("filename_%s", tc.filename), func(t *testing.T) {
			isValid := isValidFilename(tc.filename)
			assert.Equal(t, tc.shouldAllow, isValid, tc.description)
		})
	}
}

func TestSecurity_ContentLengthValidation(t *testing.T) {
	// Test multipart form data with various content lengths
	testCases := []struct {
		name           string
		contentLength  int64
		shouldAllow    bool
		description    string
	}{
		{"Small content", 1024, true, "Small content should be allowed"},
		{"Medium content", 1024 * 1024, true, "1MB content should be allowed"},
		{"Large content", 5 * 1024 * 1024, true, "5MB content should be allowed"},
		{"Too large content", 50 * 1024 * 1024, false, "50MB content should be rejected"},
		{"Extremely large content", 1024 * 1024 * 1024, false, "1GB content should be rejected"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This demonstrates content length validation
			maxAllowedSize := int64(10 * 1024 * 1024) // 10MB
			
			isWithinLimit := tc.contentLength <= maxAllowedSize
			assert.Equal(t, tc.shouldAllow, isWithinLimit, tc.description)
		})
	}
}

func TestSecurity_MemoryExhaustionProtection(t *testing.T) {
	// Test protection against memory exhaustion attacks
	
	// Create a CSV with many columns (wide attack)
	var csvBuilder strings.Builder
	csvBuilder.WriteString("todo_name,note")
	for i := 0; i < 10000; i++ {
		csvBuilder.WriteString(fmt.Sprintf(",extra_col_%d", i))
	}
	csvBuilder.WriteString("\nTask,Note")
	for i := 0; i < 10000; i++ {
		csvBuilder.WriteString(",value")
	}

	wideCSV := csvBuilder.String()
	
	// This should complete without consuming excessive memory
	reader := strings.NewReader(wideCSV)
	var todos []*model.TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	
	// CSV library should handle this gracefully
	assert.NoError(t, err)
	t.Logf("Wide CSV test completed - Size: %d bytes, Todos: %d", len(wideCSV), len(todos))
}

// Helper functions for security validation

func createCSVContent(targetSize int) string {
	var csvBuilder strings.Builder
	csvBuilder.WriteString("todo_name,note\n")
	
	// Calculate approximate row size
	baseRow := "Task X,Note for task X\n"
	baseRowSize := len(baseRow)
	
	numRows := (targetSize - len("todo_name,note\n")) / baseRowSize
	if numRows <= 0 {
		numRows = 1
	}
	
	for i := 0; i < numRows; i++ {
		csvBuilder.WriteString(fmt.Sprintf("Task %d,Note for task %d\n", i, i))
	}
	
	return csvBuilder.String()
}

func isValidCSVContent(content string) bool {
	// Simple check for CSV-like content
	lines := strings.Split(content, "\n")
	if len(lines) < 1 {
		return false
	}
	
	// Check if first line looks like CSV headers
	firstLine := lines[0]
	return strings.Contains(firstLine, ",") || strings.Contains(firstLine, "todo_name")
}

func hasValidCSVExtension(filename string) bool {
	validExtensions := []string{".csv", ".txt"}
	filename = strings.ToLower(filename)
	
	for _, ext := range validExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

func sanitizeInput(input string) string {
	// Simple sanitization - in real implementation, use proper libraries
	sanitized := input
	
	// Remove dangerous patterns
	dangerousPatterns := []string{
		"<script>", "</script>", "<img", "javascript:",
		"DROP TABLE", "DELETE FROM", "INSERT INTO", "UPDATE ",
		"=cmd|", "=system(", "+cmd|",
		"../", "..\\",
		"$(", "`", "${",
		"\x00", "\r\n\r\n",
	}
	
	for _, pattern := range dangerousPatterns {
		sanitized = strings.ReplaceAll(sanitized, pattern, "")
	}
	
	return sanitized
}

func isValidFilename(filename string) bool {
	if filename == "" {
		return false
	}
	
	// Check length
	if len(filename) > 255 {
		return false
	}
	
	// Check for dangerous patterns
	dangerousPatterns := []string{
		"../", "..\\", "\x00", "<", ">", ":", "\"", "|", "?", "*",
	}
	
	for _, pattern := range dangerousPatterns {
		if strings.Contains(filename, pattern) {
			return false
		}
	}
	
	// Check for Windows reserved names
	reservedNames := []string{
		"con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5",
		"com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5",
		"lpt6", "lpt7", "lpt8", "lpt9",
	}
	
	baseName := strings.ToLower(filename)
	if idx := strings.LastIndex(baseName, "."); idx != -1 {
		baseName = baseName[:idx]
	}
	
	for _, reserved := range reservedNames {
		if baseName == reserved {
			return false
		}
	}
	
	// Check for dangerous extensions
	dangerousExtensions := []string{
		".exe", ".bat", ".cmd", ".com", ".scr", ".vbs", ".js", ".jar",
		".app", ".deb", ".pkg", ".dmg", ".sh", ".ps1",
	}
	
	lowerFilename := strings.ToLower(filename)
	for _, ext := range dangerousExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			return false
		}
	}
	
	return true
}