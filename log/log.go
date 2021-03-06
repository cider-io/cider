package log

import (
	"cider/config"
	"fmt"
	"log"
	"os"
)

var logger *log.Logger

func Output(prefix string, callDepth int, message string) {
	logger.SetPrefix(prefix)
	logger.Output(callDepth, message)
}

// Error: Log an error
func Error(a ...interface{}) {
	if config.LoggingLevel >= 1 {
		Output("ERROR ", 3, fmt.Sprintln(a...))
	}
}

// Warning: Log a warning
func Warning(a ...interface{}) {
	if config.LoggingLevel >= 2 {
		Output("WARNING ", 3, fmt.Sprintln(a...))
	}
}

// Info: Log info
func Info(a ...interface{}) {
	if config.LoggingLevel >= 3 {
		Output("INFO ", 3, fmt.Sprintln(a...))
	}
}

// Debug: Log a debugging message
func Debug(a ...interface{}) {
	if config.LoggingLevel >= 4 {
		Output("DEBUG ", 3, fmt.Sprintln(a...))
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
