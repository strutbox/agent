package main

import (
	"log"
	"os/exec"

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
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("websocket: panic:", r)
				}
			}()
			handleMessage(c, out)
		}()
	}
}

func handleMessage(c *websocket.Conn, message Message) {
	switch message["type"].(string) {
	case "play":
		url := message["data"].(string)
		log.Println("websocket: play:", url)
		path, err := cacheManager.get(url)
		if err != nil {
			log.Println("fetch:", err)
			return
		}
		play(path)
	case "load":
		url := message["data"].(string)
		log.Println("websocket: load:", url)
		_, err := cacheManager.get(url)
		if err != nil {
			log.Println("fetch:", err)
			return
		}
	case "ping":
		log.Println("websocket: ping")
		c.WriteMessage(websocket.BinaryMessage, BuildMessage(map[string]interface{}{"type": "pong"}))
	default:
		log.Println("websocket: unknown message type:", message["type"].(string))
	}
}

func play(path string) {
	log.Println("play:", path)
	if err := exec.Command(playerBin, path).Run(); err != nil {
		log.Println("play:", err)
	}
}
