package sysinfo

import (
	"cider/handle"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type SysInfo struct {
	Os string
	Arch string
	AvailableCores int
	TotalMemory int
	FreeMemory int
}

// GetSysInfo: returns system specific info, memory in bytes
func GetSysInfo() SysInfo {

	multFactor := map[byte]int{'G': 1024 * 1024 * 1024, 'M': 1024 * 1024, 'K': 1024}

	os := runtime.GOOS
	arch := runtime.GOARCH
	var totalMemory, freeMemory int

	cmd := ""
	if os == "darwin" {
		cmd = "top -l 1 -s 0 | grep PhysMem"
	} else if os == "linux" {
		cmd = "free -b"
	} else {
		cmd = "systeminfo"
	}

	memOut, err := exec.Command("bash", "-c", cmd).Output()
	handle.Warning(err)

	memory := string(memOut[:])
	if os == "darwin" {
		fields := strings.Fields(memory)
		totalMemory, err = strconv.Atoi(fields[1][:len(fields[1])-1])
		handle.Warning(err)
		totalMemory = totalMemory * multFactor[fields[1][len(fields[1])-1]]

		freeMemory, err = strconv.Atoi(fields[5][:len(fields[5])-1])
		handle.Warning(err)
		freeMemory = freeMemory * multFactor[fields[5][len(fields[5])-1]]

	} else if os == "linux" {
		fields := strings.Fields(strings.Split(memory, "\n")[1])
		totalMemory, err = strconv.Atoi(fields[1])
		handle.Warning(err)
		freeMemory, err = strconv.Atoi(fields[3])
		handle.Warning(err)

	} else {
		// Windows

		// We have the following fields that we need to
		//   extract from the mem output
		//
		// Total Physical Memory:     7,168 MB
		// Available Physical Memory: 5,374 MB
		//

		totalMemoryPattern, _ := regexp.Compile("Total Physical Memory: *(?P<amount>.*) (?P<units>.*B)")
		submatches := totalMemoryPattern.FindStringSubmatch(memory)
		amount, _ := strconv.ParseInt(strings.ReplaceAll(submatches[1], ",", ""), 10, 0)
		totalMemory = int(amount) * multFactor[submatches[2][0]]

		availableMemoryPattern, _ := regexp.Compile("Available Physical Memory: *(?P<amount>.*) (?P<units>.*B)")
		submatches = availableMemoryPattern.FindStringSubmatch(memory)
		amount, _ = strconv.ParseInt(strings.ReplaceAll(submatches[1], ",", ""), 10, 0)
		freeMemory = int(amount) * multFactor[submatches[2][0]]
	}

	return SysInfo{Os: os, Arch: arch, TotalMemory: totalMemory, FreeMemory: freeMemory, AvailableCores: runtime.NumCPU()}
}
