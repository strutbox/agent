package main

import (
	"log"
	"syscall"
)

func handleHealthFailure() {
	log.Println("health: unable to connect to service, rebooting")
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}
