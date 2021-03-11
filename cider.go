package main

import (
	"cider/log"
	"errors"
)

func main() {
	log.InitLogger(log.ToCiderLog)
	log.Logger.Println("First log :)")
	log.HandleError(log.Warning, errors.New("This is a warning."))
	log.HandleError(log.Error, errors.New("Oops it broke."))
}
