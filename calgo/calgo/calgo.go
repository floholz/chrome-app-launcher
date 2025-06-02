package calgo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/badgerodon/penv"
	"github.com/gorilla/mux"
)

type ChromeConfigRequest struct {
	Path string `json:"path"`
}

var chromePath string
var serverPort = 5252

func init() {
	// check for dev
	if port := os.Getenv("CALGO_PORT"); port != "" {
		portInt, err := strconv.ParseInt(port, 10, 32)
		if err == nil {
			serverPort = int(portInt)
		}
	}

	// find chrome path
	chromePath = os.Getenv("CALGO_CHROME_PATH")
	if chromePath != "" {
		return
	}

	path, err := findChromeExecutable()
	if err != nil {
		log.Fatal(err)
	}

	err = penv.SetEnv("CALGO_CHROME_PATH", path)
	if err != nil {
		log.Fatal(err)
	}
	chromePath = path
}

func Start(ctx context.Context) error {
	r := mux.NewRouter()
	r.HandleFunc("/open/{encodedUrl}", openHandler).Methods("GET")
	r.HandleFunc("/status", statusHandler).Methods("GET")
	r.HandleFunc("/config/{action}", configHandler).Methods("POST")

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(serverPort),
		Handler: r,
	}

	// Run server in a goroutine
	go func() {
		fmt.Printf("Server running on http://localhost:%d\n", serverPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Attempt graceful shutdown
	fmt.Println("Shutting down server...")
	return server.Shutdown(context.Background())
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
	errExec := launchChromeApp(decodedURL)
	if errExec != nil {
		log.Println(errExec)
		http.Error(w, errExec.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Starting: %s\n", decodedURL)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("CalGo server is running"))
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action := vars["action"]

	switch action {
	case "reset":
		err := penv.UnsetEnv("CALGO_CHROME_PATH")
		if err != nil {
			http.Error(w, "Failed to reset server!", http.StatusInternalServerError)
		}
		w.Write([]byte("Reset server successfully!"))
		return
	case "chrome":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ChromeConfigRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		err = penv.SetEnv("CALGO_CHROME_PATH", req.Path)
		if err != nil {
			http.Error(w, "Failed to set chrome executable path!", http.StatusInternalServerError)
			return
		}
		chromePath = req.Path
		w.Write([]byte("Set chrome executable path successfully!"))
		return

	default:
		http.Error(w, "Unknown config action: "+action, http.StatusBadRequest)
		return
	}
}

func launchChromeApp(url string) error {
	if chromePath == "" {
		return fmt.Errorf("chrome executable path not set")
	}

	execCmd := exec.Command(chromePath, "--app="+url)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}

func findChromeExecutable() (string, error) {
	// Step 1: Check CHROME_PATH env variable
	if path := os.Getenv("CHROME_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Step 2: Check common paths
	possiblePaths := []string{}

	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		programFiles := os.Getenv("ProgramFiles")

		possiblePaths = append(possiblePaths,
			filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"),
		)
	} else {
		possiblePaths = append(possiblePaths,
			"/usr/bin/google-chrome",
			"/opt/google/chrome/google-chrome",
			"/usr/bin/chromium-browser",
		)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Step 3: Try looking in PATH
	binName := "google-chrome"
	if runtime.GOOS == "windows" {
		binName = "chrome.exe"
	}

	if chromePath, err := exec.LookPath(binName); err == nil {
		return chromePath, nil
	}

	return "", fmt.Errorf("google-chrome executable not found")
}
