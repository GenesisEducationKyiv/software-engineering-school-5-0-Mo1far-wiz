package handlers

import "log"

func logError(err error, message string) {
	log.Printf("ERROR: %s: %v", message, err)
}
