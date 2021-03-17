package log

import (
	"cider/config"
	"log"
	"os"
)

var logger *log.Logger

func Handle(prefix string, callDepth int, txt string) {
	logger.SetPrefix(prefix)
	logger.Output(callDepth, txt)
}

func Error(txt string) {
	Handle("ERROR ", 2, txt)
}

func Fatal(txt string) {
	Error(txt)
	os.Exit(1)
}

func Panic(txt string) {
	Error(txt)
	panic(txt)
}

func Warning(txt string) {
	if config.LogLevel < 2 {
		return
	}
	Handle("WARNING ", 2, txt)
}

func Info(txt string) {
	if config.LogLevel < 3 {
		return
	}
	Handle("INFO ", 2, txt)
}

func Debug(txt string) {
	if config.LogLevel < 4 {
		return
	}
	Handle("DEBUG ", 2, txt)
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
	logger = log.New(logFile, prefix, flags)
}
