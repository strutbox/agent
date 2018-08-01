package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var privateKey *rsa.PrivateKey

func runAgent() {
	log.Println("agent: starting")
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("agent: panic:", r)
				}
				time.Sleep(1 * time.Second)

			}()
			agentMain()
		}()
	}
}

var certOnce sync.Once

func agentMain() {
	certOnce.Do(func() {
		ensureCertificateExist()
		loadCertificate()
	})

	bootstrap()
	time.Sleep(1 * time.Second)
}

func ensureCertificateExist() {
	if _, err := os.Stat("private.key"); err == nil {
		return
	}

	log.Println("agent: generating new private.key")

	file, err := os.Create("private.key")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	pem.Encode(file, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func loadCertificate() {
	file, err := os.Open("private.key")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" {
		panic("failed to decode private.key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	privateKey = key
}

func bootstrap() {
	publicKeyDer, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}

	pubKey := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDer,
	})

	decoder := &MessageDecoder{privateKey}

	req, err := http.NewRequest("POST", BootstrapHost+"/api/0/agent/bootstrap/", bytes.NewReader(pubKey))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-pem-file")
	req.Header.Set("User-Agent", fmt.Sprintf("Strut/%s-%s (%s)", Version, BuildVersion, SerialNumber()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic("bad response: " + string(body))
	}

	out, err := decoder.Decode(body)
	if err != nil {
		panic(err)
	}

	runWebsocket(out["websocket"].(string), decoder)
}
