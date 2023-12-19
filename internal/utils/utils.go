package utils

import (
	"fmt"
	"log"
	"time"
)

func LogInfo(message string) {
	logMessage("INFO", message)
}

func LogError(message string) {
	logMessage("ERROR", message)
}

func logMessage(level, message string) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	log.Println(fmt.Sprintf("[%s] [%s] %s", currentTime, level, message))
}
