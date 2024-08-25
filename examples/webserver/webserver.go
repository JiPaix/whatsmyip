// Fetches the IP once every 24 Hours
// The machine current ip is returned at http://localhost:8080/ip
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jipaix/whatsmyip"
)

type IPCache struct {
	mu        sync.RWMutex
	ip        string
	lastFetch time.Time
}

func (c *IPCache) updateIP() {
	for {
		ip, _, err := whatsmyip.Get()
		if err != nil {
			fmt.Printf("Error fetching IP: %v\n", err)
		} else {
			c.mu.Lock()
			c.ip = ip
			c.lastFetch = time.Now()
			c.mu.Unlock()
			fmt.Printf("IP updated: %s at %s\n", ip, c.lastFetch.Format(time.RFC3339))
		}
		time.Sleep(24 * time.Hour)
	}
}

func (c *IPCache) getIP() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ip
}

func main() {
	cache := &IPCache{}

	// Start the IP updating goroutine
	go cache.updateIP()

	// Create a web server
	http.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
		ip := cache.getIP()
		if ip == "" {
			http.Error(w, "IP not available yet", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintf(w, "Current IP: %s\n", ip)
	})

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
