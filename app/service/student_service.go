package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"

	"github.com/gofiber/fiber/v2"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)
	students, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	totalPages := 0
	if q.Limit > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	return helper.OKList(c, "daftar mahasiswa berhasil diambil", students, &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	student, err := s.findAccessibleStudent(ctx, current, id, "student:read:any")
	if err != nil {
		return translateError(c, err, "gagal mengambil data mahasiswa")
	}
	if !CanAccessStudent(current, studentOwnerID(student), s.perms, "student:read:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengakses data mahasiswa lain")
	}
	return helper.OK(c, "mahasiswa ditemukan", student)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)
	if errs := ValidateCreate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	student, err := s.repo.Create(ctx, model.Student{
		OwnerID: &current.UserID, NIM: req.NIM, Name: req.Name, Grade: req.Grade, IsActive: true,
	})
	if err != nil {
		return translateError(c, err, "gagal menyimpan mahasiswa")
	}

	return helper.Created(c, "mahasiswa berhasil dibuat", student, "/api/v1/students/"+strconv.Itoa(student.ID))
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if errs := ValidateReplace(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	existing, err := s.findAccessibleStudent(ctx, current, id, "student:update:any")
	if err != nil {
		return translateError(c, err, "gagal mengambil data mahasiswa")
	}
	if !CanAccessStudent(current, studentOwnerID(existing), s.perms, "student:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data mahasiswa lain")
	}

	student, err := s.repo.Update(ctx, model.Student{
		OwnerID: existing.OwnerID,
		ID:      id, NIM: strings.TrimSpace(req.NIM), Name: strings.TrimSpace(req.Name),
		Grade: req.Grade, IsActive: req.IsActive,
	})
	if err != nil {
		return translateError(c, err, "gagal memperbarui mahasiswa")
	}
	return helper.OK(c, "data mahasiswa berhasil diganti seluruhnya", student)
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if IsEmptyPatch(req) {
		return helper.Fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	currentStudent, err := s.findAccessibleStudent(ctx, current, id, "student:update:any")
	if err != nil {
		return translateError(c, err, "gagal mengambil data mahasiswa")
	}
	if !CanAccessStudent(current, studentOwnerID(currentStudent), s.perms, "student:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data mahasiswa lain")
	}

	updated, errs := ApplyPatch(currentStudent, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	student, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateError(c, err, "gagal memperbarui mahasiswa")
	}
	return helper.OK(c, "data mahasiswa berhasil diperbarui sebagian", student)
}

func studentOwnerID(student model.Student) int {
	if student.OwnerID == nil {
		return 0
	}
	return *student.OwnerID
}

func (s *StudentService) findAccessibleStudent(ctx context.Context, current model.AuthUser, id int, permission string) (model.Student, error) {
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.Student{}, err
	}
	if !CanAccessStudent(current, studentOwnerID(student), s.perms, permission) {
		return model.Student{}, repository.ErrForbidden
	}
	return student, nil
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(c, err, "gagal menghapus mahasiswa")
	}
	return helper.NoContent(c)
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

func translateError(c *fiber.Ctx, err error, fallback string) error {
	if errors.Is(err, repository.ErrForbidden) {
		return helper.Fail(c, fiber.StatusForbidden, err.Error())
	}
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, err.Error())
	}
	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Fail(c, fiber.StatusConflict, err.Error())
	}
	return helper.Fail(c, fiber.StatusInternalServerError, fallback)
}
