package log

import (
	"fmt"
	"io"
	"runtime"

	"sync"

	"github.com/Sirupsen/logrus"
)

type Entry struct {
	*logrus.Entry
	hasLocation bool
}

func NewEntry(l *logrus.Logger) *Entry {
	return &Entry{
		Entry:       logrus.NewEntry(l),
		hasLocation: false,
	}
}

var (
	PositionKey = "where"
)

func SetPositionKey(k string) {
	PositionKey = k
}

func (entry *Entry) withLocation(n int) *Entry {
	if entry.Entry.Level >= logrus.InfoLevel {
		return entry
	}

	if !entry.hasLocation {
		_, file, line, _ := runtime.Caller(n)
		entry.Entry = entry.Entry.WithField(PositionKey, fmt.Sprintf("%s:%d", file, line))
		entry.hasLocation = true
	}
	return entry
}

func (entry *Entry) WithFields(fields logrus.Fields) *Entry {
	entry.Entry = entry.Entry.WithFields(fields)
	return entry
}

func (entry *Entry) WithField(key string, value interface{}) *Entry {
	entry.Entry = entry.Entry.WithField(key, value)
	return entry
}

func (entry *Entry) WithError(e error) *Entry {
	entry.Entry = entry.Entry.WithError(e)
	return entry
}

func (entry *Entry) Debug(args ...interface{}) {
	entry.withLocation(2).Entry.Debug(args...)
}

func (entry *Entry) Print(args ...interface{}) {
	entry.withLocation(2).Entry.Print(args...)
}

func (entry *Entry) Info(args ...interface{}) {
	entry.withLocation(2).Entry.Info(args...)
}

func (entry *Entry) Warn(args ...interface{}) {
	entry.withLocation(2).Entry.Warn(args...)
}

func (entry *Entry) Warning(args ...interface{}) {
	entry.withLocation(2).Entry.Warning(args...)
}

func (entry *Entry) Error(args ...interface{}) {
	entry.withLocation(2).Entry.Error(args...)
}

func (entry *Entry) Fatal(args ...interface{}) {
	entry.withLocation(2).Entry.Fatal(args...)
}

func (entry *Entry) Panic(args ...interface{}) {
	entry.withLocation(2).Entry.Panic(args...)
}

// Entry Printf family functions

func (entry *Entry) Debugf(format string, args ...interface{}) {
	entry.withLocation(2).Entry.Debugf(fmt.Sprintf(format, args...))
}

func (entry *Entry) Infof(format string, args ...interface{}) {
	entry.withLocation(2).Entry.Infof(fmt.Sprintf(format, args...))
}

func (entry *Entry) Printf(format string, args ...interface{}) {
	entry.withLocation(2).Entry.Printf(format, args...)
}

func (entry *Entry) Warnf(format string, args ...interface{}) {
	entry.withLocation(2).Entry.Warnf(fmt.Sprintf(format, args...))
}

func (entry *Entry) Warningf(format string, args ...interface{}) {
	entry.withLocation(2).Entry.Warningf(format, args...)
}

func (entry *Entry) Errorf(format string, args ...interface{}) {
	entry.withLocation(2).Entry.Errorf(fmt.Sprintf(format, args...))
}

func (entry *Entry) Fatalf(format string, args ...interface{}) {
	entry.withLocation(2).Entry.Fatalf(fmt.Sprintf(format, args...))
}

func (entry *Entry) Panicf(format string, args ...interface{}) {
	entry.withLocation(2).Entry.Panicf(fmt.Sprintf(format, args...))
}

// Entry Println family functions

func (entry *Entry) Debugln(args ...interface{}) {
	entry.withLocation(2).Entry.Debugln(args...)
}

func (entry *Entry) Infoln(args ...interface{}) {
	entry.withLocation(2).Entry.Infoln(args...)
}

func (entry *Entry) Println(args ...interface{}) {
	entry.withLocation(2).Entry.Println(args...)
}

func (entry *Entry) Warnln(args ...interface{}) {
	entry.withLocation(2).Entry.Warnln(args...)
}

func (entry *Entry) Warningln(args ...interface{}) {
	entry.withLocation(2).Entry.Warningln(args...)
}

func (entry *Entry) Errorln(args ...interface{}) {
	entry.withLocation(2).Entry.Errorln(args...)
}

func (entry *Entry) Fatalln(args ...interface{}) {
	entry.withLocation(2).Entry.Fatalln(args...)
}

func (entry *Entry) Panicln(args ...interface{}) {
	entry.withLocation(2).Entry.Panicln(args...)
}

var (
	// std is the name of the standard logger in stdlib `log`
	std = logrus.New()
	mu  = sync.Mutex{}
)

func StandardLogger() *logrus.Logger {
	return std
}

// SetOutput sets the standard logger output.
func SetOutput(out io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	std.Out = out
}

// SetFormatter sets the standard logger formatter.
func SetFormatter(formatter logrus.Formatter) {
	mu.Lock()
	defer mu.Unlock()
	std.Formatter = formatter
}

// SetLevel sets the standard logger level.
func SetLevel(level logrus.Level) {
	mu.Lock()
	defer mu.Unlock()
	std.Level = level
}

// GetLevel returns the standard logger level.
func GetLevel() logrus.Level {
	mu.Lock()
	defer mu.Unlock()
	return std.Level
}

// AddHook adds a hook to the standard logger hooks.
func AddHook(hook logrus.Hook) {
	mu.Lock()
	defer mu.Unlock()
	std.Hooks.Add(hook)
}

// WithError creates an entry from the standard logger and adds an error to it, using the value defined in ErrorKey as key.
func WithError(err error) logrus.FieldLogger {
	return std.WithField(logrus.ErrorKey, err)
}

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithField(key string, value interface{}) *Entry {
	return NewEntry(std).WithField(key, value)
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithFields(fields logrus.Fields) *Entry {
	return NewEntry(std).WithFields(fields)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	NewEntry(std).withLocation(2).Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	NewEntry(std).withLocation(2).Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	NewEntry(std).withLocation(2).Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	NewEntry(std).withLocation(2).Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(args ...interface{}) {
	NewEntry(std).withLocation(2).Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	NewEntry(std).withLocation(2).Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	NewEntry(std).withLocation(2).Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	NewEntry(std).withLocation(2).Fatal(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	NewEntry(std).withLocation(2).Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(format string, args ...interface{}) {
	NewEntry(std).withLocation(2).Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	NewEntry(std).withLocation(2).Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	NewEntry(std).withLocation(2).Warnf(format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func Warningf(format string, args ...interface{}) {
	NewEntry(std).withLocation(2).Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	NewEntry(std).withLocation(2).Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...interface{}) {
	NewEntry(std).withLocation(2).Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	NewEntry(std).withLocation(2).Fatalf(format, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...interface{}) {
	NewEntry(std).withLocation(2).Debugln(args...)
}

// Println logs a message at level Info on the standard logger.
func Println(args ...interface{}) {
	NewEntry(std).withLocation(2).Println(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...interface{}) {
	NewEntry(std).withLocation(2).Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...interface{}) {
	NewEntry(std).withLocation(2).Warnln(args...)
}

// Warningln logs a message at level Warn on the standard logger.
func Warningln(args ...interface{}) {
	NewEntry(std).withLocation(2).Warningln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...interface{}) {
	NewEntry(std).withLocation(2).Errorln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(args ...interface{}) {
	NewEntry(std).withLocation(2).Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func Fatalln(args ...interface{}) {
	NewEntry(std).withLocation(2).Fatalln(args...)
}
