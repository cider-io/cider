package gossip

import (
	"cider/handle"
	"cider/log"
	"encoding/json"
	"strconv"
	"time"
)

// prettyPrintMember: Pretty print a membership list entry
func prettyPrintMember(ip string, member Member) {
	summary := "[" + ip + "]"
	summary += " [♥:" + strconv.Itoa(member.Heartbeat) + "]"
	summary += " [Last updated " + strconv.FormatInt(time.Since(member.LastUpdated).Milliseconds(), 10) + "ms ago]"
	if member.Failed {
		summary += " [FAILED]"
	}
	log.Info(summary)
}

// prettyPrintNode: Pretty print a node
func prettyPrintNode(message string, node Node) {
	prefix := "----    "
	indent := "  "
	prettyPrintedJson, err := json.MarshalIndent(node, prefix, indent)
	handle.Fatal(err)
	log.Info(message, string(prettyPrintedJson))
}
