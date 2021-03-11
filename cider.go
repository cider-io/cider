package main

import (
	"cider/sysinfo"
	"cider/log"
)

func main() {
	log.InitLogger(log.ToCiderLog)
	sysinfo := sysinfo.SysInfo()
	log.Logger.Println(sysinfo)
}
