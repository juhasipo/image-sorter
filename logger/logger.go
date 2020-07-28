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
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
	Debug *log.Logger
	Trace *log.Logger
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

func Initialize(logLevel LogLevel) {
	log.Printf("Initialize loggers: '%s'", logLevel.String())

	nullWriter := &NullWriter{}
	var errorWriter io.Writer = nullWriter
	var warnWriter io.Writer = nullWriter
	var infoWriter io.Writer = nullWriter
	var debugWriter io.Writer = nullWriter
	var traceWriter io.Writer = nullWriter
	if logLevel >= ERROR {
		errorWriter = os.Stderr
	}
	if logLevel >= WARN {
		warnWriter = os.Stdout
	}
	if logLevel >= INFO {
		infoWriter = os.Stdout
	}
	if logLevel >= DEBUG {
		debugWriter = os.Stdout
	}
	if logLevel >= TRACE {
		traceWriter = os.Stdout
	}

	Error = log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(warnWriter, "WARN:  ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoWriter, "INFO:  ", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(debugWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	Trace = log.New(traceWriter, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
}
