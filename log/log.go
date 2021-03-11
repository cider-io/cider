package log

import (
	"log"
	"os"
)

var Logger *log.Logger

// error logging levels
const Error = "ERROR"
const Warning = "WARNING"

func HandleError(loggingLevel string, err error) {
	if err != nil {
		Logger.SetPrefix(loggingLevel + " ")

		callDepth := 2 // print __FILE__:__LINE__ of the caller (otherwise it would print log.go:18 which is useless)
		Logger.Output(callDepth, err.Error())
		if loggingLevel == Error {
			os.Exit(1)
		}
	}
}

func InitLogger() {
	logFile, err := os.OpenFile("cider.log", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	prefix := "INFO " // all non-error handling messages are info
	flags := log.Lmicroseconds|log.Lrm shortfile
	Logger = log.New(logFile, prefix, flags)
}
