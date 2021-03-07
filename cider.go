package main

import (
	"example.com/cider/sysinfo"
	"fmt"
)

func main() {
	sysinfo := sysinfo.SysInfo()
	fmt.Println(sysinfo)
}
