package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var ring *HashRing

func main() {

	ring = NewHashRing()

	ring.AddNode("http://localhost:8080")
	ring.AddNode("http://localhost:8081")
	ring.AddNode("http://localhost:8082")

	http.HandleFunc("/", HandleRequest)

	fmt.Println("Router running on :9000")

	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		panic(err)
	}
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	nodes := ring.GetNodes(key, 2)

	fmt.Println("key:", key, "→", nodes)

	if r.URL.Path == "/set" {
		err := replicateSet(r, nodes)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		fmt.Fprintln(w, "OK")
		return
	}

	if r.URL.Path == "/get" {
		resp, err := getFromNode(r, nodes[0])

		if err == nil {
			defer resp.Body.Close()

			for k, values := range resp.Header {
				for _, value := range values {
					w.Header().Add(k, value)
				}
			}

			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
			return
		}

		fmt.Println("primary failed:", nodes[0])
		fmt.Println("trying replica:", nodes[1])

		resp, err = getFromNode(r, nodes[1])

		if err != nil {
			http.Error(w, "all replicas unavailable", http.StatusBadGateway)
			return
		}

		defer resp.Body.Close()

		for k, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(k, value)
			}
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}

	proxyToNode(w, r, nodes[0])
}

func getFromNode(r *http.Request, node string) (*http.Response, error) {
	target := node + r.URL.RequestURI()

	return http.Get(target)
}

func proxyToNode(w http.ResponseWriter, r *http.Request, node string) {
	target, err := url.Parse(node)

	if err != nil {
		http.Error(w, "invalid node", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "node unavailable", http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}

func replicateSet(r *http.Request, nodes []string) error {
	for _, node := range nodes {
		target := node + r.URL.RequestURI()

		resp, err := http.Get(target)
		if err != nil {
			return err
		}

		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("node %s returned %s", node, resp.Status)
		}
	}

	return nil
}
