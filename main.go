package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Config represents the configuration structure
type Config struct {
	Port            string   `json:"port"`
	MaxUploadSize   int64    `json:"max_upload_size"`
	Heartbeat       Duration `json:"heartbeat"`
	RequireAuth     bool     `json:"require_auth"`
	AuthUsername    string   `json:"auth_username"`
	AuthPassword    string   `json:"auth_password"`
	UploadDirectory string   `json:"upload_directory"`
}

// Duration is a custom type for decoding time.Duration from JSON
type Duration time.Duration

// UnmarshalJSON unmarshals a JSON string into a Duration
func (d *Duration) UnmarshalJSON(data []byte) error {
	var durationStr string
	if err := json.Unmarshal(data, &durationStr); err != nil {
		return fmt.Errorf("error decoding Duration: %w", err)
	}

	parsedDuration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("error parsing Duration: %w", err)
	}

	*d = Duration(parsedDuration)
	return nil
}

// CDNHandler handles HTTP requests to the CDN server
type CDNHandler struct {
	Config Config
	Logger *logrus.Logger
}

// ServeHTTP serves HTTP requests
func (c *CDNHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.handleGet(w, r)
	case http.MethodPost:
		c.handlePost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET requests
func (c *CDNHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	c.Logger.Infof("Received GET request for URL: %s", r.URL.Path)

	// Serve file for download
	filePath := filepath.Join(c.Config.UploadDirectory, r.URL.Path)
	file, err := os.Open(filePath)
	if err != nil {
		c.Logger.Errorf("Error opening file: %v", err)
		http.Error(w, "File Not Found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set the appropriate Content-Type header based on file extension
	contentType := "application/octet-stream"
	switch filepath.Ext(filePath) {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".pdf":
		contentType = "application/pdf"
	}

	w.Header().Set("Content-Type", contentType)
	if _, err := io.Copy(w, file); err != nil {
		c.Logger.Errorf("Error copying file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	c.Logger.Infof("File downloaded successfully: %s", r.URL.Path)
}

// handlePost handles POST requests
func (c *CDNHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	c.Logger.Info("Received POST request")

	// Validate request size
	r.Body = http.MaxBytesReader(w, r.Body, c.Config.MaxUploadSize)
	if err := r.ParseMultipartForm(c.Config.MaxUploadSize); err != nil {
		c.Logger.Errorf("Error parsing multipart form: %v", err)
		http.Error(w, "Payload Too Large", http.StatusRequestEntityTooLarge)
		return
	}

	// Get the uploaded file
	file, handler, err := r.FormFile("file")
	if err != nil {
		c.Logger.Errorf("Error retrieving file: %v", err)
		http.Error(w, "Error retrieving file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Create the upload directory if it doesn't exist
	uploadDir := c.Config.UploadDirectory
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.Logger.Errorf("Error creating upload directory: %v", err)
		http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
		return
	}

	// Create the file in the upload directory
	filePath := filepath.Join(uploadDir, handler.Filename)
	newFile, err := os.Create(filePath)
	if err != nil {
		c.Logger.Errorf("Error creating file: %v", err)
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer newFile.Close()

	// Copy the uploaded file to the new file
	if _, err := io.Copy(newFile, file); err != nil {
		c.Logger.Errorf("Error copying file: %v", err)
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	c.Logger.Infof("File uploaded successfully: %s", handler.Filename)
	fmt.Fprint(w, "File uploaded successfully")
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	flag.Parse()

	// Initialize logrus logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Read the configuration file
	configFile, err := os.Open(*configPath)
	if err != nil {
		logger.Fatalf("Error opening configuration file: %v", err)
	}
	defer configFile.Close()

	// Decode the configuration file
	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		logger.Fatalf("Error decoding configuration file: %v", err)
	}

	// Start a goroutine to restart the server periodically
	if config.Heartbeat > 0 {
		go func() {
			for range time.Tick(time.Duration(config.Heartbeat)) {
				logger.Info("Server heartbeat")
				server := startServer(&config, logger)
				stopServer(server, logger)
			}
		}()
	}

	// Start the initial server
	server := startServer(&config, logger)

	// Wait for termination signal
	waitForTerminationSignal()

	// Stop the server before exiting
	stopServer(server, logger)
}

// startServer creates and starts the HTTP server
func startServer(config *Config, logger *logrus.Logger) *http.Server {
	// Create a new CDNHandler with the configuration
	cdnHandler := &CDNHandler{
		Config: *config,
		Logger: logger,
	}

	// Create a new HTTP server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      cdnHandler,
		ErrorLog:     log.New(logger.Writer(), "", 0),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start the server in a separate goroutine
	go func() {
		logger.Infof("Starting CDN server on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	return server
}

// stopServer stops the HTTP server
func stopServer(server *http.Server, logger *logrus.Logger) {
	logger.Info("Stopping CDN server")

	// Set a deadline for gracefully shutting down the server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down the server with the given context
	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}
}

// waitForTerminationSignal waits for termination signals to gracefully shut down the server
func waitForTerminationSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
}
