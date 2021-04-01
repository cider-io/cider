package main

import (
	"cider/api"
	// "cider/gossip"
	"cider/log"
	"cider/sysinfo"
	"sync"
)

func main() {
	sysinfo := sysinfo.SysInfo()
	log.Info(sysinfo)

	var wg sync.WaitGroup
	wg.Add(1)

	// Commented this out for testing HTTP API
	// go func() {
	// 	gossip.Start()
	// 	wg.Done()
	// }()

	go func() {
		api.Start()
		wg.Done()
	}()

	wg.Wait()
}
