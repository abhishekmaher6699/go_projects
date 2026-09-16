package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"sync"
	"time"
)


type Backend struct {
	URL *url.URL
	Alive bool
	Mutex sync.RWMutex
}

var backends []*Backend



var counter uint64

func main() {


	var servers = []string{
		"http://localhost:8080",
		"http://localhost:8081",
	}

	for _, serverURL := range servers {
		target, err := url.Parse(serverURL)
		if err != nil {
			panic(err)
		}

		backend := &Backend{
			URL: target,
			Alive: true,
		}

		backends = append(backends, backend)
	}	

	go healthChecker()


	http.HandleFunc("/", proxyHandler)

	fmt.Println("Load balancer running on :9000")

	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		panic(err)
	}
}


func proxyHandler(w http.ResponseWriter, r *http.Request) {

	backend := getNextBackend()

	if backend == nil {
		http.Error(
			w,
			"No healthy servers available",
			http.StatusServiceUnavailable,
		)
		return
	}

	fmt.Println("Forawrding request to:", backend.URL)

	proxy := httputil.NewSingleHostReverseProxy(backend.URL)
	proxy.ServeHTTP(w, r)
}

func getNextBackend() *Backend {

	start := atomic.AddUint64(&counter, 1)

	for i := uint64(0); i < uint64(len(backends)); i++ {

		index := (start + i) % uint64(len(backends))

		backend := backends[index]

		backend.Mutex.RLock()
		alive := backend.Alive
		backend.Mutex.RUnlock()

		if alive {
			return backend
		}
	}

	return nil
}

func healthChecker() {
	for {

		for _, backend := range backends {
			healthy := checkHealth(backend)

			backend.Mutex.Lock()

			if backend.Alive != healthy {
				if healthy {
					fmt.Println("Server recovered:", backend.URL)
				} else {
					fmt.Println("Server is down:", backend.URL)
				}
			}

			backend.Alive = healthy

			backend.Mutex.Unlock()
		}

		time.Sleep(2 * time.Second)
	}
}

func checkHealth(backend *Backend) bool {

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	resp, err := client.Get(
		backend.URL.String() + "/health",
	)

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}