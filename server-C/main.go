package main

import (
	"fmt"
	"log"
	"net/http"
)

func miniHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Server C \n"))
}

func main() {

	http.HandleFunc("/", miniHandler)

	fmt.Println("running on port 8082")

	log.Fatal(http.ListenAndServe(":8082", nil))
}
