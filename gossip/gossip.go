package gossip

import (
	"cider/config"
	"cider/exportapi"
	"cider/handle"
	"cider/log"
	"cider/util"
	"cider/sysinfo"
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

type Profile struct { // member profile
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
	TFail          time.Duration // FIXME 
	TRemove        time.Duration
}

var Self Node

// heartbeat: Send the local membership list over UDP
func heartbeat() {
	// update my heartbeat
	me := Self.MembershipList[Self.IpAddress]
	me.Heartbeat++
	me.LastUpdated = time.Now()
	me.Profile.Load = exportapi.Load
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
		if ip != Self.IpAddress {
			localVal, ok := Self.MembershipList[ip]
			if (ok && !localVal.Failed && member.Heartbeat > localVal.Heartbeat) || !ok {
				member.LastUpdated = time.Now()
				Self.MembershipList[ip] = member
			}
		}
		prettyPrintMember(ip, Self.MembershipList[ip])
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

	// remove failed nodes that have exceed the removal timeout
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
		log.Warning("This node has no one it's membership list. Are you connected to the network?")
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

	// intialize node profile
	sysInfo := sysinfo.GetSysInfo()
	log.Info(sysInfo)
	profile := Profile{Load: 0, Cores: sysInfo.AvailableCores, Ram: sysInfo.TotalMemory}
	// profile.Load is updated when we heartbeat

	// initialize node
	ipAddress, err := util.GetIpAddress()
	handle.Fatal(err)
	membershipList := make(map[string]Member)
	membershipList[ipAddress] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false, Profile: profile}

	// TODO: We would probably want to have larger TFail and TRemove in the begining to allow for init.
	Self = Node{IpAddress: ipAddress, MembershipList: membershipList, TFail: config.InitialTFail, TRemove: 2 * config.InitialTFail}
	prettyPrintNode("Initial node configuration: ", Self)

	// add introducer to the membership list 
	introducer := flag.String("introducer", "sp21-cs525-g17-01.cs.illinois.edu", "Introducer's hostname or IP address")
	flag.Parse()

	// resolve the introducer's hostname/ip to an ipv4 address
	introducerIps, err := net.LookupIP(*introducer)
	handle.Fatal(err)
	introducerIp := introducerIps[0].To4().String()

	if introducerIp != ipAddress {
		membershipList[introducerIp] = Member{Heartbeat: 0, LastUpdated: time.Now(), Failed: false}
		log.Info("Added introducer", introducerIp,  "to the membershipList list.")
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
