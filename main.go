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
	publicIP := getLocalIP()
	
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Printf("Server starting on http://%s%s", publicIP, addr)
	log.Printf("Health check: http://%s%s/health", publicIP, addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}
