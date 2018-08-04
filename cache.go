package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const cacheDir = "./cache"

type CacheManager struct {
	dir         string
	downloadUrl string
	maxSize     int64
	m           sync.Mutex
	l           map[string]sync.Mutex
	cl          sync.Mutex
}

var cacheManager *CacheManager

func (c *CacheManager) get(url string) (string, error) {
	c.m.Lock()
	if c.l == nil {
		c.l = make(map[string]sync.Mutex)
	}
	l, ok := c.l[url]
	if !ok {
		c.l[url] = l
	}
	l.Lock()
	defer func() {
		l.Unlock()
		delete(c.l, url)
		c.m.Unlock()
	}()
	return c.download(url)
}

func (c *CacheManager) download(url string) (string, error) {
	cachePath := filepath.Join(cacheDir, url)
	_, err := os.Stat(cachePath)
	if err == nil {
		return cachePath, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	log.Println("cache: downloading", url)
	tmpfile, err := ioutil.TempFile(cacheDir, "")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	resp, err := http.Get(c.downloadUrl + url)
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

	go c.clean()
	return cachePath, nil
}

func (c *CacheManager) clean() {
	// Clean up cached files in a sorta LRU fashion once we've
	// exceeded our maxSize
	c.cl.Lock()
	defer c.cl.Unlock()

	f, err := os.Open(c.dir)
	if err != nil {
		log.Println("cache:", err)
		return
	}
	infos, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Println("cache:", err)
		return
	}

	dirSize := int64(0)
	for _, info := range infos {
		dirSize += info.Size()
	}
	if dirSize <= c.maxSize {
		return
	}

	// sort from oldest to newest files
	// in Go, it's a pain to go by access time, the best
	// we really have is modtime. I might adapt and touch files
	// when they are used to make this more like an LRU.
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].ModTime().Before(infos[j].ModTime())
	})

	// Remove oldest files until we are under our maxSize
	for _, info := range infos {
		if err := os.Remove(filepath.Join(c.dir, info.Name())); err == nil {
			dirSize -= info.Size()
			if dirSize <= c.maxSize {
				return
			}
		}
	}
}

func initCacheManager(downloadUrl string) {
	cacheManager = &CacheManager{
		dir:         cacheDir,
		downloadUrl: downloadUrl,
		maxSize:     1024 * 1024 * 500, // 500MB cache size
	}
}
