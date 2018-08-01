package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/gorilla/websocket"
)

func runWebsocket(ws string, decoder *MessageDecoder) {
	log.Println("websocket: connecting")
	c, _, err := websocket.DefaultDialer.Dial(ws, nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	log.Println("websocket: ready")
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			panic(err)
		}
		out, err := decoder.Decode(message)
		if err != nil {
			panic(err)
		}
		go handleMessage(out)
	}
}

func handleMessage(message Message) {
	switch message["type"].(string) {
	case "play":
		url := message["data"].(string)
		log.Println("websocket: play:", url)
		path, err := fetch(url)
		if err != nil {
			log.Println("fetch:", err)
			return
		}
		play(path)
	case "load":
		url := message["data"].(string)
		log.Println("websocket: load:", url)
		_, err := fetch(url)
		if err != nil {
			log.Println("fetch:", err)
			return
		}
	default:
		log.Println("websocket: unknown message type:", message["type"].(string))
	}
}

func sha256String(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func fetch(url string) (string, error) {
	cachePath := filepath.Join("./cache", sha256String(url))
	_, err := os.Stat(cachePath)
	if err == nil {
		return cachePath, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	log.Println("fetch:", url)
	tmpfile, err := ioutil.TempFile("./cache", "")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("http: %d", resp.StatusCode))
	}

	_, err = io.Copy(tmpfile, resp.Body)
	if err != nil {
		return "", nil
	}

	if err = os.Rename(tmpfile.Name(), cachePath); err != nil {
		return "", nil
	}

	return cachePath, nil
}

func play(path string) {
	log.Println("play:", path)

	var player string
	switch runtime.GOOS {
	case "darwin":
		player = "/usr/bin/afplay"
	case "linux":
		player = "/usr/bin/omxplayer"
	default:
		panic("Unknown GOOS")
	}

	exec.Command(player, path).Run()
}
