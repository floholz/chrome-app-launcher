package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/open/{encodedUrl}", openHandler).Methods("GET")
	r.HandleFunc("/create", createHandler).Methods("POST")

	addr := ":5252"
	fmt.Println("Server running on http://localhost" + addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func openHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encodedURL := vars["encodedUrl"]

	decodedURL, err := url.PathUnescape(encodedURL)
	if err != nil {
		http.Error(w, "Invalid URL encoding", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(decodedURL, "http") && !strings.HasPrefix(decodedURL, "localhost") {
		http.Error(w, "Invalid URL. Must be 'http', 'https' or 'localhost'", http.StatusBadRequest)
	}

	decodedURL = strings.ReplaceAll(decodedURL, "~", "://")
	decodedURL = strings.ReplaceAll(decodedURL, "+", "/")

	log.Println("Received URL:", decodedURL)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Opening: %s\n", decodedURL)
	errExec := launchChromeApp(decodedURL)
	if errExec != nil {
		log.Println(errExec)
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	// Implementation goes here
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintln(w, "Create handler not implemented yet")
}

func launchChromeApp(url string) error {
	execCmd := exec.Command("C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe", "--app="+url)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}
