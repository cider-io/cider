package debug

import (
	"log"
	"os"
)

var Logger *log.Logger

func HandleError(err error) {
	if err != nil {
		// print __FILE__:__LINE__ of the caller (otherwise it would print debug.go:14 which is useless)
		callDepth := 2 
		Logger.Output(callDepth, err.Error())
		os.Exit(1)
	}
}

func InitLogger(logFile *os.File) {
	prefix := "CIDER "
	flags := log.Ldate|log.Ltime|log.Lshortfile
	Logger = log.New(logFile, prefix, flags)
}
