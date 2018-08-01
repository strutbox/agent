package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

const DEVICE_NAME = "STRUT"

var ServiceUUID = ble.MustParse("6efa9836-b179-4387-b21a-1b7dffacfae0")

const wpaTemplate = `country=US
ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1

network={
  ssid="%s"
  psk="%s"
}
`
const wpaLocation = "/etc/wpa_supplicant/wpa_supplicant.conf"

func runBluetoothService() {
	log.Println("bluetooth: starting")
	for {
		log.Println("bluetooth: NewDevice", DEVICE_NAME)
		d, err := dev.NewDevice("default")
		if err != nil {
			log.Println(err)
			time.Sleep(250 * time.Millisecond)
			continue
		}
		log.Println("bluetooth: got device")
		ble.SetDefaultDevice(d)

		svc := ble.NewService(ServiceUUID)
		svc.AddCharacteristic(NewSSIDCharacteristic())
		svc.AddCharacteristic(NewIPAddressCharacteristic())

		ble.AddService(svc)

		log.Println("bluetooth: advertising...")
		// ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
		ble.AdvertiseNameAndServices(context.Background(), DEVICE_NAME, svc.UUID)
		// ble.AdvertiseNameAndServices(context.Background(), DEVICE_NAME)
	}
}

func NewSSIDCharacteristic() *ble.Characteristic {
	c := ble.NewCharacteristic(ble.MustParse("6efa9836-b179-4387-b21a-1b7dffacfae1"))
	c.HandleNotify(ble.NotifyHandlerFunc(func(r ble.Request, n ble.Notifier) {
		log.Println("bluetooth: fetching SSIDs")
		cmd := exec.Command("iwlist", "wlan0", "scan")
		out, err := cmd.Output()
		if err != nil {
			log.Println(err)
			return
		}
		scanner := bufio.NewScanner(bytes.NewReader(out))

		var ssid map[string]interface{}

		for scanner.Scan() {
			line := strings.Trim(scanner.Text(), " ")
			if strings.HasPrefix(line, "Cell") {
				if ssid != nil {
					b, _ := json.Marshal(ssid)
					n.Write(b)
				}
				ssid = map[string]interface{}{
					"s": "",
					"e": 0,
				}

			} else if strings.HasPrefix(line, "ESSID:") {
				ssid["s"] = line[7 : len(line)-1]
			} else if strings.HasPrefix(line, "Encryption key:") {
				if line == "Encryption key:on" {
					ssid["e"] = 1
				}
			}
		}
		if ssid != nil {
			b, _ := json.Marshal(ssid)
			n.Write(b)
		}
	}))

	c.HandleWrite(ble.WriteHandlerFunc(func(r ble.Request, w ble.ResponseWriter) {
		conn := map[string]string{}
		json.Unmarshal(r.Data(), &conn)
		fp, _ := os.OpenFile(wpaLocation+".tmp", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		fmt.Fprintf(fp, wpaTemplate, conn["s"], conn["p"])
		fp.Close()
		os.Rename(wpaLocation+".tmp", wpaLocation)

		exec.Command("ifconfig", "wlan0", "down").Run()
		exec.Command("ifconfig", "wlan0", "up").Run()
	}))
	return c
}

func NewIPAddressCharacteristic() *ble.Characteristic {
	c := ble.NewCharacteristic(ble.MustParse("6efa9836-b179-4387-b21a-1b7dffacfae2"))
	c.HandleRead(ble.ReadHandlerFunc(func(r ble.Request, w ble.ResponseWriter) {
		log.Println("bluetooth: reading IP")
		response := map[string]interface{}{
			"ip": nil,
		}
		iface, err := net.InterfaceByName("wlan0")
		if err != nil {
			log.Println("bluetooth:", err)
			return
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Println("bluetooth:", err)
			return
		}
		for _, addr := range addrs {
			switch ip := addr.(type) {
			case *net.IPNet:
				if ip.IP.DefaultMask() != nil {
					response["ip"] = ip.IP.String()
					break
				}
			}
		}
		b, _ := json.Marshal(response)
		w.Write(b)
	}))
	return c
}
