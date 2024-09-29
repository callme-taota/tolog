package tolog

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var timeZone, _ = time.LoadLocation("Asia/Shanghai")

// TestToLog tests the ToLog package.
func TestToLog(t *testing.T) {
	SetLogTimeZone(timeZone)
	t.Run("LevelInsert", LevelLogInsert)
	t.Run("TestLogFunction", ManyLogInsert)
	t.Run("TestSingle", SingleLogInsert)
}

func LevelLogInsert(t *testing.T) {
	logPrefix := "TestLevelInsert"
	logFilePath := "./logs/" + logPrefix + "-log-" + time.Now().Format(string(DateOnly)) + ".log"
	cleanLogFiles(t, logFilePath)
	SetLogPrefix(logPrefix)

	LevelLogInsertTest(t)

	checkMessageExistInFile(t, logFilePath, "This is an info message")
}

func LevelLogInsertTest(t *testing.T) {
	defer CloseLogFile()

	SetLogTimeFormat(StampNano)

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
}

// TestLogFunction tests the logging functionality.
func ManyLogInsert(t *testing.T) {
	logPrefix := "TestManyInsert"
	logFilePath := "./logs/" + logPrefix + "-log-" + time.Now().Format(string(DateOnly)) + ".log"
	cleanLogFiles(t, logFilePath)
	SetLogPrefix(logPrefix)

	ManyLogInsertTest(t)

	// Check log file content
	checkMessageExistInFile(t, logFilePath, "Test log message: Log message number 5000")
}

func ManyLogInsertTest(t *testing.T) {
	defer CloseLogFile()

	SetLogTimeFormat(StampNano)

	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1) // Increment the WaitGroup counter
		go func(i int) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			Log5kInsert(i)
		}(i)
	}

	wg.Wait()
}

func Log5kInsert(k int) {
	const numMessages = 500
	// Test log messages
	for i := 0; i < numMessages; i++ {
		message := "Log message number " + fmt.Sprintf("%d", k*(i+1))
		Infof("Test log message: %s", message).PrintAndWriteSafe()
	}
}

func SingleLogInsert(t *testing.T) {
	logPrefix := "TestSingleInsert"
	logFilePath := "./logs/" + logPrefix + "-log-" + time.Now().Format(string(DateOnly)) + ".log"
	cleanLogFiles(t, logFilePath)
	SetLogPrefix(logPrefix)

	SingleLogInsertTest(t)

	checkMessageExistInFile(t, logFilePath, "This is an single message")
}

func SingleLogInsertTest(t *testing.T) {
	defer CloseLogFile()

	SetLogTimeFormat(StampNano)

	Infof("Test log message: %s", "This is an single message").PrintAndWriteSafe()
}

func checkMessageExistInFile(t *testing.T, filePath string, message string) {
	logFile, err := os.ReadFile(filePath)
	require.NoError(t, err)
	content := string(logFile)
	assert.True(t, strings.Contains(content, message))
}

func cleanLogFiles(t *testing.T, filePath string) {
	os.Remove(filePath)
}
