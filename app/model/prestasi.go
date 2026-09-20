package model

type Prestasi struct {
	IDPrestasi   string `json:"id"`
	NamaPrestasi string `json:"nama_prestasi"`
	Juara        int    `json:"juara"`
	NIM          int    `json:"nim"`
}
