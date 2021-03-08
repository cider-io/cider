package main

import (
	"cider/debug"
	"os"
	"errors"
)

func main() {
	debug.InitLogger(os.Stdout)
	debug.Logger.Println("First log :)")
	debug.HandleError(nil)
	debug.HandleError(errors.New("oops it broke"))
}
