package config

import (
	"time"
)

// package log
const LoggingLevel = 3
const LogStdout = false
const LogFile = "cider.log"

// package gossip
const GossipPort = 7000
const HeartbeatRate = 1 * time.Second
const InitialTFail = 1 * time.Minute

// package api
const ApiPort = 6143
const NonceLength = 32
