package middleware

import (
	"log"
	"os"
	"runtime"
	"time"
)

var errorLogger *log.Logger

func InitLogger() {
	logFile, err := os.OpenFile("./log/error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	errorLogger = log.New(logFile, "", log.LstdFlags)
}

func LogError(err error, context string) {
	if err == nil {
		return
	}
	_, file, line, _ := runtime.Caller(1)

	errorLogger.Printf("[%s] ERROR: %s | Context: %s | File: %s:%d\n",
		time.Now().Format(time.RFC3339), err.Error(), context, file, line)
}