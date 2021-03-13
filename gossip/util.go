package gossip

import (
	"strconv"
	"time"
	"cider/log"
	"encoding/json"
)

// prettyPrintMember: Pretty print a membership list entry
func prettyPrintMember(ip string, member Member) {
	summary := "[" + ip + "-" + strconv.Itoa(member.Version) + "]"
	summary += " [♥:" + strconv.Itoa(member.Heartbeat) + "]"
	summary += " [Last updated " + strconv.FormatInt(time.Since(member.LastUpdated).Milliseconds(), 10) + " ago]"
	if member.Failed {
		summary += " [FAILED]"
	}
	log.Logger.Println(summary)
}

// prettyPrintNode: Pretty print a node
func prettyPrintNode (message string, node Node) {
	prefix := "----    "
	indent := "  "
	output, err := json.MarshalIndent(node, prefix, indent)
	log.HandleError(log.Error, err)
	log.Logger.Println(message, string(output))
}
