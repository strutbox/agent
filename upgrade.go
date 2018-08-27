package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func downloadFile(url, dest string, mode os.FileMode) error {
	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func startUpgradeFromURL(url string) {
	dest, _ := os.Executable()
	info, _ := os.Stat(dest)

	os.Remove(dest + ".tmp")

	if err := downloadFile(url, dest+".tmp", info.Mode()); err != nil {
		log.Println(err)
		return
	}

	if err := exec.Command(dest+".tmp", "-v").Run(); err != nil {
		log.Println(err)
		return
	}

	if err := os.Rename(dest+".tmp", dest); err != nil {
		log.Println(err)
		return
	}

	log.Println("successful upgrade, shutting down")
	os.Exit(0)
}
