package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var serialNumber string
var once sync.Once

func SerialNumber() string {
	once.Do(getSerialNumber)
	return serialNumber
}

func getSerialNumber() {
	cmd := exec.Command("/usr/sbin/ioreg", "-c", "IOPlatformExpertDevice", "-d", "2")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	re := regexp.MustCompile(`(?m)"IOPlatformSerialNumber"\s+=\s+\"(.+?)\"`)
	serialNumber = fmt.Sprintf("%016s", strings.ToLower(string(re.FindSubmatch(out.Bytes())[1])))
}
