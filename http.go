package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func runHttpService() {
	log.Println("http: starting")
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "serial=%s version=%s build=%s\n", SerialNumber(), Version, BuildVersion)
	})

	for {
		log.Println("http: listening on :6969...")
		err := http.ListenAndServe(":6969", nil)
		if err != nil {
			log.Println(err)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
