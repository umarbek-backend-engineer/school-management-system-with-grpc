package utils

import (
	"fmt"
	"log"
	"os"
)

// Simple Error handler so that I do not have to do it manually evrytime there is an error check, 
// just call this function and it will print the error in terminal and send the  error to the client
func ErrorHandler(err error, message string) error {
	errorLogger := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger.Println(message, err)
	return fmt.Errorf("%s: %w", message, err)
}
