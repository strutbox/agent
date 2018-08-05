package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"
)

func runHttpService() {
	log.Println("http: starting")
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK\n"))
	})
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "serial=%s version=%s build=%s\n", SerialNumber(), Version, BuildVersion)
	})
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command("systemctl", "status", "strut.service")
		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Run(); err != nil {
			fmt.Fprint(w, err.Error())
		}
		w.Write([]byte{'\n'})
	})
	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command("journalctl", "-u", "strut.service")
		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Run(); err != nil {
			fmt.Fprint(w, err.Error())
		}
		w.Write([]byte{'\n'})
	})
	http.HandleFunc("/cache", func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command("ls", "-lh", cacheManager.dir)
		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Run(); err != nil {
			fmt.Fprint(w, err.Error())
		}
		w.Write([]byte{'\n'})
	})
	http.HandleFunc("/exec", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		cmd := exec.Command("bash", "-c", string(body))
		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Run(); err != nil {
			fmt.Fprint(w, err.Error())
		}
		w.Write([]byte{'\n'})
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
