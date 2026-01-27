package logger

import (
	"log"
	"os"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

// Init initializes the logger
func Init() {
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Info logs an info message
func Info(v ...interface{}) {
	infoLogger.Println(v...)
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}

// Error logs an error message
func Error(v ...interface{}) {
	errorLogger.Println(v...)
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}
