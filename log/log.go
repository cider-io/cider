package log

import (
	"cider/config"
	"log"
	"os"
)

var logger *log.Logger

func output(prefix string, callDepth int, message string) {
	logger.SetPrefix(prefix)
	logger.Output(callDepth, message)
}

// Error: Log an error
func Error(message string) {
	output("ERROR ", 2, message)
}

// Warning: Log a warning
func Warning(message string) {
	if config.LoggingLevel >= 2 {
		output("WARNING ", 2, message)
	}	
}

// Info: Log info
func Info(message string) {
	if config.LoggingLevel >= 3 {
		output("INFO ", 2, message)
	}
}

// Debug: Log a debugging message
func Debug(message string) {
	if config.LoggingLevel >= 4 {
		output("DEBUG ", 2, message)
	}
}

// init: Initialize the logger
func init() {
	var logFile *os.File
	var err error

	if config.LogStdout {
		logFile = os.Stdout
	} else {
		// create the log file if it doesn't exist, if it it does exist clear (truncate) it
		logFile, err = os.OpenFile(config.LogFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	prefix := ""
	flags := log.Lmicroseconds | log.Lshortfile
	logger = log.New(logFile, prefix, flags)
}
