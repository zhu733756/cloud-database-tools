package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
)

type UserData struct {
	mu    sync.RWMutex
	Cache map[string][]byte
}

func NewDataMgr() *UserData {
	return &UserData{
		Cache: make(map[string][]byte),
	}
}

var dataMgr = NewDataMgr()

func write(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")

	dataMgr.mu.Lock()
	defer dataMgr.mu.Unlock()
	// dataMgr.Cache[userID] = make([]byte, 1000000)
	dataMgr.Cache[userID] = make([]byte, 0, 1000000)
	log.Printf("Added data for user %s. Total users: %d\n", userID, len(dataMgr.Cache))
}

func read(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")

	dataMgr.mu.RLock()
	defer dataMgr.mu.RUnlock()

	if userData, exists := dataMgr.Cache[userID]; exists {
		log.Printf("Found data for user %s, len: %d, cap: %d\n", userID, len(userData), cap(userData))
	} else {
		log.Printf("not found %s. Total users: %d\n", userID, len(dataMgr.Cache))
	}
}

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
	http.HandleFunc("/write", write)
	http.HandleFunc("/read", read)
	http.ListenAndServe(":8080", nil)
}
