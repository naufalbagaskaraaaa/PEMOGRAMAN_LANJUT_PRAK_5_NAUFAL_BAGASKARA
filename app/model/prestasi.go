package model

type Prestasi struct {
	id_prestasi        string    `json:"id"`
	nama_prestasi      string    `json:"nama_prestasi"`
	juara     int   `json:"juara"`
	nim  int      `json:"nim"`
}