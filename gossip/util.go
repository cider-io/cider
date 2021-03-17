package gossip

import (
	"cider/log"
	"encoding/json"
	"os"
	"strconv"
	"time"
)

// prettyPrintMember: Pretty print a membership list entry
func prettyPrintMember(ip string, member Member) {
	summary := "[" + ip + "]"
	summary += " [♥:" + strconv.Itoa(member.Heartbeat) + "]"
	summary += " [Last updated " + strconv.FormatInt(time.Since(member.LastUpdated).Milliseconds(), 10) + " ago]"
	if member.Failed {
		summary += " [FAILED]"
	}
	log.Info(summary)
}

// prettyPrintNode: Pretty print a node
func prettyPrintNode(message string, node Node) {
	prefix := "----    "
	indent := "  "
	output, err := json.MarshalIndent(node, prefix, indent)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	log.Info(message + string(output))
}
