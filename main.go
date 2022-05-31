package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Printf("Starting TLS server on :8443\n")

	handler := http.NewServeMux()
	handler.HandleFunc("/create", createFargateProfileHandler)
	handler.HandleFunc("/delete", deleteFargateProfileHandler)
}
