package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	mu       sync.Mutex
	minLevel = LevelInfo
)

func SetLevel(l Level) {
	mu.Lock()
	minLevel = l
	mu.Unlock()
}

func stamp() string {
	return time.Now().Format("15:04:05")
}

func write(level Level, tag, color, msg string) {
	mu.Lock()
	defer mu.Unlock()
	if level < minLevel {
		return
	}
	os.Stdout.WriteString(fmt.Sprintf("%s%s%s %s%s%-5s%s %s\n",
		gray, stamp(), reset,
		bold, color, tag, reset,
		msg))
}

func Debug(msg string) { write(LevelDebug, "DEBUG", dim, msg) }
func Info(msg string)  { write(LevelInfo, "INFO", cyan, msg) }
func Warn(msg string)  { write(LevelWarn, "WARN", yellow, msg) }
func Error(msg string) { write(LevelError, "ERROR", red, msg) }
func OK(msg string)    { write(LevelInfo, "OK", green, msg) }

func Debugf(format string, a ...any) { Debug(fmt.Sprintf(format, a...)) }
func Infof(format string, a ...any)  { Info(fmt.Sprintf(format, a...)) }
func Warnf(format string, a ...any)  { Warn(fmt.Sprintf(format, a...)) }
func Errorf(format string, a ...any) { Error(fmt.Sprintf(format, a...)) }
func OKf(format string, a ...any)    { OK(fmt.Sprintf(format, a...)) }

func Fatal(msg string) {
	write(LevelError, "FATAL", red, msg)
	os.Exit(1)
}

func Fatalf(format string, a ...any) {
	Fatal(fmt.Sprintf(format, a...))
}

// socks5Writer implements io.Writer, routing to our Info logger.
// Used to create a *log.Logger for things-go/go-socks5.
type socks5Writer struct{}

func (w socks5Writer) Write(p []byte) (int, error) {
	Info(fmt.Sprintf("[socks5] %s", string(p)))
	return len(p), nil
}

// NewSocks5Logger returns a *log.Logger that writes to our logger.
func NewSocks5Logger() *log.Logger {
	return log.New(socks5Writer{}, "", 0)
}
