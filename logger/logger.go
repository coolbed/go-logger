package logger

// Level represents log level.
type Level int32

// Unit represents file size uint, such as KB/MB/GB/TB.
type Unit int64

// RollType represents file rolling type.
type RollType int //dailyRolling ,rollingFile

// Log file suffix formats.
const (
	DateFormat = "2006-01-02"
	HourFormat = "2006-01-02_15"
)

var logLevel Level = 2 // default debug level

// Supported log file size units.
const (
	_       = iota
	KB Unit = 1 << (iota * 10)
	MB
	GB
	TB
)

// Supported log levels.
const (
	OFF   Level = -1
	ALL   Level = 0
	TRACE Level = 1
	DEBUG Level = 2
	INFO  Level = 3
	WARN  Level = 4
	ERROR Level = 5
	FATAL Level = 6
)

// Supported log file rolling types.
const (
	RollTypeDaily RollType = iota
	RollTypeHourly
	RollTypeFile
)

// SetConsole set if output the log to console.
func SetConsole(isConsole bool) {
	defaultlog.setConsole(isConsole)
}

// SetLevel set the log level.
func SetLevel(_level Level) {
	defaultlog.setLevel(_level)
}

// SetFormat set the log format.
func SetFormat(logFormat string) {
	defaultlog.setFormat(logFormat)
}

// SetRollingFile set the rolling file.
func SetRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit Unit) {
	defaultlog.setRollingFile(fileDir, fileName, maxNumber, maxSize, _unit)
}

// SetRollingDaily set the daily rolling file.
func SetRollingDaily(fileDir, fileName string) {
	defaultlog.setRollingDaily(fileDir, fileName)
}

// SetRollingHourly set the hourly rolling file.
func SetRollingHourly(fileDir, fileName string) {
	defaultlog.setRollingHourly(fileDir, fileName)
}

// Trace print trace level log.
func Trace(v ...interface{}) {
	defaultlog.trace(v...)
}

// Debug print debug level log.
func Debug(v ...interface{}) {
	defaultlog.debug(v...)
}

// Info print info level log.
func Info(v ...interface{}) {
	defaultlog.info(v...)
}

// Warn print warn level log.
func Warn(v ...interface{}) {
	defaultlog.warn(v...)
}

// Error print error level log.
func Error(v ...interface{}) {
	defaultlog.error(v...)
}

// Fatal print fatal level log.
func Fatal(v ...interface{}) {
	defaultlog.fatal(v...)
}

// SetLevelFile set level file.
func SetLevelFile(level Level, dir, fileName string) {
	defaultlog.setLevelFile(level, dir, fileName)
}
