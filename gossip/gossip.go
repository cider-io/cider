package gossip

import (
	"cider/log"
	"cider/util"
	"cider/config"
	"net"
	"encoding/gob"
	"sync"
	"time"
	"strconv"
)

type Member struct { // membership list entry
	Version int
	Heartbeat int
	LastUpdated time.Time
	Failed bool
}

type Node struct {
	IpAddress string
	MembershipList map[string]Member // maps member IP address to Member struct
	Infected bool
}

var Self Node

// heartbeat: Send the local membership list over UDP
func heartbeat(neighborIp string) {
	// update my heartbeat
	me := Self.MembershipList[Self.IpAddress]
	me.Heartbeat++
	Self.MembershipList[Self.IpAddress] = me

	connection, err := net.Dial("udp", neighborIp + ":" + strconv.Itoa(config.GossipPort))
	log.HandleError(log.Error, err)
	encoder := gob.NewEncoder(connection)
	encoder.Encode(Self.MembershipList)
}

// listenForGossip: Report incoming gossip to updateMembershipList
func listenForGossip() {
	udpAddress := net.UDPAddr{IP: net.ParseIP(Self.IpAddress), Port: config.GossipPort, Zone: ""}
	udpConnection, err := net.ListenUDP("udp", &udpAddress)
	log.HandleError(log.Error, err)
	for {
		var neighborsMembershipList map[string]Member
		decoder := gob.NewDecoder(udpConnection)
		err = decoder.Decode(&neighborsMembershipList)
		log.HandleError(log.Warning, err)
		updateMembershipList(neighborsMembershipList)
	}
}

// TODO: gossip:
func gossip() {
	startTime := time.Now()
	for {
		if time.Since(startTime) > config.HeartbeatRate {
			heartbeat(Self.IpAddress)
			startTime = time.Now()
		}
	}
}

// TODO: updateMembershipList:
func updateMembershipList(neighborsMembershipList map[string]Member) {
	for ip, member := range neighborsMembershipList {
		prettyPrintMember(ip, member)
	}
}

// Start: Run gossip for group membership and failure detection
func Start() {
	var wg sync.WaitGroup
	wg.Add(1)

	// initialize node
	ipAddress, err := util.GetIpAddress()
	log.HandleError(log.Error, err)
	membershipList := make(map[string]Member)
	// TODO: add introducer to the membership list after adding the introducer cli arg 
	membershipList[ipAddress] = Member{Version: 0, Heartbeat: 0, LastUpdated: time.Now(), Failed: false}
	Self = Node{IpAddress: ipAddress, MembershipList: membershipList, Infected: false}
	
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

