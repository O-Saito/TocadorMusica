package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	initialized bool
)

func InitLogger() error {
	if initialized {
		return nil
	}

	err := os.MkdirAll("logs", 0755)
	if err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02")
	logFile, err := os.OpenFile(
		filepath.Join("logs", "tocadormusica-"+timestamp+".log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	infoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	initialized = true
	Info("Logger initialized")
	return nil
}

func Info(format string, v ...interface{}) {
	if infoLogger == nil {
		fmt.Printf("INFO: "+format+"\n", v...)
		return
	}
	infoLogger.Output(2, fmt.Sprintf(format, v...))
}

func Error(format string, v ...interface{}) {
	if errorLogger == nil {
		fmt.Printf("ERROR: "+format+"\n", v...)
		return
	}
	errorLogger.Output(2, fmt.Sprintf(format, v...))
}
