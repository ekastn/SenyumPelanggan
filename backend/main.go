package main

import (
	"backend/config"
	"backend/middleware"
	"backend/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Koneksi ke MongoDB
	config.ConnectDB()

	// Setup server
	r := gin.Default()
	r.Use(middleware.SetupCORS())

	// Static file routes
	r.Static("/uploads", "./uploads")

	// Routing
	routes.SetupRoutes(r)

	// Jalankan server
	r.Run(":8080")
}
