package tolog

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type LogStatus string

// Constants representing different log levels.
const (
	StatusInfo    LogStatus = "info"
	StatusWarning LogStatus = "warning"
	StatusError   LogStatus = "error"
	StatusDebug   LogStatus = "debug"
	StatusNotice  LogStatus = "notice"
	StatusUnknown LogStatus = "unknown"
)

type DateFormat string

const (
	Layout      DateFormat = "01/02 03:04:05PM '06 -0700" // The reference time, in numerical order.
	ANSIC       DateFormat = "Mon Jan _2 15:04:05 2006"
	UnixDate    DateFormat = "Mon Jan _2 15:04:05 MST 2006"
	RubyDate    DateFormat = "Mon Jan 02 15:04:05 -0700 2006"
	RFC822      DateFormat = "02 Jan 06 15:04 MST"
	RFC822Z     DateFormat = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	RFC850      DateFormat = "Monday, 02-Jan-06 15:04:05 MST"
	RFC1123     DateFormat = "Mon, 02 Jan 2006 15:04:05 MST"
	RFC1123Z    DateFormat = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	RFC3339     DateFormat = "2006-01-02T15:04:05Z07:00"
	RFC3339Nano DateFormat = "2006-01-02T15:04:05.999999999Z07:00"
	Kitchen     DateFormat = "3:04PM"
	// Handy time stamps.
	Stamp      DateFormat = "Jan _2 15:04:05"
	StampMilli DateFormat = "Jan _2 15:04:05.000"
	StampMicro DateFormat = "Jan _2 15:04:05.000000"
	StampNano  DateFormat = "Jan _2 15:04:05.000000000"
	DateTime   DateFormat = "2006-01-02 15:04:05"
	DateOnly   DateFormat = "2006-01-02"
	TimeOnly   DateFormat = "15:04:05"
)

var logFileDateFormat = DateOnly
var logTimeFormat = DateTime

var (
	// Background color codes for different log levels.
	colorInfoBg    = "\033[48;5;27m"  // blue background
	colorWarningBg = "\033[48;5;226m" // orange background
	colorErrorBg   = "\033[48;5;196m" // red background
	colorDebugBg   = "\033[48;5;45m"  // green background
	colorNoticeBg  = "\033[48;5;165m" // purple background
	colorReset     = "\033[0m"        // reset color
)

// Global variable to store the current log date.
var currentLogDate string

// LogfilePrefix The prefix of the log file, default is null. Use set prefix to set.
var LogfilePrefix = ""

// LogWithColor The variable of whether to use color in the log, default is true.
var LogWithColor = true

// LogTimeZone The time zoon logger will print time at. Default is Local.
var LogTimeZone = time.Local

// Variables for managing log file and writing to file concurrently.
var logFile *os.File
var writeChannel chan string
var closeChannel chan struct{}
var isLogFileClosed bool = true
var wg sync.WaitGroup

// The size of go channel, default 300.
var channelSize = 300

// The time of writing to file, default 500ms.
var logTicker = time.Millisecond * 500

// ToLog represents a log entry with various attributes.
type ToLog struct {
	logType    LogStatus
	logContext string
	logTime    string
	FullLog    string
}

// Options is a function type for specifying log options using functional options pattern.
type Options func(l *ToLog)

// WithType sets the log type using functional options.
func WithType(level LogStatus) Options {
	return func(l *ToLog) {
		if level != StatusInfo && level != StatusWarning && level != StatusError && level != StatusNotice && level != StatusDebug {
			level = StatusUnknown
		}
		l.logType = level
		CreateFullLog(l)
	}
}

// WithContext sets the log context using functional options.
func WithContext(ctx string) Options {
	return func(l *ToLog) {
		l.logContext = ctx
	}
}

// SetLogWithColor sets the log shows colors or not.
func SetLogWithColor(flag bool) {
	LogWithColor = flag
}

// SetLogPrefix sets the log file prefix.
func SetLogPrefix(prefix string) {
	LogfilePrefix = prefix
	CloseLogFile()
	initLog()
}

// SetLogChannelSize set the size of go channel for cache.
func SetLogChannelSize(size int) {
	if size < 101 {
		return
	}
	channelSize = size
}

// SetLogTickerTime set the duration of saving log to file.
func SetLogTickerTime(duration time.Duration) {
	logTicker = duration
}

// SetLogFileDateFormat sets the date format for log file.
func SetLogFileDateFormat(format DateFormat) {
	logFileDateFormat = format
}

// SetLogTimeFormat sets the date format for log time.
func SetLogTimeFormat(format DateFormat) {
	logTimeFormat = format
}

// SetLogTimeZone sets the time zone for log time.
func SetLogTimeZone(zone *time.Location) {
	LogTimeZone = zone
}

// Log creates a new ToLog instance with default values and applies any specified options.
func Log(options ...Options) *ToLog {
	tolog := &ToLog{
		logType:    StatusInfo,
		logContext: "",
		logTime:    time.Now().In(LogTimeZone).Format(string(logTimeFormat)),
	}

	for _, option := range options {
		option(tolog)
	}

	return tolog
}

// Context sets the log context for an existing ToLog instance.
func (l *ToLog) Context(ctx string) *ToLog {
	l.logContext = ctx
	CreateFullLog(l)
	return l
}

// Type sets the log type for an existing ToLog instance.
func (l *ToLog) Type(le string) *ToLog {
	level := strings.ToLower(le)
	if level != string(StatusInfo) && level != string(StatusWarning) && level != string(StatusError) && level != string(StatusNotice) && level != string(StatusDebug) {
		level = string(StatusUnknown)
	}
	l.logType = LogStatus(level)
	CreateFullLog(l)
	return l
}

// Info sets the log type to "info" and sets the log context for an existing ToLog instance.
func Info(ctx string) *ToLog {
	l := Log()
	l.logType = StatusInfo
	l.logContext = ctx
	CreateFullLog(l)
	return l
}

// Infof sets the log type to "info" and sets the formatted log context for an existing ToLog instance.
func Infof(format string, a ...any) *ToLog {
	l := Log()
	l.logType = StatusInfo
	l.logContext = fmt.Sprintf(format, a...)
	CreateFullLog(l)
	return l
}

// Infoln sets the log type to "info" and sets the log context with a newline for an existing ToLog instance.
func Infoln(a ...any) *ToLog {
	l := Log()
	l.logType = StatusInfo
	l.logContext = fmt.Sprintln(a...)
	CreateFullLog(l)
	return l
}

// Warning sets the log type to "warning" and sets the log context for an existing ToLog instance.
func Warning(ctx string) *ToLog {
	l := Log()
	l.logType = StatusWarning
	l.logContext = ctx
	CreateFullLog(l)
	return l
}

// Warningf sets the log type to "warning" and sets the formatted log context for an existing ToLog instance.
func Warningf(format string, a ...any) *ToLog {
	l := Log()
	l.logType = StatusWarning
	l.logContext = fmt.Sprintf(format, a...)
	CreateFullLog(l)
	return l
}

// Warningln sets the log type to "warning" and sets the log context with a newline for an existing ToLog instance.
func Warningln(a ...any) *ToLog {
	l := Log()
	l.logType = StatusWarning
	l.logContext = fmt.Sprintln(a...)
	CreateFullLog(l)
	return l
}

// Error sets the log type to "error" and sets the log context for an existing ToLog instance.
func Error(ctx string) *ToLog {
	l := Log()
	l.logType = StatusError
	l.logContext = ctx
	CreateFullLog(l)
	return l
}

// Errorf sets the log type to "error" and sets the formatted log context for an existing ToLog instance.
func Errorf(format string, a ...any) *ToLog {
	l := Log()
	l.logType = StatusError
	l.logContext = fmt.Sprintf(format, a...)
	CreateFullLog(l)
	return l
}

// Errorln sets the log type to "error" and sets the log context with a newline for an existing ToLog instance.
func Errorln(a ...any) *ToLog {
	l := Log()
	l.logType = StatusError
	l.logContext = fmt.Sprintln(a...)
	CreateFullLog(l)
	return l
}

// Notice sets the log type to "notice" and sets the log context for an existing ToLog instance.
func Notice(ctx string) *ToLog {
	l := Log()
	l.logType = StatusNotice
	l.logContext = ctx
	CreateFullLog(l)
	return l
}

// Noticef sets the log type to "notice" and sets the formatted log context for an existing ToLog instance.
func Noticef(format string, a ...any) *ToLog {
	l := Log()
	l.logType = StatusNotice
	l.logContext = fmt.Sprintf(format, a...)
	CreateFullLog(l)
	return l
}

// Noticeln sets the log type to "notice" and sets the log context with a newline for an existing ToLog instance.
func Noticeln(a ...any) *ToLog {
	l := Log()
	l.logType = StatusNotice
	l.logContext = fmt.Sprintln(a...)
	CreateFullLog(l)
	return l
}

// Debug sets the log type to "debug" and sets the log context for an existing ToLog instance.
func Debug(ctx string) *ToLog {
	l := Log()
	l.logType = StatusDebug
	l.logContext = ctx
	CreateFullLog(l)
	return l
}

// Debugf sets the log type to "debug" and sets the formatted log context for an existing ToLog instance.
func Debugf(format string, a ...any) *ToLog {
	l := Log()
	l.logType = StatusDebug
	l.logContext = fmt.Sprintf(format, a...)
	CreateFullLog(l)
	return l
}

// Debugln sets the log type to "debug" and sets the log context with a newline for an existing ToLog instance.
func Debugln(a ...any) *ToLog {
	l := Log()
	l.logType = StatusDebug
	l.logContext = fmt.Sprintln(a...)
	CreateFullLog(l)
	return l
}

// PrintLog prints the full log to the console for an existing ToLog instance.
func (l *ToLog) PrintLog() *ToLog {
	CreateFullLog(l)
	fmt.Println(l.FullLog)
	return l
}

// CreateFullLog creates the full log message by combining log time, type, and context.
func CreateFullLog(l *ToLog) {
	var bgColor string

	if !LogWithColor {
		fullLog := "[" + l.logTime + "] [" + string(l.logType) + "] " + " " + l.logContext
		l.FullLog = fullLog
		return
	}
	switch l.logType {
	case StatusInfo:
		bgColor = colorInfoBg
	case StatusWarning:
		bgColor = colorWarningBg
	case StatusError:
		bgColor = colorErrorBg
	case StatusDebug:
		bgColor = colorDebugBg
	case StatusNotice:
		bgColor = colorNoticeBg
	default:
		bgColor = ""
	}

	fullLog := "[" + l.logTime + "] " + bgColor + " " + string(l.logType) + " " + colorReset + " " + l.logContext
	l.FullLog = fullLog
	return
}

// Deprecated:  WriteSafe instead
func (l *ToLog) Write() {
	CreateFullLog(l)
	if logFile == nil {
		err := initLog()
		if err != nil {
			return
		}
	}
	if LogWithColor {
		logFile.WriteString(stripColors(l.FullLog) + "\n")

	} else {
		logFile.WriteString(l.FullLog + "\n")
	}
	return
}

// WriteSafe writes the full log to the log file using a concurrent channel.
func (l *ToLog) WriteSafe() {
	CreateFullLog(l)
	if logFile == nil {
		err := initLog()
		if err != nil {
			return
		}
	}
	writeChannel <- l.FullLog + "\n"
}

// Deprecated:  PrintAndWriteSafe instead
func (l *ToLog) PrintAndWrite() {
	CreateFullLog(l)
	fmt.Println(l.FullLog)
	if logFile == nil {
		err := initLog()
		if err != nil {
			return
		}
	}
	if LogWithColor {
		logFile.WriteString(stripColors(l.FullLog) + "\n")

	} else {
		logFile.WriteString(l.FullLog + "\n")
	}
	return
}

func (l *ToLog) PrintAndWriteSafe() {
	CreateFullLog(l)
	fmt.Println(l.FullLog)
	if logFile == nil {
		err := initLog()
		if err != nil {
			return
		}
	}
	writeChannel <- l.FullLog + "\n"
}

// writeToFile is a goroutine that continuously writes log entries to the log file using the channel.
func writeToFile() {
	defer wg.Done()
	buffer := []string{}
	ticker := time.NewTicker(logTicker)
	defer ticker.Stop()
	for {
		select {
		case logEntry := <-writeChannel:
			buffer = append(buffer, logEntry)
			if len(buffer) >= 100 {
				flushBuffer(&buffer)
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				flushBuffer(&buffer)
			}
		case <-closeChannel:
			if len(buffer) > 0 {
				flushBuffer(&buffer)
			}

			for len(writeChannel) > 0 {
				logEntry := <-writeChannel
				buffer = append(buffer, logEntry)
				if len(buffer) >= 100 {
					flushBuffer(&buffer)
				}
			}

			if len(buffer) > 0 {
				flushBuffer(&buffer)
			}

			return
		}
	}
}

// flushBuffer writes the contents of the buffer to the log file.
func flushBuffer(buffer *[]string) {
	checkLogFileDate()
	data := strings.Join(*buffer, "")
	if LogWithColor {
		data = stripColors(data)
	}
	_, err := logFile.WriteString(data)
	if err != nil {
		fmt.Println("[error]", err)
		return
	}
	*buffer = (*buffer)[:0]
}

// checkLogFileDate can change file over a day
func checkLogFileDate() {
	currentDay := time.Now().In(LogTimeZone).Format(string(logFileDateFormat))
	if currentLogDate != currentDay {
		CloseLogFile()
		initLog()
	}
}

// initLog initializes the log file and sets up the writeToFile goroutine.
func initLog() error {
	currentDay := time.Now().In(LogTimeZone).Format(string(logFileDateFormat))
	logFilePath := ""
	if LogfilePrefix != "" {
		logFilePath = "./logs/" + LogfilePrefix + "-log-" + currentDay + ".log"
	} else {
		logFilePath = "./logs/log-" + currentDay + ".log"
	}
	currentLogDate = currentDay

	// Create the logs directory if it doesn't exist
	logDir := "./logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err = os.Mkdir(logDir, 0755)
		if err != nil {
			fmt.Println("[error] Failed to create logs directory:", err)
			return err
		}
	}

	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("[error]", err)
		return err
	}
	logFile = file

	isLogFileClosed = false

	writeChannel = make(chan string, channelSize)
	closeChannel = make(chan struct{})
	wg.Add(1)
	go writeToFile()

	return nil
}

// CloseLogFile closes the log file.
func CloseLogFile() {
	if logFile == nil || isLogFileClosed {
		return
	}

	close(closeChannel)

	if writeChannel != nil { // wait the writeToFile goroutine to finish
		close(writeChannel)
	}

	wg.Wait() // wait the writeToFile goroutine to finish

	err := logFile.Close()
	if err != nil {
		log.Fatal("Failed to close log file:", err)
		return
	}
	isLogFileClosed = true
	logFile = nil
}

var replacements = []struct {
	old string
	new string
}{
	{colorInfoBg, ""},
	{colorWarningBg, ""},
	{colorErrorBg, ""},
	{colorDebugBg, ""},
	{colorNoticeBg, ""},
	{colorReset, ""},
}

// stripColors removes ANSI color codes from a string
func stripColors(log string) string {
	for _, r := range replacements {
		log = strings.ReplaceAll(log, r.old, r.new)
	}
	return log
}
