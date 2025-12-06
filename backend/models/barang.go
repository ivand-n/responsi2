package models

type Barang struct {
	id          int
	nama        string
	harga       int
	jumlah      int
	tgl_masuk   string
	tgl_expired string
}