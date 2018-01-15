package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"
)

var (
	Version      string
	BuildVersion string
)

var versionFlag = flag.Bool("v", false, "prints current version")

type Program func() int

func init() { flag.Parse() }

func printVersion() int {
	fmt.Printf("version=%s build=%s runtime=%s/%s\n", Version, BuildVersion, runtime.GOOS, runtime.GOARCH)
	return 0
}

func main() {
	var prog Program

	if *versionFlag {
		prog = printVersion
	} else {
		prog = main2
	}

	os.Exit(prog())
}

func main2() int {
	go checkForUpdates(Version, BuildVersion)
	go runHttpService()
	go runBluetoothService()

	select {}

	return 0
}

func checkForUpdates(version, build string) {
	for range time.Tick(1 * time.Second) {
		// log.Println("checking for update")
		// log.Println(runtime.GOOS, runtime.GOARCH)
	}
}
