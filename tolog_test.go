package tolog

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestLogFunction tests the logging functionality.
func TestLogFunction(t *testing.T) {
	// Initialize the log
	err := initLog()
	if err != nil {
		t.Fatalf("Failed to initialize log: %v", err)
	}
	defer CloseLogFile()

	SetLogTimeFormat(StampNano)
	// Test log messages
	testMessages := []struct {
		level   LogStatus
		message string
	}{
		{StatusInfo, "This is an info message"},
		{StatusWarning, "This is a warning message"},
		{StatusError, "This is an error message"},
		{StatusDebug, "This is a debug message"},
		{StatusNotice, "This is a notice message"},
	}
	Infof("Test log message: %s", testMessages[0].message).PrintAndWriteSafe()
	Warningf("Test log message: %s", testMessages[1].message).PrintAndWriteSafe()
	Errorf("Test log message: %s", testMessages[2].message).PrintAndWriteSafe()
	Debugf("Test log message: %s", testMessages[3].message).PrintAndWriteSafe()
	Noticef("Test log message: %s", testMessages[4].message).PrintAndWriteSafe()

	for i := 1; i <= 10; i++ {
		go ManyInsert(i)
	}

	// Allow some time for the logs to be written
	time.Sleep(4 * time.Second)
	// Check log file content
	logFilePath := "./logs/log-" + time.Now().Format(string(DateOnly)) + ".log"
	fileContent, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	content := string(fileContent)
	for _, tm := range testMessages {
		if !strings.Contains(content, tm.message) {
			t.Errorf("Expected log message '%s' not found", tm.message)
		}
	}
	if !strings.Contains(content, "Test log message: Log message number 5000") {
		t.Errorf("Expected log message '%s' not found", "Test log message: Log message number 5000")
	}
}

func ManyInsert(k int) {
	const numMessages = 500
	// Test log messages
	for i := 0; i < numMessages; i++ {
		message := "Log message number " + fmt.Sprintf("%d", k*(i+1))
		Infof("Test log message: %s", message).PrintAndWriteSafe()
	}
}
