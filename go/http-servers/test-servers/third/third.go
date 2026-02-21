package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from backend 3\n"))
	})
	log.Fatal(http.ListenAndServe(":9003", nil))
}
