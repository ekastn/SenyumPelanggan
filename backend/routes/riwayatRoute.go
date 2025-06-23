package routes

import (
	"backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Endpoint aktif
	r.POST("/riwayat", controllers.CreateRiwayat)
	// Endpoint berikutnya bisa diaktifkan nanti
	r.GET("/riwayat", controllers.GetRiwayat)
	r.GET("/laporan/export", controllers.ExportExcel)
	r.POST("/deteksi", controllers.JalankanDeteksi)
}
