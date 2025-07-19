package main

import (
	"log"
	"pkg/logger"
)

const logFile = "application.log"

func main() {
	logger, err := logger.NewLogger(logFile)
	if err != nil {
		log.Fatal(err)
	}
	logger.ConsoleLogInfo("Logger initialized.")

}
