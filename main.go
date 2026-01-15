package main

import (
	"flag"
	"log"
	"net"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/user/go-db-rest-api/database"
	"github.com/user/go-db-rest-api/handlers"
	"github.com/minio/selfupdate"
	"runtime"
	"encoding/json"
	"os"
	"os/exec"
)

var Version = "dev"

func main() {
	// Define flags
	dbType := flag.String("db-type", "", "Database type (mysql, postgres, sqlserver)")
	dbHost := flag.String("db-host", "", "Database host")
	dbPort := flag.Int("db-port", 0, "Database port")
	dbUser := flag.String("db-user", "", "Database user")
	dbPass := flag.String("db-password", "", "Database password")
	dbName := flag.String("db-name", "", "Database name")
	port := flag.Int("port", 8080, "Server port")
	updateURL := flag.String("update-url", "https://pub-44895d9062cf4fbba19c0876bcbe2fbc.r2.dev/bin/version.json", "URL for auto-update (e.g., https://pub-xxx.r2.dev/bin/version.json)")
	checkUpdate := flag.Bool("check-update", true, "Check for updates on startup")
	showVersion := flag.Bool("version", false, "Show version and exit app")
	flag.Parse()

	if *showVersion {
		fmt.Printf("\033[36mgo-db-rest-api\033[0m version \033[32m%s\033[0m\n", Version)
		return
	}

	if *checkUpdate && *updateURL != "" {
		runAutoUpdate(*updateURL)
	}

	// Initialize connection manager
	manager := database.NewConnectionManager()

	// If flags are provided, try to connect automatically with ID "default"
	if *dbType != "" && *dbHost != "" {
		config := database.DBConfig{
			Type:     database.DBType(*dbType),
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Password: *dbPass,
			DBName:   *dbName,
		}
		if err := manager.Connect("default", config); err != nil {
			log.Printf("Warning: failed to connect to default database: %v", err)
		} else {
			log.Println("Connected to default database successfully")
		}
	}

	// Initialize handlers
	config := database.DBConfig{
		Type:     database.DBType(*dbType),
		Host:     *dbHost,
		Port:     *dbPort,
		User:     *dbUser,
		Password: *dbPass,
		DBName:   *dbName,
	}
	h := handlers.NewHandler(manager, config)

	// Initialize Gin router
	r := gin.Default()

	// Routes
	r.POST("/connect", h.Connect)
	r.POST("/query", h.Query)
	r.DELETE("/disconnect/:id", h.Disconnect)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	ips := getLocalIPs()
	
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Printf("Server starting on port %d", *port)
	log.Println("Available at:")
	for _, ip := range ips {
		log.Printf("  - http://%s:%d", ip, *port)
	}
	log.Printf("Health check: http://localhost:%d/health", *port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
}

func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []string{"localhost"}
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	if len(ips) == 0 {
		return []string{"localhost"}
	}
	return ips
}

type UpdateInfo struct {
	Version  string            `json:"version"`
	Binaries map[string]string `json:"binaries"`
}

func runAutoUpdate(checkURL string) {
	log.Printf("Checking for updates at %s...", checkURL)

	resp, err := http.Get(checkURL)
	if err != nil {
		log.Printf("Failed to check for updates: %v", err)
		return
	}
	defer resp.Body.Close()

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Printf("Failed to parse update info: %v", err)
		return
	}

	if info.Version <= Version {
		log.Printf("Current version %s is up to date (latest: %s)", Version, info.Version)
		return
	}

	log.Printf("New version available: %s (current: %s)", info.Version, Version)

	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	binaryPath, ok := info.Binaries[platform]
	if !ok {
		log.Printf("No binary available for platform %s", platform)
		return
	}

	// Resolve binary URL
	u, err := url.Parse(checkURL)
	if err != nil {
		log.Printf("Failed to parse update URL: %v", err)
		return
	}
	
	binaryURL, err := u.Parse(binaryPath)
	if err != nil {
		log.Printf("Failed to resolve binary URL: %v", err)
		return
	}

	log.Printf("Downloading update from %s...", binaryURL.String())
	resp, err = http.Get(binaryURL.String())
	if err != nil {
		log.Printf("Failed to download update: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to download update: status code %d", resp.StatusCode)
		return
	}

	// Basic check to avoid overwriting binary with an HTML error page
	contentType := resp.Header.Get("Content-Type")
	if contentType == "text/html" {
		log.Printf("Failed to download update: received HTML instead of binary (possibly a 404 page)")
		return
	}

	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		log.Printf("Failed to apply update: %v", err)
		return
	}

	log.Println("Update applied successfully! Restarting...")
	restartExecutable()
}
