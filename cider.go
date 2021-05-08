package main

import (
	"cider/api"
	"cider/gossip"
	"sync"
)

func main() {
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
