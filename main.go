package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Config struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	ServicePort string `json:"service_port"`
	UploadsDir  string `json:"uploads_dir"`
}

var (
	config Config
)

func main() {
	loadConfig("config.json")

	http.HandleFunc("/", serveCDN)
	http.HandleFunc("/upload", handleUpload)

	log.Printf("Starting CDN server on port %s...\n", config.ServicePort)
	log.Fatal(http.ListenAndServe(":"+config.ServicePort, nil))
}

func loadConfig(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error opening config file:", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error decoding config file:", err)
	}
}

func serveCDN(w http.ResponseWriter, r *http.Request) {
	// Log the request information
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)

	// Check if the request has valid authentication
	if !checkAuthentication(r) {
		log.Println("Authentication failed")
		w.Header().Set("WWW-Authenticate", `Basic realm="CDN Authentication"`)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "401 Unauthorized\n")
		return
	}

	// Get the requested file path
	filePath := filepath.Join(config.UploadsDir, r.URL.Path)

	// Check if the file exists
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File not found: %s", r.URL.Path)
			http.NotFound(w, r)
		} else {
			log.Println("Internal Server Error")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Serve the file
	log.Printf("Serving file: %s", r.URL.Path)
	http.ServeFile(w, r, filePath)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	// Log the request information
	log.Printf("Received upload request: %s %s", r.Method, r.URL.Path)

	// Check if the request has valid authentication
	if !checkAuthentication(r) {
		log.Println("Authentication failed")
		w.Header().Set("WWW-Authenticate", `Basic realm="CDN Authentication"`)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "401 Unauthorized\n")
		return
	}

	// Parse the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println("Bad Request")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create the uploads directory if it doesn't exist
	err = os.MkdirAll(config.UploadsDir, os.ModePerm)
	if err != nil {
		log.Println("Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create a new file in the uploads directory
	dst, err := os.Create(filepath.Join(config.UploadsDir, header.Filename))
	if err != nil {
		log.Println("Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	_, err = io.Copy(dst, file)
	if err != nil {
		log.Println("Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("File uploaded successfully: %s", header.Filename)
	fmt.Fprintf(w, "File uploaded successfully!")
}

func checkAuthentication(r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	return ok && username == config.Username && password == config.Password
}
