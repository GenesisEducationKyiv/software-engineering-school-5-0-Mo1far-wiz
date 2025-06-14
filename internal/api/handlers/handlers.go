package handlers

import "log"

func logErrorF(err error, message string) {
	log.Printf("ERROR: %s: %v", message, err)
}
