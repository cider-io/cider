package gossip

import (
	"cider/config"
	"cider/log"
	"cider/util"
	"cider/handle"
	"encoding/gob"
	"math"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
	"errors"
)

type Member struct { // membership list entry
	Heartbeat   int
	LastUpdated time.Time
	Failed      bool
}

type Node struct {
	IpAddress      string
	MembershipList map[string]Member // maps member IP address to Member struct
	TFail          time.Duration
	TRemove        time.Duration
}

var Self Node

// heartbeat: Send the local membership list over UDP
func heartbeat() {
	// update my heartbeat
	me := Self.MembershipList[Self.IpAddress]
	me.Heartbeat++
	me.LastUpdated = time.Now()
	Self.MembershipList[Self.IpAddress] = me

	keys := make([]string, 0, len(Self.MembershipList))
	for k, val := range Self.MembershipList {
		if k != Self.IpAddress && !val.Failed {
			keys = append(keys, k)
		}
	}

	numGossipNodes := math.Max(math.Round(math.Log2(float64(len(Self.MembershipList)))),
		float64(len(keys)))

	if len(keys) > 0 {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
		for i := 0; i < int(numGossipNodes); i++ {
			connection, err := net.Dial("udp", keys[i]+":"+strconv.Itoa(config.GossipPort))
			handle.Error(err)
			encoder := gob.NewEncoder(connection)
			encoder.Encode(Self.MembershipList)
			connection.Close()
		}
	}
}

// updateMembershipList: Update the membership list based on gossips from neighbors
func updateMembershipList(neighborsMembershipList map[string]Member) {
	for ip, member := range neighborsMembershipList {
		resolvedIps, err := net.LookupIP(ip)
		handle.Error(err)
		resolvedIp := resolvedIps[0].To4().String()
		if resolvedIp != Self.IpAddress {
			localVal, ok := Self.MembershipList[resolvedIp]
			if (ok && !localVal.Failed && member.Heartbeat > localVal.Heartbeat) || !ok {
				member.LastUpdated = time.Now()
				Self.MembershipList[resolvedIp] = member
			}
		}
		prettyPrintMember(resolvedIp, Self.MembershipList[resolvedIp])
	}
}

// listenForGossip: Report incoming gossip to updateMembershipList
func listenForGossip() {
	udpAddress := net.UDPAddr{IP: net.ParseIP(Self.IpAddress), Port: config.GossipPort, Zone: ""}
	udpConnection, err := net.ListenUDP("udp", &udpAddress)
	handle.Error(err)

	for {
		var neighborsMembershipList map[string]Member
		decoder := gob.NewDecoder(udpConnection)
		err = decoder.Decode(&neighborsMembershipList)
		handle.Warning(err)
		updateMembershipList(neighborsMembershipList)
	}
}

func failureDetection() {
	numGossipNodes := math.Max(math.Round(math.Log2(float64(len(Self.MembershipList)))), 1)

	Self.TFail = time.Duration(numGossipNodes) * time.Second
	Self.TRemove = 2 * Self.TFail

	removeList := make([]string, 0, len(Self.MembershipList))
	for ip, member := range Self.MembershipList {
		if ip != Self.IpAddress && !member.Failed && time.Since(member.LastUpdated) > Self.TFail {
			member.Failed = true
			Self.MembershipList[ip] = member
			log.Debug("Node marked as failed: " + ip)
		} else if member.Failed && time.Since(member.LastUpdated) > Self.TRemove {
			removeList = append(removeList, ip)
		}
	}

	for _, ip := range removeList {
		delete(Self.MembershipList, ip)
		log.Debug("Node removed: " + ip)
	}

	if len(removeList) > 0 && len(Self.MembershipList) == 1 {
		handle.Error(errors.New("This node has possibly been marked as " +
			"failed by all other nodes in the cluster. Attempt to restart"))
	}
}

// gossip: Gossips the local membership list to N random neighbors
// 				N = log2(TOTAL_CLUSTER_NODES)
func gossip() {
	startTime := time.Now()
	for {
		if time.Since(startTime) > config.HeartbeatRate {
			failureDetection()
			heartbeat()
			startTime = time.Now()
		}
	}
}

// Start: Run gossip for group membership and failure detection
func Start() {
	var wg sync.WaitGroup
	wg.Add(1)

	// initialize node
	ipAddress, err := util.GetIpAddress()
	handle.Error(err)
	membershipList := make(map[string]Member)
	// TODO: add introducer to the membership list after adding the introducer cli arg
	membershipList[ipAddress] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}

	// For Testing purposes only. Remove when we have a robust way of introducing nodes.
	//
	// membershipList["sp21-cs525-g17-01.cs.illinois.edu"] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}
	// membershipList["sp21-cs525-g17-02.cs.illinois.edu"] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}

	// TODO: We would probably want to have larger TFail and TRemove in the begining to allow for init.
	Self = Node{IpAddress: ipAddress, MembershipList: membershipList, TFail: config.InitialTFail, TRemove: 2 * config.InitialTFail}

	prettyPrintNode("Initial node configuration: ", Self)

	log.Info("Starting gossip")

	go func() {
		listenForGossip()
		wg.Done()
	}()

	go func() {
		gossip()
		wg.Done()
	}()

	wg.Wait()
	handle.Error(errors.New("Gossip has exited"))
}
