package main

import (
	"context"
	"csv-importer-backend/cmd/csv-importer/model"
	"csv-importer-backend/cmd/csv-importer/repository"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPerformance_ConcurrentEventCreation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	repo := repository.NewEventRepo(gormDB)

	// Number of concurrent operations
	numGoroutines := 10
	numEventsPerGoroutine := 100
	totalEvents := numGoroutines * numEventsPerGoroutine

	// Set up mock expectations for all operations
	for i := 0; i < totalEvents; i++ {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "events"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
			WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
		mock.ExpectCommit()
	}

	var wg sync.WaitGroup
	errors := make(chan error, totalEvents)
	
	startTime := time.Now()

	// Launch concurrent goroutines
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for i := 0; i < numEventsPerGoroutine; i++ {
				event := model.Event{
					ID:         fmt.Sprintf("perf-test-%d-%d", goroutineID, i),
					Name:       fmt.Sprintf("Performance Test Event %d-%d", goroutineID, i),
					Status:     model.Created,
					CreateDate: time.Now(),
					UpdateDate: time.Now(),
				}
				
				err := repo.CreateEvent(context.Background(), event)
				if err != nil {
					errors <- err
				}
			}
		}(g)
	}

	wg.Wait()
	close(errors)
	
	duration := time.Since(startTime)

	// Check for errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			t.Errorf("Error in concurrent operation: %v", err)
			errorCount++
		}
	}

	assert.Equal(t, 0, errorCount, "No errors should occur during concurrent operations")
	assert.NoError(t, mock.ExpectationsWereMet())

	t.Logf("Created %d events concurrently in %v (%.2f events/sec)", 
		totalEvents, duration, float64(totalEvents)/duration.Seconds())
}

func TestPerformance_ConcurrentEventListing(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	repo := repository.NewEventRepo(gormDB)

	// Number of concurrent read operations
	numGoroutines := 20
	numReadsPerGoroutine := 50
	totalReads := numGoroutines * numReadsPerGoroutine

	// Mock data
	rows := sqlmock.NewRows([]string{"id", "name", "status", "create_date", "update_date", "delete_date"}).
		AddRow("event-1", "Event 1", "draft", time.Now(), time.Now(), nil).
		AddRow("event-2", "Event 2", "start", time.Now(), time.Now(), nil)

	// Set up mock expectations for all read operations
	for i := 0; i < totalReads; i++ {
		mock.ExpectQuery(`SELECT \* FROM "events"`).
			WillReturnRows(rows)
	}

	var wg sync.WaitGroup
	results := make(chan []model.Event, totalReads)
	errors := make(chan error, totalReads)
	
	startTime := time.Now()

	// Launch concurrent goroutines
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for i := 0; i < numReadsPerGoroutine; i++ {
				events, err := repo.ListEvents(context.Background())
				if err != nil {
					errors <- err
				} else {
					results <- events
				}
			}
		}(g)
	}

	wg.Wait()
	close(errors)
	close(results)
	
	duration := time.Since(startTime)

	// Check results
	errorCount := 0
	for err := range errors {
		if err != nil {
			t.Errorf("Error in concurrent read: %v", err)
			errorCount++
		}
	}

	resultCount := 0
	for events := range results {
		assert.Len(t, events, 2, "Each read should return 2 events")
		resultCount++
	}

	assert.Equal(t, 0, errorCount, "No errors should occur during concurrent reads")
	assert.Equal(t, totalReads, resultCount, "All reads should complete successfully")
	assert.NoError(t, mock.ExpectationsWereMet())

	t.Logf("Completed %d concurrent reads in %v (%.2f reads/sec)", 
		totalReads, duration, float64(totalReads)/duration.Seconds())
}

func TestPerformance_LargeCSVProcessing(t *testing.T) {
	// Test processing of large CSV files
	testSizes := []struct {
		name      string
		numRows   int
		maxTime   time.Duration
	}{
		{"Small CSV", 100, time.Millisecond * 10},
		{"Medium CSV", 1000, time.Millisecond * 100},
		{"Large CSV", 10000, time.Second * 1},
		{"Very Large CSV", 100000, time.Second * 10},
	}

	for _, testSize := range testSizes {
		t.Run(testSize.name, func(t *testing.T) {
			// Generate CSV content
			var csvBuilder strings.Builder
			csvBuilder.WriteString("todo_name,note\n")
			
			for i := 0; i < testSize.numRows; i++ {
				csvBuilder.WriteString(fmt.Sprintf("Task %d,Note for task %d with some additional content to make it realistic\n", i, i))
			}
			
			csvContent := csvBuilder.String()
			
			startTime := time.Now()
			
			reader := strings.NewReader(csvContent)
			var todos []*model.TodoCSV
			err := gocsv.Unmarshal(reader, &todos)
			
			duration := time.Since(startTime)
			
			assert.NoError(t, err, "CSV processing should succeed")
			assert.Len(t, todos, testSize.numRows, "Should parse correct number of rows")
			assert.Less(t, duration, testSize.maxTime, "Processing should complete within expected time")
			
			t.Logf("Processed %d rows in %v (%.2f rows/sec, %.2f MB/sec)", 
				testSize.numRows, duration, 
				float64(testSize.numRows)/duration.Seconds(),
				float64(len(csvContent))/1024/1024/duration.Seconds())
		})
	}
}

func TestPerformance_MemoryUsageMonitoring(t *testing.T) {
	// Monitor memory usage during CSV processing
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// Create a reasonably large CSV
	var csvBuilder strings.Builder
	csvBuilder.WriteString("todo_name,note\n")
	
	numRows := 50000
	for i := 0; i < numRows; i++ {
		csvBuilder.WriteString(fmt.Sprintf("Task %d,This is a longer note for task %d to test memory usage during processing\n", i, i))
	}
	
	csvContent := csvBuilder.String()
	
	// Process the CSV
	reader := strings.NewReader(csvContent)
	var todos []*model.TodoCSV
	err := gocsv.Unmarshal(reader, &todos)
	
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	assert.NoError(t, err)
	assert.Len(t, todos, numRows)
	
	memoryUsed := m2.TotalAlloc - m1.TotalAlloc
	heapUsed := m2.HeapAlloc - m1.HeapAlloc
	
	t.Logf("Memory usage - Total allocated: %d bytes, Heap: %d bytes", memoryUsed, heapUsed)
	t.Logf("CSV size: %d bytes, Memory efficiency: %.2f%%", 
		len(csvContent), float64(len(csvContent))/float64(memoryUsed)*100)
	
	// Memory usage should be reasonable (not more than 10x the CSV size)
	assert.Less(t, memoryUsed, uint64(len(csvContent)*10), "Memory usage should be reasonable")
}

func TestPerformance_ConcurrentCSVProcessing(t *testing.T) {
	// Test concurrent CSV processing
	numGoroutines := 10
	rowsPerCSV := 1000
	
	// Generate test data
	csvContents := make([]string, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		var csvBuilder strings.Builder
		csvBuilder.WriteString("todo_name,note\n")
		
		for j := 0; j < rowsPerCSV; j++ {
			csvBuilder.WriteString(fmt.Sprintf("Task %d-%d,Note for task %d-%d\n", i, j, i, j))
		}
		csvContents[i] = csvBuilder.String()
	}
	
	var wg sync.WaitGroup
	results := make(chan int, numGoroutines)
	errors := make(chan error, numGoroutines)
	
	startTime := time.Now()
	
	// Process CSVs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int, csvContent string) {
			defer wg.Done()
			
			reader := strings.NewReader(csvContent)
			var todos []*model.TodoCSV
			err := gocsv.Unmarshal(reader, &todos)
			
			if err != nil {
				errors <- err
			} else {
				results <- len(todos)
			}
		}(i, csvContents[i])
	}
	
	wg.Wait()
	close(results)
	close(errors)
	
	duration := time.Since(startTime)
	
	// Check results
	totalRows := 0
	for rowCount := range results {
		totalRows += rowCount
		assert.Equal(t, rowsPerCSV, rowCount, "Each CSV should parse correct number of rows")
	}
	
	errorCount := 0
	for err := range errors {
		if err != nil {
			t.Errorf("Error in concurrent CSV processing: %v", err)
			errorCount++
		}
	}
	
	assert.Equal(t, 0, errorCount, "No errors should occur during concurrent processing")
	assert.Equal(t, numGoroutines*rowsPerCSV, totalRows, "All rows should be processed")
	
	t.Logf("Processed %d CSV files (%d total rows) concurrently in %v (%.2f files/sec)", 
		numGoroutines, totalRows, duration, float64(numGoroutines)/duration.Seconds())
}

func TestPerformance_DatabaseConnectionPooling(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Set connection pool limits
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	repo := repository.NewEventRepo(gormDB)

	// Test concurrent operations with limited connection pool
	numGoroutines := 20
	operationsPerGoroutine := 10
	totalOperations := numGoroutines * operationsPerGoroutine

	// Set up mock expectations
	rows := sqlmock.NewRows([]string{"id", "name", "status", "create_date", "update_date", "delete_date"})
	for i := 0; i < totalOperations; i++ {
		mock.ExpectQuery(`SELECT \* FROM "events"`).
			WillReturnRows(rows)
	}

	var wg sync.WaitGroup
	completedOps := make(chan bool, totalOperations)
	
	startTime := time.Now()
	
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for i := 0; i < operationsPerGoroutine; i++ {
				_, err := repo.ListEvents(context.Background())
				assert.NoError(t, err, "Database operation should succeed despite connection pooling")
				completedOps <- true
			}
		}()
	}
	
	wg.Wait()
	close(completedOps)
	
	duration := time.Since(startTime)
	
	// Count completed operations
	completed := 0
	for range completedOps {
		completed++
	}
	
	assert.Equal(t, totalOperations, completed, "All operations should complete")
	assert.NoError(t, mock.ExpectationsWereMet())
	
	t.Logf("Completed %d database operations with connection pooling in %v (%.2f ops/sec)", 
		totalOperations, duration, float64(totalOperations)/duration.Seconds())
}

// Benchmarks for performance testing

func BenchmarkRepository_CreateEvent(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		b.Fatal(err)
	}

	repo := repository.NewEventRepo(gormDB)

	// Set up mock expectations for all benchmark iterations
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "events"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
			WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
		mock.ExpectCommit()
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		event := model.Event{
			ID:         fmt.Sprintf("bench-event-%d", i),
			Name:       fmt.Sprintf("Benchmark Event %d", i),
			Status:     model.Created,
			CreateDate: time.Now(),
			UpdateDate: time.Now(),
		}
		
		err := repo.CreateEvent(context.Background(), event)
		if err != nil {
			b.Fatalf("CreateEvent failed: %v", err)
		}
	}
}

func BenchmarkRepository_ListEvents(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		b.Fatal(err)
	}

	repo := repository.NewEventRepo(gormDB)

	rows := sqlmock.NewRows([]string{"id", "name", "status", "create_date", "update_date", "delete_date"}).
		AddRow("event-1", "Event 1", "draft", time.Now(), time.Now(), nil).
		AddRow("event-2", "Event 2", "start", time.Now(), time.Now(), nil)

	// Set up mock expectations for all benchmark iterations
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery(`SELECT \* FROM "events"`).
			WillReturnRows(rows)
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := repo.ListEvents(context.Background())
		if err != nil {
			b.Fatalf("ListEvents failed: %v", err)
		}
	}
}

func BenchmarkCSV_Processing(b *testing.B) {
	// Generate test CSV content
	var csvBuilder strings.Builder
	csvBuilder.WriteString("todo_name,note\n")
	
	for i := 0; i < 1000; i++ {
		csvBuilder.WriteString(fmt.Sprintf("Task %d,Note for task %d\n", i, i))
	}
	csvContent := csvBuilder.String()

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(csvContent)
		var todos []*model.TodoCSV
		err := gocsv.Unmarshal(reader, &todos)
		if err != nil {
			b.Fatalf("CSV processing failed: %v", err)
		}
		if len(todos) != 1000 {
			b.Fatalf("Expected 1000 todos, got %d", len(todos))
		}
	}
}

func BenchmarkCSV_LargeFile(b *testing.B) {
	// Generate large CSV content
	var csvBuilder strings.Builder
	csvBuilder.WriteString("todo_name,note\n")
	
	for i := 0; i < 10000; i++ {
		csvBuilder.WriteString(fmt.Sprintf("Task %d,This is a longer note for task %d to test performance\n", i, i))
	}
	csvContent := csvBuilder.String()

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(csvContent)
		var todos []*model.TodoCSV
		err := gocsv.Unmarshal(reader, &todos)
		if err != nil {
			b.Fatalf("CSV processing failed: %v", err)
		}
		if len(todos) != 10000 {
			b.Fatalf("Expected 10000 todos, got %d", len(todos))
		}
	}
}