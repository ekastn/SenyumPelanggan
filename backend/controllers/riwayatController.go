package controllers

import (
	"backend/config"
	"backend/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// POST /riwayat
func CreateRiwayat(c *gin.Context) {
	file, err := c.FormFile("foto")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Foto tidak ditemukan"})
		return
	}

	os.MkdirAll("uploads", os.ModePerm)
	filename := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal simpan foto"})
		return
	}

	durasiDeteksi := c.PostForm("durasi_deteksi")
	durasiNetral := c.PostForm("durasi_netral")
	durasiBahagia := c.PostForm("durasi_bahagia")
	durasiTidakBahagia := c.PostForm("durasi_tidak_bahagia")
	presentaseNetral := c.PostForm("presentase_netral")
	presentaseBahagia := c.PostForm("presentase_bahagia")
	presentaseTidakBahagia := c.PostForm("presentase_tidak_bahagia")
	emosiDominan := c.PostForm("emosi_dominan")

	var data models.RiwayatEmosi
	data.ID = primitive.NewObjectID()
	data.Waktu = time.Now().Format(time.RFC3339)
	fmt.Sscanf(durasiDeteksi, "%f", &data.DurasiDeteksi)
	fmt.Sscanf(durasiNetral, "%f", &data.DurasiNetral)
	fmt.Sscanf(durasiBahagia, "%f", &data.DurasiBahagia)
	fmt.Sscanf(durasiTidakBahagia, "%f", &data.DurasiTidakBahagia)
	fmt.Sscanf(presentaseNetral, "%f", &data.Presentase.Netral)
	fmt.Sscanf(presentaseBahagia, "%f", &data.Presentase.Bahagia)
	fmt.Sscanf(presentaseTidakBahagia, "%f", &data.Presentase.TidakBahagia)
	data.EmosiDominan = emosiDominan
	data.PathFoto = "/" + filename

	collection := config.GetCollection("riwayat_emosi")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal simpan data ke database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil disimpan"})
}

// GET /riwayat
type FilterParams struct {
	Tanggal string
	Minggu  string
	Bulan   string
	Tahun   string
	Dari    string
	Sampai  string
}

func buildTimeFilter(params FilterParams) bson.M {
	filter := bson.M{}

	if params.Tanggal != "" {
		t, err := time.Parse("2006-01-02", params.Tanggal)
		if err == nil {
			filter["waktu"] = bson.M{"$gte": t.Format(time.RFC3339), "$lt": t.Add(24 * time.Hour).Format(time.RFC3339)}
		}
	} else if params.Minggu != "" {
		t, err := time.Parse("2006-01-02", params.Minggu)
		if err == nil {
			filter["waktu"] = bson.M{"$gte": t.AddDate(0, 0, -6).Format(time.RFC3339), "$lt": t.Add(24 * time.Hour).Format(time.RFC3339)}
		}
	} else if params.Bulan != "" {
		t, err := time.Parse("2006-01", params.Bulan)
		if err == nil {
			filter["waktu"] = bson.M{"$gte": t.Format(time.RFC3339), "$lt": t.AddDate(0, 1, 0).Format(time.RFC3339)}
		}
	} else if params.Tahun != "" {
		t, err := time.Parse("2006", params.Tahun)
		if err == nil {
			filter["waktu"] = bson.M{"$gte": t.Format(time.RFC3339), "$lt": t.AddDate(1, 0, 0).Format(time.RFC3339)}
		}
	} else if params.Dari != "" && params.Sampai != "" {
		t1, err1 := time.Parse("2006-01", params.Dari)
		t2, err2 := time.Parse("2006-01", params.Sampai)
		if err1 == nil && err2 == nil {
			filter["waktu"] = bson.M{"$gte": t1.Format(time.RFC3339), "$lt": t2.AddDate(0, 1, 0).Format(time.RFC3339)}
		}
	}
	return filter
}

func GetRiwayat(c *gin.Context) {
	collection := config.GetCollection("riwayat_emosi")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := FilterParams{
		Tanggal: c.Query("tanggal"),
		Minggu:  c.Query("minggu"),
		Bulan:   c.Query("bulan"),
		Tahun:   c.Query("tahun"),
		Dari:    c.Query("dari"),
		Sampai:  c.Query("sampai"),
	}
	filter := buildTimeFilter(params)

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil data dari database"})
		return
	}
	defer cursor.Close(ctx)

	var riwayat []models.RiwayatEmosi
	if err := cursor.All(ctx, &riwayat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal decode data riwayat"})
		return
	}

	c.JSON(http.StatusOK, riwayat)
}

func ExportExcel(c *gin.Context) {
	collection := config.GetCollection("riwayat_emosi")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := FilterParams{
		Tanggal: c.Query("tanggal"),
		Minggu:  c.Query("minggu"),
		Bulan:   c.Query("bulan"),
		Tahun:   c.Query("tahun"),
		Dari:    c.Query("dari"),
		Sampai:  c.Query("sampai"),
	}
	filter := buildTimeFilter(params)

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil data dari database"})
		return
	}
	defer cursor.Close(ctx)

	var riwayat []models.RiwayatEmosi
	if err := cursor.All(ctx, &riwayat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal decode data"})
		return
	}

	var totalDeteksi, netralDur, bahagiaDur, tidakBahagiaDur float64
	for _, data := range riwayat {
		totalDeteksi++
		netralDur += data.DurasiNetral
		bahagiaDur += data.DurasiBahagia
		tidakBahagiaDur += data.DurasiTidakBahagia
	}

	totalDurasi := netralDur + bahagiaDur + tidakBahagiaDur
	if totalDurasi == 0 {
		totalDurasi = 1
	}

	dominan := "Bahagia"
	if netralDur > bahagiaDur && netralDur > tidakBahagiaDur {
		dominan = "Netral"
	} else if tidakBahagiaDur > bahagiaDur {
		dominan = "Tidak Bahagia"
	}

	f := excelize.NewFile()
	sheet := "Laporan"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	headers := []string{
		"Waktu Laporan", "Total Deteksi", "Durasi Netral", "Durasi Bahagia", "Durasi Tidak Bahagia",
		"Presentase Netral", "Presentase Bahagia", "Presentase Tidak Bahagia", "Emosi Dominan",
	}
	for i, h := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheet, cell, h)
	}

	f.SetCellValue(sheet, "A2", time.Now().Format("2006-01-02 15:04:05"))
	f.SetCellValue(sheet, "B2", totalDeteksi)
	f.SetCellValue(sheet, "C2", fmt.Sprintf("%.2f", netralDur))
	f.SetCellValue(sheet, "D2", fmt.Sprintf("%.2f", bahagiaDur))
	f.SetCellValue(sheet, "E2", fmt.Sprintf("%.2f", tidakBahagiaDur))
	f.SetCellValue(sheet, "F2", fmt.Sprintf("%.2f", (netralDur/totalDurasi)*100))
	f.SetCellValue(sheet, "G2", fmt.Sprintf("%.2f", (bahagiaDur/totalDurasi)*100))
	f.SetCellValue(sheet, "H2", fmt.Sprintf("%.2f", (tidakBahagiaDur/totalDurasi)*100))
	f.SetCellValue(sheet, "I2", dominan)

	filename := fmt.Sprintf("laporan_ringkasan_%d.xlsx", time.Now().Unix())
	filepath := "uploads/" + filename
	if err := f.SaveAs(filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal simpan file Excel"})
		return
	}

	c.FileAttachment(filepath, filename)
}

// POST /deteksi
func JalankanDeteksi(c *gin.Context) {
	var payload struct {
		Frames   []string `json:"frames"`
		Interval float64  `json:"interval"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	// Kirim JSON payload ke Python via stdin
	cmd := exec.Command("python", "deteksi_batch.py")
	cmd.Dir = "../emotion-core"

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menjalankan Python"})
		return
	}

	go func() {
		defer stdin.Close()
		input, _ := json.Marshal(payload)
		stdin.Write(input)
	}()

	output, _ := io.ReadAll(stdout)
	cmd.Wait()

	// Parse output dari Python
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca hasil dari Python"})
		return
	}

	// Kirim hasil ke endpoint POST /riwayat
	pathFoto := result["path_foto"].(string)
	file, err := os.Open("../emotion-core/" + pathFoto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka foto"})
		return
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("durasi_deteksi", fmt.Sprint(result["durasi_deteksi"]))
	writer.WriteField("durasi_netral", fmt.Sprint(result["durasi_netral"]))
	writer.WriteField("durasi_bahagia", fmt.Sprint(result["durasi_bahagia"]))
	writer.WriteField("durasi_tidak_bahagia", fmt.Sprint(result["durasi_tidak_bahagia"]))
	writer.WriteField("presentase_netral", fmt.Sprint(result["presentase_netral"]))
	writer.WriteField("presentase_bahagia", fmt.Sprint(result["presentase_bahagia"]))
	writer.WriteField("presentase_tidak_bahagia", fmt.Sprint(result["presentase_tidak_bahagia"]))
	writer.WriteField("emosi_dominan", fmt.Sprint(result["emosi_dominan"]))

	part, _ := writer.CreateFormFile("foto", filepath.Base(pathFoto))
	io.Copy(part, file)
	writer.Close()

	resp, err := http.Post("http://localhost:8080/riwayat", writer.FormDataContentType(), body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal kirim ke riwayat"})
		return
	}
	defer resp.Body.Close()
	resBody, _ := io.ReadAll(resp.Body)

	c.JSON(http.StatusOK, gin.H{
		"message": "Deteksi selesai dan data disimpan",
		"result":  string(resBody),
	})
}
