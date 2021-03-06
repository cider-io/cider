package gossip

import (
	"cider/config"
	"cider/exportapi"
	"cider/handle"
	"cider/log"
	"cider/sysinfo"
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

type ResourceProfile struct {
	Load       int
	Cores      int
	Ram        int
	Reputation int
}

type Member struct { // membership list entry
	Heartbeat   int
	LastUpdated time.Time
	Failed      bool
	Profile     ResourceProfile
}

type Node struct {
	IpAddress      string
	MembershipList map[string]Member // maps member IP address to Member struct
	FailureTimeout time.Duration
	RemovalTimeout time.Duration
}

var Self Node

// heartbeat: Send the local membership list over UDP
func heartbeat() {
	// update my heartbeat
	me := Self.MembershipList[Self.IpAddress]
	me.Heartbeat++
	me.LastUpdated = time.Now()
	me.Profile.Load = exportapi.Load
	me.Profile.Reputation = exportapi.Reputation
	Self.MembershipList[Self.IpAddress] = me

	// activeMembers doesn't include the node itself
	activeMembers := make([]string, 0, len(Self.MembershipList))
	for ip, member := range Self.MembershipList {
		if ip != Self.IpAddress && !member.Failed {
			activeMembers = append(activeMembers, ip)
		}
	}

	// gossip to a max of log_2(cluster_size) members
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
		if ip != Self.IpAddress {
			localVal, ok := Self.MembershipList[ip]
			if (ok && !localVal.Failed && member.Heartbeat > localVal.Heartbeat) || !ok {
				member.LastUpdated = time.Now()
				Self.MembershipList[ip] = member
			}
		}
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

	// FIXME replace these magic numbers
	Self.FailureTimeout = 5 * time.Duration(numGossipNodes) * time.Second
	Self.RemovalTimeout = 2 * Self.FailureTimeout

	// remove failed nodes that have exceed the removal timeout
	removeList := make([]string, 0, len(Self.MembershipList))
	for ip, member := range Self.MembershipList {
		if ip != Self.IpAddress && !member.Failed && time.Since(member.LastUpdated) > Self.FailureTimeout {
			member.Failed = true
			Self.MembershipList[ip] = member
			log.Debug("Node marked as failed:", ip)
		} else if member.Failed && time.Since(member.LastUpdated) > Self.RemovalTimeout {
			removeList = append(removeList, ip)
		}
	}

	for _, ip := range removeList {
		delete(Self.MembershipList, ip)
		log.Debug("Node removed:", ip)
	}

	if len(removeList) > 0 && len(Self.MembershipList) == 1 {
		log.Warning("Are you connected to the network? We can't find any active CIDER nodes.")
	}
}

// gossip: Gossips the local membership list to at most log2(cluster_size) random neighbors
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

	// CLI flags
	introducer := flag.String("introducer", "sp21-cs525-g17-01.cs.illinois.edu", "Introducer's hostname or IP address")
	resourceConstrained := flag.Bool("resource-constrained", false, "Simulate resource constrained nodes")
	flag.Parse()

	// intialize node profile
	sysInfo := sysinfo.GetSysInfo()
	log.Info(sysInfo)

	// TODO: Reputation needs to be obtained from the persistent storage
	profile := ResourceProfile{Load: 0, Cores: sysInfo.AvailableCores, Ram: sysInfo.TotalMemory, Reputation: 0}
	if *resourceConstrained { // simulate resource-constrained nodes that cannot be used for compute
		profile = ResourceProfile{Load: 0, Cores: 0, Ram: 0, Reputation: 0}
	}
	// profile.Load is updated when we heartbeat

	// initialize node
	ipAddress, err := util.GetIpAddress()
	handle.Fatal(err)
	membershipList := make(map[string]Member)
	membershipList[ipAddress] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false, Profile: profile}

	Self = Node{IpAddress: ipAddress, MembershipList: membershipList, FailureTimeout: config.InitialFailureTimeout, RemovalTimeout: 2 * config.InitialFailureTimeout}
	prettyPrintNode("Initial node configuration: ", Self)

	// resolve the introducer's hostname/ip to an ipv4 address
	introducerIps, err := net.LookupIP(*introducer)
	handle.Fatal(err)
	introducerIp := introducerIps[0].To4().String()

	// add introducer to the membership list
	if introducerIp != ipAddress {
		membershipList[introducerIp] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}
		log.Info("Starting gossip")
	} else {
		log.Info("Starting gossip as the introducer")
	}

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
