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

	// Static file route harus di atas Run
	r.Static("/uploads", "./uploads") // akses foto

	// Routing
	routes.SetupRoutes(r)

	// Jalankan server
	r.Run(":8080")
}
