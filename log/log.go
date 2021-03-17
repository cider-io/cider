package log

import (
	"cider/config"
	"log"
	"os"
)

var Logger *log.Logger

// logging locations
const ToStdout = true

func Error(txt string) {
	// if err != nil {
	// 	Logger.SetPrefix("ERROR ")
	// 	callDepth := 2
	// 	Logger.Output(callDepth, err.Error())
	// 	os.Exit(1)
	// }
	Logger.SetPrefix("ERROR ")
	callDepth := 2
	Logger.Output(callDepth, txt)
}

func Warning(txt string) {
	if config.LogLevel < 2 {
		return
	}
	Logger.SetPrefix("WARNING ")
	callDepth := 2
	Logger.Output(callDepth, txt)
}

func Info(txt string) {
	if config.LogLevel < 3 {
		return
	}
	Logger.SetPrefix("INFO ")
	callDepth := 2
	Logger.Output(callDepth, txt)
}

func Debug(txt string) {
	if config.LogLevel < 4 {
		return
	}
	Logger.SetPrefix("DEBUG ")
	callDepth := 2
	Logger.Output(callDepth, txt)
}

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

	prefix := "INFO " // all non-error handling messages are info
	flags := log.Lmicroseconds | log.Lshortfile
	Logger = log.New(logFile, prefix, flags)
}
