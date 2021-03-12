package gossip

import (
	"cider/log"
	"net"
	"regexp"
	"errors"
	"encoding/gob"
	"sync"
	"time"
	"strconv"
)

type Member struct {
	Version int
	Heartbeat int
	LastUpdated time.Time
	Failed bool
}

var MyView map[string]Member // maps member IP address to Member struct
var MyIp string
var infected bool

const gossipPort = 7000
const heartbeatRate = 1 * time.Second

// heartbeat: Send the local membership list over UDP
func heartbeat(neighborIp string) {
	// update my heartbeat
	me := MyView[MyIp]
	me.Heartbeat++
	MyView[MyIp] = me

	connection, err := net.Dial("udp", neighborIp + ":" + strconv.Itoa(gossipPort))
	log.HandleError(log.Error, err)
	encoder := gob.NewEncoder(connection)
	encoder.Encode(MyView)
}

// listenForGossip: Report incoming gossip to updateView
func listenForGossip() {
	udpAddress := net.UDPAddr{IP: net.ParseIP(MyIp), Port: gossipPort, Zone: ""}
	udpConnection, err := net.ListenUDP("udp", &udpAddress)
	log.HandleError(log.Error, err)
	for {
		var neighborsView map[string]Member
		decoder := gob.NewDecoder(udpConnection)
		err = decoder.Decode(&neighborsView)
		log.HandleError(log.Warning, err)
		updateView(neighborsView)
	}
}

// TODO: gossip:
func gossip() {
	startTime := time.Now()
	for {
		if time.Since(startTime) > heartbeatRate {
			heartbeat(MyIp)
			startTime = time.Now()
		}
	}
}

// TODO: updateView:
func updateView(neighborsView map[string]Member) {
	for ip, member := range neighborsView {
		prettyPrintMember(ip, member)
	}
}

// getMyIp: Return this device's IP address on the WLAN
func getMyIp() (string, error) {
	interfaces, err := net.Interfaces()
	log.HandleError(log.Error, err)

	// FIXME: ethernet usually shows up before wifi
	// user should be able to configure which interface they want CIDER client to run on
	lanPattern, err := regexp.Compile("(?i:.*(wifi|wi-fi|eth).*)")
	log.HandleError(log.Error, err)

	for _, iface := range interfaces {
		interfaceIsUp := net.FlagUp & iface.Flags == net.FlagUp 
		interfaceIsLan := lanPattern.MatchString(iface.Name)

		if interfaceIsUp && interfaceIsLan {
			unicastAddresses, err := iface.Addrs()
			log.HandleError(log.Error, err)

			for _, address := range unicastAddresses {
				switch value := address.(type) {
				case *net.IPNet:
					if value.IP.To4() != nil {
						return value.IP.String(), nil
					}
				}
			}
		}
	}
	return "", errors.New("Your device is not connected to the LAN.")
}

func prettyPrintMember(ip string, member Member) {
	summary := "[" + ip + "-" + strconv.Itoa(member.Version) + "]"
	summary += " [♥:" + strconv.Itoa(member.Heartbeat) + "]"
	summary += " [Last updated " + strconv.FormatInt(time.Since(member.LastUpdated).Milliseconds(), 10) + " ago]"
	if member.Failed {
		summary += " [FAILED]"
	}
	log.Logger.Println(summary)
}

// Start: Run gossip for group membership and failure detection
func Start() {
	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	log.Logger.Println("Starting gossip")

	infected = false // the node is initially uninfected

	// initial membership list
	MyView = make(map[string]Member)
	MyIp, err = getMyIp()
	log.HandleError(log.Error, err)
	MyView[MyIp] = Member{Version: 0, Heartbeat: 0, LastUpdated: time.Now(), Failed: false}

	// TODO: add introducer to the membership list after adding the introducer cli arg 
	log.Logger.Println("Inital membership list:", MyView)

	go func() {
		listenForGossip()
		wg.Done()
	}()
	
	go func() {
		gossip()
		wg.Done()
	}()

	wg.Wait()
}

