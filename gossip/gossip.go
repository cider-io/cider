package gossip

import (
	"cider/log"
	"net"
	"regexp"
	"errors"
)

type Member struct {
	Version int
	Failed bool
}

var group map[string]Member // maps member IP address to Member struct
var MyIp string

func Start() {
	var err error
	log.Logger.Println("Starting gossip")

	// initial membership list
	group := make (map[string]Member)
	MyIp, err = getMyIp()
	log.HandleError(log.Error, err)
	group[MyIp] = Member{Version: 0}

	// TODO: add introducer to the membership list after adding the introducer cli arg 
	log.Logger.Println("Inital membership list:", group)
}

func getMyIp() (string, error) {
	interfaces, err := net.Interfaces()
	log.HandleError(log.Error, err)

	// TODO: this only covers WLANs, we'd also want to support wired LANs
	wifiPattern, err := regexp.Compile("(?i:.*(wifi|wi-fi).*)")
	log.HandleError(log.Error, err)

	for _, iface := range interfaces {
		interfaceIsUp := net.FlagUp & iface.Flags == net.FlagUp 
		interfaceIsWifi := wifiPattern.MatchString(iface.Name)

		if interfaceIsUp && interfaceIsWifi {
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
	return "", errors.New("Your device is not connected to the WLAN.")
}
