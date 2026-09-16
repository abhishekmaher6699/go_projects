package main


import (
	"flag"
	"fmt"
	"net/http"
	"sync"
)


var store = make(map[string]string)
var mu sync.RWMutex

func main() {

	port := flag.Int("port", 8080, "KV server port")
	flag.Parse()


	http.HandleFunc("/set", setHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/delete", deleteHandler)


	addr := fmt.Sprintf(":%d", *port)
	fmt.Println("KV store running on", addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func setHandler(w http.ResponseWriter, r *http.Request) {

	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	store[key] = value
	mu.Unlock()

	fmt.Fprintln(w, "OK")
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	mu.RLock()
	value, exists := store[key]
	mu.RUnlock()

	if !exists {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	fmt.Fprintln(w, value)
}


func deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	mu.Lock()
	delete(store, key)
	mu.Unlock()

	fmt.Fprintln(w, "OK")
}