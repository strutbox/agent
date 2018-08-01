package main

import (
	"bufio"
	"os"
	"regexp"
	"sync"
)

var serialNumber string
var once sync.Once

func SerialNumber() string {
	once.Do(getSerialNumber)
	return serialNumber
}

func getSerialNumber() {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	re := regexp.MustCompile(`(?i)^serial\s+:\s+([a-zA-Z0-9]+)`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if m := re.FindSubmatch(scanner.Bytes()); len(m) > 0 {
			serialNumber = string(m[1])
			return
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	panic("Cannot find serial number for device")
}
