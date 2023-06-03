// main.go

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
    "io"
	"path/filepath"
)

const (
	uploadPath = "uploads/"   // Directory to store uploaded files
	username   = "admin"      // Username for authentication
	password   = "password"   // Password for authentication
)

func main() {
	http.HandleFunc("/", serveCDN)
	http.HandleFunc("/upload", handleUpload)

	log.Println("Starting CDN server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
	filePath := filepath.Join("src", r.URL.Path)

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
	err = os.MkdirAll(uploadPath, os.ModePerm)
	if err != nil {
		log.Println("Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create a new file in the uploads directory
	dst, err := os.Create(filepath.Join(uploadPath, header.Filename))
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
	if ok && username == "admin" && password == "password" {
		log.Println("Authentication successful")
		return true
	}
	log.Println("Authentication failed")
	return false
}
