package sysinfo

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// SysInfo returns system specific info, memory in bytes
func SysInfo() map[string]string {

	sysinfo := make(map[string]string)
	multFactor := make(map[byte]int)

	multFactor['G'] = 1024 * 1024 * 1024
	multFactor['M'] = 1024 * 1024
	multFactor['K'] = 1024

	os := runtime.GOOS
	arch := runtime.GOARCH

	cmd := ""
	if os == "darwin" {
		cmd = "top -l 1 -s 0 | grep PhysMem"
	} else if os == "linux" {
		cmd = "free -b"
	} else {
		cmd = "mem"
	}

	memOut, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}

	memory := string(memOut[:])
	if os == "darwin" {
		fields := strings.Fields(memory)
		totalMemory, err := strconv.Atoi(fields[1][:len(fields[1])-1])
		if err != nil {
			fmt.Printf("%s", err)
		}
		totalMemory = totalMemory * multFactor[fields[1][len(fields[1])-1]]

		freeMemory, err := strconv.Atoi(fields[5][:len(fields[5])-1])
		if err != nil {
			fmt.Printf("%s", err)
		}
		freeMemory = freeMemory * multFactor[fields[5][len(fields[5])-1]]

		sysinfo["totalMemory"] = strconv.Itoa(totalMemory)
		sysinfo["freeMemory"] = strconv.Itoa(freeMemory)

	} else if os == "linux" {
		fields := strings.Fields(strings.Split(memory, "\n")[1])
		totalMemory, err := strconv.Atoi(fields[1])
		if err != nil {
			fmt.Printf("%s", err)
		}
		freeMemory, err := strconv.Atoi(fields[3])
		if err != nil {
			fmt.Printf("%s", err)
		}

		sysinfo["totalMemory"] = strconv.Itoa(totalMemory)
		sysinfo["freeMemory"] = strconv.Itoa(freeMemory)

	} else {
		// Windows
		// TODO: Implement parsing for windows
	}

	sysinfo["os"] = os
	sysinfo["arch"] = arch
	sysinfo["ncpu"] = fmt.Sprintf("%d", runtime.NumCPU())

	return sysinfo
}
