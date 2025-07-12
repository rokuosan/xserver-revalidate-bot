package main

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

func NewLogger(output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}

	return &Logger{
		infoLogger:  log.New(output, "INFO: ", log.LstdFlags),
		errorLogger: log.New(output, "ERROR: ", log.LstdFlags),
	}
}

func (l *Logger) LogRequest(record UserAgentRecord) {
	l.infoLogger.Printf("[%s] %s %s - UA: %s - From: %s",
		record.Timestamp.Format("2006-01-02 15:04:05"),
		record.Method,
		record.RequestURI,
		record.UserAgent,
		record.RemoteAddr)
}

func (l *Logger) LogError(message, context string) {
	l.errorLogger.Printf("%s - Context: %s", message, context)
}

func (l *Logger) LogInfo(message string) {
	l.infoLogger.Println(message)
}
