package config

import (
	"time"
)

// package gossip
const LogLevel = 4
const LogStdout = true
const LogFile = "cider.log"
const GossipPort = 7000
const HeartbeatRate = 1 * time.Second
