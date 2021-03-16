package log

import (
	"log"
	"os"
)

var Logger *log.Logger

// error logging levels
const Error = "ERROR"
const Warning = "WARNING"
const Debug = "DEBUG"

// logging locations
const ToStdout = true
const ToCiderLog = false

func HandleLog(loggingLevel string, err error) {
	if err != nil {
		Logger.SetPrefix(loggingLevel + " ")

		callDepth := 2 // print __FILE__:__LINE__ of the caller (otherwise it would print log.go:18 which is useless)
		Logger.Output(callDepth, err.Error())
		if loggingLevel == Error {
			os.Exit(1)
		}
	}
}

func InitLogger(toStdout bool) {
	var logFile *os.File
	var err error

	if toStdout {
		logFile = os.Stdout
	} else {
		// create the log file if it doesn't exist, if it it does exist clear (truncate) it
		logFile, err = os.OpenFile("cider.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	prefix := "INFO " // all non-error handling messages are info
	flags := log.Lmicroseconds | log.Lshortfile
	Logger = log.New(logFile, prefix, flags)
}
