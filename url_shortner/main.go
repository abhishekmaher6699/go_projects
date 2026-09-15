package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"context"
	"time"
	"sync"

	"github.com/redis/go-redis/v9"
	_ "modernc.org/sqlite"
)

var serverID uint64 = 2
const (
	serverBits = 10
	sequenceBits = 12
	serverIDMask = (1 << serverBits) - 1
	sequenceMask = (1 << sequenceBits) - 1
)

var sequence uint64
var lastTimestamp int64

var db *sql.DB
var redisClient *redis.Client

var idMu sync.Mutex

func main() {

	var err error
	db, err = sql.Open("sqlite", "urls.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			code TEXT PRIMARY KEY,
			original_url TEXT NOT NULL
		)
	`)
	if err != nil {
		panic(err)
	}

	redisClient = redis.NewClient((&redis.Options{
		Addr: "localhost:6379",
	}))

	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)

	fmt.Println("Server running on :8081")

	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		panic(err)
	}
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := r.URL.Query().Get("url")

	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	id := generateID()
	code := encodeBase62(id)

	_, err := db.Exec(
		"INSERT INTO urls (code, original_url) VALUES (?, ?)",
		code,
		url,
	)
	if err != nil {
		fmt.Println("DATABASE ERROR:", err)
		http.Error(w, "Could not save URL", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "http://localhost:8080/"+code)

}



func redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return

	}
	code := r.URL.Path[1:]

	url, err := redisClient.Get(r.Context(), code).Result()

	if err == nil {
		fmt.Println("CACHE HIT:", code)

		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	if err != redis.Nil {
		fmt.Println("REDIS ERROR:", err)
	}

	fmt.Println("CACHE MISS:", code)

	err = db.QueryRow(
		"SELECT original_url FROM urls WHERE code = ?",
		code,
	).Scan(&url)

	if err == sql.ErrNoRows {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}


	err = redisClient.Set(r.Context(), code, url, 0).Err()
	if err != nil {
		fmt.Println("REDIS SET ERROR:", err)
	}

	http.Redirect(w, r, url, http.StatusFound)
}


const base62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
func encodeBase62(num uint64) string {
	if num == 0 {
		return "0"
	}

	var result []byte

	for num > 0 {
		remainder := num % 62
		result = append(result, base62[remainder])
		num = num / 62
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}


func generateID() uint64 {
	idMu.Lock()
	defer idMu.Unlock()

	now := time.Now().UnixMilli()

	if now == lastTimestamp {
		sequence = (sequence + 1) & sequenceMask

		if sequence == 0 {
			for now <= lastTimestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		sequence = 0
	}

	lastTimestamp = now

	id := uint64(now)<< (serverBits + sequenceBits)
	id |= (serverID & serverIDMask) << sequenceBits
	id |= sequence

	return id
}