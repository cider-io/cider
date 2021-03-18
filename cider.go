package main

import (
	"cider/gossip"
	"cider/log"
	"cider/sysinfo"
	"fmt"
)

func main() {
	sysinfo := sysinfo.SysInfo()
	log.Info(fmt.Sprint(sysinfo))
	gossip.Start()
}
