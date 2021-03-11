package main

import (
	"cider/sysinfo"
	"cider/log"
	"cider/gossip"
)

func main() {
	log.InitLogger(log.ToStdout)
	sysinfo := sysinfo.SysInfo()
	log.Logger.Println(sysinfo)
	gossip.Start()
}
