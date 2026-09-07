package service

import (
	"strings"

	"latihan-fiber/app/model"
)

func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	return validateFields(req.NIM, req.Name, req.Grade)
}

func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	return validateFields(req.NIM, req.Name, req.Grade)
}

func ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)
		if nim == "" {
			errs["nim"] = "tidak boleh kosong"
		} else {
			current.NIM = nim
		}
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = name
		}
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 4 {
			errs["grade"] = "harus antara 0.00 hingga 4.00"
		} else {
			current.Grade = *req.Grade
		}
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

func validateFields(nim, name string, grade float64) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(nim) == "" {
		errs["nim"] = "wajib diisi"
	}
	if strings.TrimSpace(name) == "" {
		errs["name"] = "wajib diisi"
	}
	if grade < 0 || grade > 4 {
		errs["grade"] = "harus antara 0.00 hingga 4.00"
	}
	return errs
}
