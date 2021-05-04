package gossip

import (
	"cider/handle"
	"cider/log"
	"cider/sysinfo"
	"encoding/json"
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
	prettyPrintedJson, err := json.MarshalIndent(node, prefix, indent)
	handle.Fatal(err)
	log.Info(message, string(prettyPrintedJson))
}

// Get the local node profile during init
func gatherSystemInfo() {
	sysinfo := sysinfo.SysInfo()
	log.Info(sysinfo)
	node := Self.MembershipList[Self.IpAddress]
	node.NodeProfile.Cores, _ = strconv.Atoi(sysinfo["ncpu"])
	node.NodeProfile.Ram, _ = strconv.Atoi(sysinfo["totalMemory"])
	node.NodeProfile.Load = 0

	// TODO: (potential) Update to get reputation from persistent storage
	node.NodeProfile.Reputation = 0

	Self.MembershipList[Self.IpAddress] = node
}
