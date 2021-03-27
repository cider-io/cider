package handle

import (
	"cider/log"
	"os"
)

// Fatal: One-line error checker
func Fatal(err error) {
	if err != nil {
		log.Output("FATAL", 3, err.Error())
		os.Exit(1)
	}
} 

// Warning: One-line warning checker
func Warning(err error) {
	if err != nil {
		log.Output("WARNING", 3, err.Error())
	}
}
