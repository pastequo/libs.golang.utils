// Package logutil provides logger.
package logutil

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var hostname string

func init() {
	hostname = os.Getenv("HOSTNAME")
}

// SetOutput sets the log output.
func SetOutput(file *os.File) {
	logrus.SetOutput(os.Stdout)
}

// Format is the log output format.
type Format int

// All different supported log formatter.
const (
	JSON Format = iota
	Text
)

// SetFormatter sets the formatter of the logs.
func SetFormatter(format Format) {

	switch format {
	case JSON:
		logrus.Printf("Logs format set to JSON (%d).", format)
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case Text:
		logrus.Printf("Logs format set to TEXT (%d).", format)
		logrus.SetFormatter(&logrus.TextFormatter{})
	default:
		logrus.Printf("Logs format set to TEXT (unexpected value: %d).", format)
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
}

// Level is the log level.
type Level int

// All different supported logLevel.
const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

// SetLevel sets the level of the logs.
func SetLevel(level Level) {

	switch level {
	case TraceLevel:
		logrus.Printf("Logs level set to TRACE (%d).", level)
		logrus.SetLevel(logrus.TraceLevel)
	case DebugLevel:
		logrus.Printf("Logs level set to DEBUG (%d).", level)
		logrus.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		logrus.Printf("Logs level set to INFO (%d).", level)
		logrus.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		logrus.Printf("Logs level set to WARN (%d).", level)
		logrus.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		logrus.Printf("Logs level set to ERROR (%d).", level)
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.Printf("Logs level set to WARN (unexpected value: %d).", level)
		logrus.SetLevel(logrus.WarnLevel)
	}
}

var (
	queryCount uint64
	countLock  sync.Mutex
)

// UpdateContext adds the queryIDKey to the context.
func UpdateContext(ctx context.Context, prefix, verb, path string) context.Context {

	if prefix != "" {
		queryID := generateQueryID(prefix)
		ctx = context.WithValue(ctx, queryIDKey, &queryID)
	}

	if verb != "" {
		ctx = context.WithValue(ctx, verbKey, &verb)
	}

	if path != "" {
		ctx = context.WithValue(ctx, pathKey, &path)
	}

	return ctx
}

// GetDefaultLogger returns a logger without any context, except the method calling it.
func GetDefaultLogger() *logrus.Entry {

	logger := newLogger(context.Background())

	return withMethodName(logger, getMethodName())
}

// GetLogger returns a logger with fields containing some Value stored in the context.
func GetLogger(ctx context.Context) *logrus.Entry {

	logger := newLogger(ctx)

	return withMethodName(logger, getMethodName())
}

// GetMethodLogger returns a logger with the method name forced to the given value.
func GetMethodLogger(methodName string) *logrus.Entry {

	logger := newLogger(context.Background())

	return withMethodName(logger, methodName)
}

func newLogger(ctx context.Context) *logrus.Entry {

	fields := logrus.Fields{}

	queryID := getContextValue(ctx, queryIDKey)
	if queryID != nil {
		fields["queryID"] = *queryID
	}

	verb := getContextValue(ctx, verbKey)
	if verb != nil {
		fields["verb"] = *verb
	}

	path := getContextValue(ctx, pathKey)
	if path != nil {
		fields["path"] = *path
	}

	if hostname != "" {
		fields["hostname"] = hostname
	}

	return logrus.WithFields(fields)
}

func withMethodName(logger *logrus.Entry, methodName string) *logrus.Entry {

	if methodName != "" {
		logger = logger.WithField("method", methodName)
	}

	return logger
}

func getMethodName() string {

	// 4 as stack depth should be enough to get the real caller. (2 should be enough)
	stack := make([]uintptr, 4)
	depth := runtime.Callers(3, stack) // Can skip the first 3 as it's Callers < getMethodName < Get(*)Logger

	if depth < 1 {
		return ""
	}

	frames := runtime.CallersFrames(stack)

	for f, hasNext := frames.Next(); hasNext; {

		tmp := strings.Split(f.Function, "/")
		if len(tmp) == 0 {
			continue
		}

		shortName := tmp[len(tmp)-1]

		if !strings.HasPrefix(shortName, "logutil.") {
			return shortName
		}
	}

	return ""
}

func generateQueryID(prefix string) string {

	countLock.Lock()
	defer countLock.Unlock()

	queryID := fmt.Sprintf("%s-%v", prefix, queryCount)
	queryCount++

	return queryID
}

type contextKey int

const (
	queryIDKey contextKey = iota
	verbKey
	pathKey
)

func getContextValue(ctx context.Context, key contextKey) *string {

	if ctx == nil {
		return nil
	}

	val := ctx.Value(key)
	if val == nil {
		return nil
	}

	return val.(*string)
}
