package util

import (
	"cider/log"
	"errors"
	"net"
	"regexp"
)

// GetIpAddress: Return this device's IP address on the WLAN
func GetIpAddress() (string, error) {
	interfaces, err := net.Interfaces()
	log.HandleLog(log.Error, err)

	// FIXME: ethernet usually shows up before wifi
	// user should be able to configure which interface they want CIDER client to run on
	lanPattern, err := regexp.Compile("(?i:.*(wifi|wi-fi|eth|en|utun).*)")
	log.HandleLog(log.Error, err)

	for _, iface := range interfaces {
		interfaceIsUp := net.FlagUp&iface.Flags == net.FlagUp
		interfaceIsLan := lanPattern.MatchString(iface.Name)

		if interfaceIsUp && interfaceIsLan {
			unicastAddresses, err := iface.Addrs()
			log.HandleLog(log.Error, err)

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
