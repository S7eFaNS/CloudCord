package main

import (
	"fmt"
	"net/http"
	"time"
)

func handleOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "200 File storage! Current time is: %s", time.Now())
}

func main() {
	http.HandleFunc("/", handleOK) // 200 Ok

	fmt.Println("Starting server on :8082...")
	http.ListenAndServe(":8082", nil)
}
