package gossip

import (
	"cider/config"
	"cider/log"
	"cider/util"
	"encoding/gob"
	"math"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

type Member struct { // membership list entry
	Heartbeat   int
	LastUpdated time.Time
	Failed      bool
}

type Node struct {
	IpAddress      string
	MembershipList map[string]Member // maps member IP address to Member struct
}

var Self Node

// heartbeat: Send the local membership list over UDP
func heartbeat() {
	// update my heartbeat
	me := Self.MembershipList[Self.IpAddress]
	me.Heartbeat++
	me.LastUpdated = time.Now()
	Self.MembershipList[Self.IpAddress] = me

	numGossipNodes := math.Max(math.Round(math.Log2(float64(len(Self.MembershipList)))), 1)
	keys := make([]string, 0, len(Self.MembershipList))
	for k := range Self.MembershipList {
		if k != Self.IpAddress {
			keys = append(keys, k)
		}
	}

	if len(keys) > 0 {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
		for i := 0; i < int(numGossipNodes); i++ {
			connection, err := net.Dial("udp", keys[i]+":"+strconv.Itoa(config.GossipPort))
			if err != nil {
				log.HandleLog(log.Error, err)
			} else {
				encoder := gob.NewEncoder(connection)
				encoder.Encode(Self.MembershipList)
			}
			connection.Close()
		}
	}
}

// listenForGossip: Report incoming gossip to updateMembershipList
func listenForGossip() {
	udpAddress := net.UDPAddr{IP: net.ParseIP(Self.IpAddress), Port: config.GossipPort, Zone: ""}
	udpConnection, err := net.ListenUDP("udp", &udpAddress)
	log.HandleLog(log.Error, err)
	for {
		var neighborsMembershipList map[string]Member
		decoder := gob.NewDecoder(udpConnection)
		err = decoder.Decode(&neighborsMembershipList)
		log.HandleLog(log.Warning, err)
		updateMembershipList(neighborsMembershipList)
	}
}

// gossip: Gossips the local membership list to N random neighbors
// 				N = log2(TOTAL_CLUSTER_NODES)
func gossip() {
	startTime := time.Now()
	for {
		if time.Since(startTime) > config.HeartbeatRate {
			heartbeat()
			startTime = time.Now()
		}
	}
}

// TODO: updateMembershipList:
func updateMembershipList(neighborsMembershipList map[string]Member) {
	for ip, member := range neighborsMembershipList {
		resolvedIps, err := net.LookupIP(ip)
		if err != nil {
			log.HandleLog(log.Error, err)
		}
		resolvedIp := resolvedIps[0].To4().String()
		if resolvedIp != Self.IpAddress {
			localVal, ok := Self.MembershipList[resolvedIp]
			if (ok && member.Heartbeat > localVal.Heartbeat) || !ok {
				member.LastUpdated = time.Now()
				Self.MembershipList[resolvedIp] = member
			}
		}
		prettyPrintMember(resolvedIp, Self.MembershipList[resolvedIp])
	}
}

// Start: Run gossip for group membership and failure detection
func Start() {
	var wg sync.WaitGroup
	wg.Add(1)

	// initialize node
	ipAddress, err := util.GetIpAddress()
	log.HandleLog(log.Error, err)
	membershipList := make(map[string]Member)
	// TODO: add introducer to the membership list after adding the introducer cli arg
	membershipList[ipAddress] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}

	// For Testing purposes only. Remove when we have a robust way of introducing nodes.
	//
	// membershipList["sp21-cs525-g17-01.cs.illinois.edu"] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}
	// membershipList["sp21-cs525-g17-02.cs.illinois.edu"] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}

	Self = Node{IpAddress: ipAddress, MembershipList: membershipList}

	prettyPrintNode("Initial node configuration:", Self)

	log.Logger.Println("Starting gossip")

	go func() {
		listenForGossip()
		wg.Done()
	}()

	go func() {
		gossip()
		wg.Done()
	}()

	wg.Wait()
	log.Logger.Fatalln("Gossip has exited")
}
