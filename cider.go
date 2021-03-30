package main

import (
	"cider/gossip"
	"cider/log"
	"cider/sysinfo"
	"cider/api"
	"sync"
)

func main() {
	sysinfo := sysinfo.SysInfo()
	log.Info(sysinfo)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		gossip.Start()
		wg.Done()
	}()

	go func() {
		api.Start()
		wg.Done()
	}()

	wg.Wait()
}
