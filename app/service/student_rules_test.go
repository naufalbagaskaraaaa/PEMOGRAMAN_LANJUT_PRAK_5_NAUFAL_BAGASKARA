package service

import (
	"testing"

	"latihan-fiber/app/model"
)

func TestValidateCreate(t *testing.T) {
	req := model.CreateStudentRequest{NIM: "", Name: "Budi"}
	errs := ValidateCreate(req)

	if len(errs) == 0 {
		t.Error("seharusnya ada error karena NIM kosong")
	}
	if errs["nim"] == "" {
		t.Error("pesan error untuk NIM tidak ditemukan")
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{ID: 1, NIM: "123", Name: "Andi", IsActive: true}
	newName := "Andi Saputra"

	result, errs := ApplyPatch(initial, model.PatchStudentRequest{Name: &newName})

	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if result.Name != newName {
		t.Errorf("nama seharusnya berubah menjadi %s", newName)
	}
	if result.NIM != "123" {
		t.Error("field yang tidak dikirim (NIM) seharusnya tidak berubah")
	}
}

func TestIsEmptyPatch(t *testing.T) {
	req := model.PatchStudentRequest{}
	if !IsEmptyPatch(req) {
		t.Error("seharusnya mengembalikan true karena semua field nil")
	}
}
