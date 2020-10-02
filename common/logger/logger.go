package logger

import (
	"io"
	"log"
	"os"
	"strings"
)

type LogLevel int

const (
	ERROR LogLevel = iota
	WARN
	INFO
	DEBUG
	TRACE
)

var (
	nullWriter = &NullWriter{}
	Info       *log.Logger
	Warn       *log.Logger
	Error      *log.Logger
	Debug      *log.Logger
	Trace      *log.Logger
)

func StringToLogLevel(value string) LogLevel {
	switch strings.ToLower(value) {
	case "error":
		return ERROR
	case "warn":
		return WARN
	case "info":
		return INFO
	case "debug":
		return DEBUG
	case "trace":
		return TRACE
	}
	log.Printf("Invalid log level: '%s'. Returning INFO", value)
	return INFO
}

func (s LogLevel) String() string {
	switch s {
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	}
	return "UNKNOWN"
}

type NullWriter struct {
	io.Writer
}

func (s *NullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func init() {
	Error = log.New(nullWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(nullWriter, "WARN:  ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(nullWriter, "INFO:  ", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(nullWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	Trace = log.New(nullWriter, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Initialize(logLevel LogLevel) {
	log.Printf("Initialize loggers: '%s'", logLevel.String())

	if logLevel >= ERROR {
		Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
	if logLevel >= WARN {
		Warn = log.New(os.Stdout, "WARN:  ", log.Ldate|log.Ltime|log.Lshortfile)
	}
	if logLevel >= INFO {
		Info = log.New(os.Stdout, "INFO:  ", log.Ldate|log.Ltime|log.Lshortfile)
	}
	if logLevel >= DEBUG {
		Debug = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
	if logLevel >= TRACE {
		Trace = log.New(os.Stdout, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
}
