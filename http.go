package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func runHttpService() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "version=%s build=%s\n", Version, BuildVersion)
	})

	for {
		err := http.ListenAndServe(":6969", nil)
		log.Println(err)
		time.Sleep(1 * time.Second)
	}
}
