package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	ServicePort string `json:"service_port"`
	UploadsDir  string `json:"uploads_dir"`
}

var (
	config        Config
	logger        *logrus.Logger
	filenameRegex *regexp.Regexp
)

func main() {
	// Initialize logger
	logger = logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		DisableTimestamp: false,
		FullTimestamp:    true,
	}

	// Load configuration
	err := loadConfig("config.json")
	if err != nil {
		logger.Fatalf("Error loading config file: %s", err)
	}

	// Initialize filename regular expression
	filenameRegex = regexp.MustCompile(`^[a-zA-Z0-9-_\.]+$`)

	// Initialize router
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.HandleFunc("/", serveCDN)
	router.HandleFunc("/upload", handleUpload)

	// Start server
	address := fmt.Sprintf(":%s", config.ServicePort)
	logger.Infof("Starting CDN server on port %s...", config.ServicePort)
	logger.Infof("Serving files from %s", config.UploadsDir)
	err = http.ListenAndServe(address, router)
	if err != nil {
		logger.Fatalf("Server error: %s", err)
	}
}

func loadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening config file: %s", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return fmt.Errorf("error decoding config file: %s", err)
	}

	return nil
}

func serveCDN(w http.ResponseWriter, r *http.Request) {
	// Log the request information
	logger.Infof("Received request: %s %s", r.Method, r.URL.Path)

	// Check if the request has valid authentication
	if !checkAuthentication(r) {
		logger.Warn("Authentication failed")
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
			logger.Infof("File not found: %s", r.URL.Path)
			http.NotFound(w, r)
		} else {
			logger.Error("Internal Server Error:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Serve the file
	logger.Infof("Serving file: %s", r.URL.Path)
	http.ServeFile(w, r, filePath)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	// Log the request information
	logger.Infof("Received upload request: %s %s", r.Method, r.URL.Path)

	// Check if the request has valid authentication
	if !checkAuthentication(r) {
		logger.Warn("Authentication failed")
		w.Header().Set("WWW-Authenticate", `Basic realm="CDN Authentication"`)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "401 Unauthorized\n")
		return
	}

	// Parse the uploaded file
	err := r.ParseMultipartForm(32 << 20) // Max file size: 32MB
	if err != nil {
		logger.Error("Bad Request:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		logger.Error("No file provided in the request:", err)
		http.Error(w, "No file provided in the request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate filename
	filename := sanitizeFilename(handler.Filename)
	if !isValidFilename(filename) {
		logger.Errorf("Invalid filename: %s", handler.Filename)
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Create the uploads directory if it doesn't exist
	err = os.MkdirAll(config.UploadsDir, os.ModePerm)
	if err != nil {
		logger.Error("Internal Server Error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create a new file in the uploads directory
	dst, err := os.Create(filepath.Join(config.UploadsDir, filename))
	if err != nil {
		logger.Error("Internal Server Error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	_, err = io.Copy(dst, file)
	if err != nil {
		logger.Error("Internal Server Error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Infof("File uploaded successfully: %s", filename)
	fmt.Fprintf(w, "File uploaded successfully!")
}

func checkAuthentication(r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	return ok && username == config.Username && password == config.Password
}

func sanitizeFilename(filename string) string {
	return strings.TrimSpace(filename)
}

func isValidFilename(filename string) bool {
	return filenameRegex.MatchString(filename)
}
