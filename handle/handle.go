package handle

import (
	"cider/log"
	"os"
)

// Fatal: One-line error handler
func Fatal(err error) {
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
} 

// Warning: One-line warning handler
func Warning(err error) {
	if err != nil {
		log.Warning(err)
	}
}
