package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func handleOK(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "200 notifications! Current time is: %s", time.Now())

	log.Printf("Request received: Method: %s, Path: %s, Headers: %v\n", r.Method, r.URL.Path, r.Header)

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/", handleOK) // 200 Ok

	fmt.Println("Starting server on :8083...")

	http.ListenAndServe(":8083", nil)

	log.Fatal(http.ListenAndServe(":8083", nil))
}
