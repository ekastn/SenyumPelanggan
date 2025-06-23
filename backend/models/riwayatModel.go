package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Struct untuk nested field presentase
type Presentase struct {
	Netral       float64 `bson:"netral" json:"netral"`
	Bahagia      float64 `bson:"bahagia" json:"bahagia"`
	TidakBahagia float64 `bson:"tidak_bahagia" json:"tidak_bahagia"`
}

// Struct utama RiwayatEmosi
type RiwayatEmosi struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Waktu              string             `bson:"waktu" json:"waktu"`
	DurasiDeteksi      float64            `bson:"durasi_deteksi" json:"durasi_deteksi"`
	DurasiNetral       float64            `bson:"durasi_netral" json:"durasi_netral"`
	DurasiBahagia      float64            `bson:"durasi_bahagia" json:"durasi_bahagia"`
	DurasiTidakBahagia float64            `bson:"durasi_tidak_bahagia" json:"durasi_tidak_bahagia"`
	Presentase         Presentase         `bson:"presentase" json:"presentase"`
	EmosiDominan       string             `bson:"emosi_dominan" json:"emosi_dominan"`
	PathFoto           string             `bson:"path_foto" json:"path_foto"`
}
