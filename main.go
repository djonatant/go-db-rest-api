package main

import (
	"flag"
	"log"
	"net"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/go-db-rest-api/database"
	"github.com/user/go-db-rest-api/handlers"
)

func main() {
	// Define flags
	dbType := flag.String("db-type", "", "Database type (mysql, postgres, sqlserver)")
	dbHost := flag.String("db-host", "", "Database host")
	dbPort := flag.Int("db-port", 0, "Database port")
	dbUser := flag.String("db-user", "", "Database user")
	dbPass := flag.String("db-password", "", "Database password")
	dbName := flag.String("db-name", "", "Database name")
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

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
	h := handlers.NewHandler(manager)

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
