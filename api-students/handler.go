package main

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var students []Student
var nextID = 1

func findStudentIndex(id int) int {
	for i, s := range students {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func isNIMExists(nim string, excludeID int) bool {
	for _, s := range students {
		if strings.EqualFold(s.NIM, nim) && s.ID != excludeID {
			return true
		}
	}
	return false
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)
	var filtered []Student

	for _, s := range students {
		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue
		}
		if q.Search != "" && !strings.Contains(strings.ToLower(s.Name), strings.ToLower(q.Search)) {
			continue
		}
		filtered = append(filtered, s)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		var less bool
		switch q.Sort {
		case "nim":
			less = filtered[i].NIM < filtered[j].NIM
		case "name":
			less = strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		case "grade":
			less = filtered[i].Grade < filtered[j].Grade
		default:
			less = filtered[i].ID < filtered[j].ID
		}
		if q.Order == "desc" {
			return !less
		}
		return less
	})

	total := len(filtered)
	totalPages := (total + q.Limit - 1) / q.Limit
	start := (q.Page - 1) * q.Limit
	if start > total {
		start = total
	}
	end := start + q.Limit
	if end > total {
		end = total
	}

	return okList(c, "daftar mahasiswa berhasil diambil", filtered[start:end], &Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	idx := findStudentIndex(id)
	if idx == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	return ok(c, "mahasiswa ditemukan", students[idx])
}

func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}
	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 4.0 {
		errs["grade"] = "harus bernilai antara 0.0 - 4.0"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Cek Status 409 Conflict jika NIM ganda
	if isNIMExists(req.NIM, 0) {
		return fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
	}

	newStudent := Student{
		ID:       nextID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
	}
	nextID++
	students = append(students, newStudent)

	return created(c, "mahasiswa berhasil ditambahkan", newStudent, "/api/v1/students/"+strconv.Itoa(newStudent.ID))
}

func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	idx := findStudentIndex(id)
	if idx == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if req.Name == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 || req.Grade > 4.0 {
		errs["grade"] = "harus bernilai antara 0.0 - 4.0"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	if isNIMExists(req.NIM, id) {
		return fail(c, fiber.StatusConflict, "NIM sudah terdaftar pada mahasiswa lain")
	}

	students[idx].NIM = req.NIM
	students[idx].Name = req.Name
	students[idx].Grade = req.Grade
	students[idx].IsActive = req.IsActive

	return ok(c, "data mahasiswa berhasil diganti seluruhnya", students[idx])
}

func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	idx := findStudentIndex(id)
	if idx == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)
		if nim == "" {
			return failValidation(c, map[string]string{"nim": "tidak boleh kosong"})
		}
		if isNIMExists(nim, id) {
			return fail(c, fiber.StatusConflict, "NIM sudah terdaftar pada mahasiswa lain")
		}
		students[idx].NIM = nim
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return failValidation(c, map[string]string{"name": "tidak boleh kosong"})
		}
		students[idx].Name = name
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 4.0 {
			return failValidation(c, map[string]string{"grade": "harus bernilai antara 0.0 - 4.0"})
		}
		students[idx].Grade = *req.Grade
	}
	if req.IsActive != nil {
		students[idx].IsActive = *req.IsActive
	}

	return ok(c, "data mahasiswa berhasil diperbarui sebagian", students[idx])
}

func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	idx := findStudentIndex(id)
	if idx == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	students = append(students[:idx], students[idx+1:]...)
	return noContent(c)
}
