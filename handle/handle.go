package handle

import (
	"cider/log"
	"os"
)

// Error: One-line error handler
func Error(err error) {
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
} 

// Warning: One-line warning handler
func Warning(err error) {
	if err != nil {
		log.Warning(err.Error())
	}
}
