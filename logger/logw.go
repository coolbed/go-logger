package logger

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	skip       = 4
	defaultlog = getdefaultLogger()
)

// Logger represents a logger.
type Logger struct {
	lb *logBean
}

// SetConsole set if output the log to console.
func (l *Logger) SetConsole(isConsole bool) {
	l.lb.setConsole(isConsole)
}

// SetLevel set the log level.
func (l *Logger) SetLevel(_level Level) {
	l.lb.setLevel(_level)
}

// SetFormat set the log format.
func (l *Logger) SetFormat(logFormat string) {
	l.lb.setFormat(logFormat)
}

// SetRollingFile set the rolling file.
func (l *Logger) SetRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit Unit) {
	l.lb.setRollingFile(fileDir, fileName, maxNumber, maxSize, _unit)
}

// SetRollingDaily set the daily rolling file.
func (l *Logger) SetRollingDaily(fileDir, fileName string) {
	l.lb.setRollingDaily(fileDir, fileName)
}

// SetRollingHourly set the hourly rolling file.
func (l *Logger) SetRollingHourly(fileDir, fileName string) {
	l.lb.setRollingHourly(fileDir, fileName)
}

// Trace print trace level log.
func (l *Logger) Trace(v ...interface{}) {
	l.lb.trace(v...)
}

// Debug print debug level log.
func (l *Logger) Debug(v ...interface{}) {
	l.lb.debug(v...)
}

// Info print info level log.
func (l *Logger) Info(v ...interface{}) {
	l.lb.info(v...)
}

// Warn print warn level log.
func (l *Logger) Warn(v ...interface{}) {
	l.lb.warn(v...)
}

// Error print error level log.
func (l *Logger) Error(v ...interface{}) {
	l.lb.error(v...)
}

// Fatal print fatal level log.
func (l *Logger) Fatal(v ...interface{}) {
	l.lb.fatal(v...)
}

// SetLevelFile set level file.
func (l *Logger) SetLevelFile(level Level, dir, fileName string) {
	l.lb.setLevelFile(level, dir, fileName)
}

type logBean struct {
	mu              *sync.Mutex
	logLevel        Level
	maxFileSize     int64
	maxFileCount    int32
	consoleAppender bool
	rolltype        RollType
	format          string
	id              string
	d, i, w, e, f   string //id
}

type fileBeanFactory struct {
	fbs map[string]*fileBean
	mu  *sync.RWMutex
}

var fbf = &fileBeanFactory{fbs: make(map[string]*fileBean, 0), mu: new(sync.RWMutex)}

func (f *fileBeanFactory) add(dir, filename string, _suffix int, maxsize int64, maxfileCount int32) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := md5str(fmt.Sprint(dir, filename))
	if _, ok := f.fbs[id]; !ok {
		f.fbs[id] = newFileBean(dir, filename, _suffix, maxsize, maxfileCount)
	}
}

func (f *fileBeanFactory) get(id string) *fileBean {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.fbs[id]
}

type fileBean struct {
	id           string
	dir          string
	filename     string
	_suffix      int
	_date        *time.Time
	_hour        *time.Time
	mu           *sync.RWMutex
	logfile      *os.File
	lg           *log.Logger
	filesize     int64
	maxFileSize  int64
	maxFileCount int32
}

// GetLogger returns a new created logger.
func GetLogger() (l *Logger) {
	l = new(Logger)
	l.lb = getdefaultLogger()
	return
}

func getdefaultLogger() (lb *logBean) {
	lb = &logBean{}
	lb.mu = new(sync.Mutex)
	lb.setConsole(true)
	return
}

func (lb *logBean) setConsole(isConsole bool) {
	lb.consoleAppender = isConsole
}

func (lb *logBean) setLevelFile(level Level, dir, fileName string) {
	key := md5str(fmt.Sprint(dir, fileName))
	switch level {
	case DEBUG:
		lb.d = key
	case INFO:
		lb.i = key
	case WARN:
		lb.w = key
	case ERROR:
		lb.e = key
	case FATAL:
		lb.f = key
	default:
		return
	}
	var _suffix = 0
	if lb.maxFileCount < 1<<31-1 {
		for i := 1; i < int(lb.maxFileCount); i++ {
			if isExist(dir + "/" + fileName + "." + strconv.Itoa(i)) {
				_suffix = i
			} else {
				break
			}
		}
	}
	fbf.add(dir, fileName, _suffix, lb.maxFileSize, lb.maxFileCount)
}

func (lb *logBean) setLevel(_level Level) {
	lb.logLevel = _level
}

func (lb *logBean) setFormat(logFormat string) {
	lb.format = logFormat
}

func (lb *logBean) setRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit Unit) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if maxNumber > 0 {
		lb.maxFileCount = maxNumber
	} else {
		lb.maxFileCount = 1<<31 - 1
	}
	lb.maxFileSize = maxSize * int64(_unit)
	lb.rolltype = RollTypeFile
	mkdirlog(fileDir)
	var _suffix = 0
	for i := 1; i < int(maxNumber); i++ {
		if isExist(fileDir + "/" + fileName + "." + strconv.Itoa(i)) {
			_suffix = i
		} else {
			break
		}
	}
	lb.id = md5str(fmt.Sprint(fileDir, fileName))
	fbf.add(fileDir, fileName, _suffix, lb.maxFileSize, lb.maxFileCount)
}

func (lb *logBean) setRollingDaily(fileDir, fileName string) {
	lb.rolltype = RollTypeDaily
	mkdirlog(fileDir)
	lb.id = md5str(fmt.Sprint(fileDir, fileName))
	fbf.add(fileDir, fileName, 0, 0, 0)
}

func (lb *logBean) setRollingHourly(fileDir, fileName string) {
	lb.rolltype = RollTypeDaily
	mkdirlog(fileDir)
	lb.id = md5str(fmt.Sprint(fileDir, fileName))
	fbf.add(fileDir, fileName, 0, 0, 0)
}

func (lb *logBean) console(v ...interface{}) {
	s := fmt.Sprint(v...)
	if lb.consoleAppender {
		_, file, line, _ := runtime.Caller(skip)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		if lb.format == "" {
			log.Println(file, strconv.Itoa(line), s)
		} else {
			vs := make([]interface{}, 0)
			vs = append(vs, file)
			vs = append(vs, strconv.Itoa(line))
			for _, vv := range v {
				vs = append(vs, vv)
			}
			log.Printf(fmt.Sprint("%s %s ", lb.format, "\n"), vs...)
		}
	}
}

func (lb *logBean) log(level string, v ...interface{}) {
	defer catchError()
	s := fmt.Sprint(v...)
	length := len([]byte(s))
	lg := fbf.get(lb.id)
	var _level = ALL
	switch level {
	case "TRACE":
		if lb.d != "" {
			lg = fbf.get(lb.d)
		}
		_level = TRACE
	case "DEBUG":
		if lb.d != "" {
			lg = fbf.get(lb.d)
		}
		_level = DEBUG
	case "INFO":
		if lb.i != "" {
			lg = fbf.get(lb.i)
		}
		_level = INFO
	case "WARN":
		if lb.w != "" {
			lg = fbf.get(lb.w)
		}
		_level = WARN
	case "ERROR":
		if lb.e != "" {
			lg = fbf.get(lb.e)
		}
		_level = ERROR
	case "FATAL":
		if lb.f != "" {
			lg = fbf.get(lb.f)
		}
		_level = FATAL
	}
	if lg != nil {
		lb.fileCheck(lg)
		lg.addsize(int64(length))
		if lb.logLevel <= _level {
			if lg != nil {
				if lb.format == "" {
					lg.write(level, s)
				} else {
					lg.writef(lb.format, v...)
				}
			}
			lb.console(v...)
		}
	} else {
		lb.console(v...)
	}
}

func (lb *logBean) trace(v ...interface{}) {
	lb.log("TRACE", v...)
}

func (lb *logBean) debug(v ...interface{}) {
	lb.log("DEBUG", v...)
}
func (lb *logBean) info(v ...interface{}) {
	lb.log("INFO", v...)
}
func (lb *logBean) warn(v ...interface{}) {
	lb.log("WARN", v...)
}
func (lb *logBean) error(v ...interface{}) {
	lb.log("ERROR", v...)
}
func (lb *logBean) fatal(v ...interface{}) {
	lb.log("FATAL", v...)
}

func (lb *logBean) fileCheck(fb *fileBean) {
	defer catchError()
	if lb.isMustRename(fb) {
		lb.mu.Lock()
		defer lb.mu.Unlock()
		if lb.isMustRename(fb) {
			fb.rename(lb.rolltype)
		}
	}
}

func (lb *logBean) isMustRename(fb *fileBean) bool {
	switch lb.rolltype {
	case RollTypeDaily:
		t, _ := time.Parse(DateFormat, time.Now().Format(DateFormat))
		if t.After(*fb._date) {
			return true
		}
	case RollTypeFile:
		return fb.isOverSize()
	case RollTypeHourly:
		t, _ := time.Parse(HourFormat, time.Now().Format(HourFormat))
		if t.After(*fb._hour) {
			return true
		}
	}

	return false
}

func (fb *fileBean) nextSuffix() int {
	return int(fb._suffix%int(fb.maxFileCount) + 1)
}

func newFileBean(fileDir, fileName string, _suffix int, maxSize int64, maxfileCount int32) (fb *fileBean) {
	t, _ := time.Parse(DateFormat, time.Now().Format(DateFormat))
	th, _ := time.Parse(HourFormat, time.Now().Format(HourFormat))
	fb = &fileBean{dir: fileDir, filename: fileName, _date: &t, _hour: &th, mu: new(sync.RWMutex)}
	fb.logfile, _ = os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	fb.lg = log.New(fb.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	fb._suffix = _suffix
	fb.maxFileSize = maxSize
	fb.maxFileCount = maxfileCount
	fb.filesize = fileSize(fileDir + "/" + fileName)
	fb._date = &t
	fb._hour = &th
	return
}

func (fb *fileBean) rename(rolltype RollType) {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	fb.close()
	nextfilename := ""
	switch rolltype {
	case RollTypeDaily:
		nextfilename = fmt.Sprint(fb.dir, "/", fb.filename, ".", fb._date.Format(DateFormat))
	case RollTypeFile:
		nextfilename = fmt.Sprint(fb.dir, "/", fb.filename, ".", fb.nextSuffix())
		fb._suffix = fb.nextSuffix()
	case RollTypeHourly:
		nextfilename = fmt.Sprint(fb.dir, "/", fb.filename, ".", fb._hour.Format(HourFormat))
	}
	if isExist(nextfilename) {
		os.Remove(nextfilename)
	}
	os.Rename(fb.dir+"/"+fb.filename, nextfilename)
	fb.logfile, _ = os.OpenFile(fb.dir+"/"+fb.filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	fb.lg = log.New(fb.logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	fb.filesize = fileSize(fb.dir + "/" + fb.filename)
	t, _ := time.Parse(DateFormat, time.Now().Format(DateFormat))
	fb._date = &t
	th, _ := time.Parse(HourFormat, time.Now().Format(HourFormat))
	fb._hour = &th
}

func (fb *fileBean) addsize(size int64) {
	atomic.AddInt64(&fb.filesize, size)
}

func (fb *fileBean) write(level string, v ...interface{}) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()
	s := fmt.Sprint(v...)
	fb.lg.Output(skip+1, fmt.Sprintln(level, s))
}

func (fb *fileBean) writef(format string, v ...interface{}) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()
	fb.lg.Output(skip+1, fmt.Sprintf(format, v...))
}

func (fb *fileBean) isOverSize() bool {
	return fb.filesize >= fb.maxFileSize
}

func (fb *fileBean) close() {
	fb.logfile.Close()
}

func mkdirlog(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0666); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

func fileSize(file string) int64 {
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func md5str(s string) string {
	m := md5.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}

func catchError() {
	if err := recover(); err != nil {
		fmt.Println(string(debug.Stack()))
	}
}
