package gossip

import (
	"cider/config"
	"cider/exportapi"
	"cider/handle"
	"cider/log"
	"cider/util"
	"encoding/gob"
	"errors"
	"flag"
	"math"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

type Profile struct { // Node profile
	Reputation int
	Load       int
	Cores      int
	Ram        int
}

type Member struct { // membership list entry
	Heartbeat   int
	LastUpdated time.Time
	Failed      bool
	Profile     Profile
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
	me.Profile.Load = exportapi.GetCurrentLoad()
	Self.MembershipList[Self.IpAddress] = me

	// activeMembers doesn't include the node itself
	activeMembers := make([]string, 0, len(Self.MembershipList))
	for ip, member := range Self.MembershipList {
		if ip != Self.IpAddress && !member.Failed {
			activeMembers = append(activeMembers, ip)
		}
	}

	logBase2 := math.Round(math.Log2(float64(len(Self.MembershipList))))
	numActiveMembers := float64(len(activeMembers))
	numGossipNodes := int(math.Min(logBase2, numActiveMembers))

	if len(activeMembers) > 0 {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(activeMembers), func(i, j int) { activeMembers[i], activeMembers[j] = activeMembers[j], activeMembers[i] })
		for i := 0; i < numGossipNodes; i++ {
			connection, err := net.Dial("udp", activeMembers[i]+":"+strconv.Itoa(config.GossipPort))
			handle.Fatal(err)
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
		handle.Fatal(err)
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
	handle.Fatal(err)

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

	Self.TFail = 5 * time.Duration(numGossipNodes) * time.Second
	Self.TRemove = 2 * Self.TFail

	removeList := make([]string, 0, len(Self.MembershipList))
	for ip, member := range Self.MembershipList {
		if ip != Self.IpAddress && !member.Failed && time.Since(member.LastUpdated) > Self.TFail {
			member.Failed = true
			Self.MembershipList[ip] = member
			log.Debug("Node marked as failed:", ip)
		} else if member.Failed && time.Since(member.LastUpdated) > Self.TRemove {
			removeList = append(removeList, ip)
		}
	}

	for _, ip := range removeList {
		delete(Self.MembershipList, ip)
		log.Debug("Node removed:", ip)
	}

	if len(removeList) > 0 && len(Self.MembershipList) == 1 {
		handle.Fatal(errors.New("this node has possibly been marked as " +
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
	handle.Fatal(err)
	membershipList := make(map[string]Member)
	// TODO: add introducer to the membership list after adding the introducer cli arg

	membershipList[ipAddress] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}

	// For Testing purposes only. Remove when we have a robust way of introducing nodes.
	//
	// membershipList["sp21-cs525-g17-01.cs.illinois.edu"] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}
	// membershipList["sp21-cs525-g17-02.cs.illinois.edu"] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}

	// TODO: We would probably want to have larger TFail and TRemove in the begining to allow for init.
	Self = Node{IpAddress: ipAddress, MembershipList: membershipList, TFail: config.InitialTFail, TRemove: 2 * config.InitialTFail}

	gatherSystemInfo()

	prettyPrintNode("Initial node configuration: ", Self)

	introducer := flag.String("introducer", "sp21-cs525-g17-01.cs.illinois.edu", "Introducer's hostname or IP address")
	flag.Parse()

	if *introducer != ipAddress {
		log.Info("Add introducer to the membershipList list: " + *introducer)
		membershipList1 := make(map[string]Member)
		membershipList1[*introducer] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}
		updateMembershipList(membershipList1)
	}
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
	handle.Fatal(errors.New("gossip has exited"))
}
