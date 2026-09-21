package model

type Prestasi struct {
	IDPrestasi   string  `json:"id_prestasi"`
	NamaPrestasi string  `json:"nama_prestasi"`
	Juara        *int    `json:"juara,omitempty"`
	NIM          *string `json:"nim,omitempty"`
	NamaMhs      *string `json:"nama_mahasiswa,omitempty"`
}

type CreatePrestasiRequest struct {
	IDPrestasi   string  `json:"id_prestasi"`
	NamaPrestasi string  `json:"nama_prestasi"`
	Juara        *int    `json:"juara"`
	NIM          *string `json:"nim"`
}
