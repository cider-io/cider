package main

import (
	"cider/sysinfo"
	"fmt"
)

func main() {
	sysinfo := sysinfo.SysInfo()
	fmt.Println(sysinfo)
}
