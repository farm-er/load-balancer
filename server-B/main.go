package main

import (
	"fmt"
	"log"
	"net/http"
)

func miniHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Server B \n"))
}

func main() {

	http.HandleFunc("/", miniHandler)

	fmt.Println("running on port 8081")

	log.Fatal(http.ListenAndServe(":8081", nil))
}
