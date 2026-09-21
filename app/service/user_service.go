package service

import (
	"errors"
	"strconv"
	"strings"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"

	"github.com/gofiber/fiber/v2"
)

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(repo repository.UserRepository, perms *helper.PermissionSet) *UserService {
	return &UserService{repo: repo, perms: perms}
}

func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar user")
	}
	return helper.OK(c, "daftar user berhasil diambil", users)
}

func (s *UserService) Create(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Email) == "" {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "username dan email wajib diisi")
	}
	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.repo.Create(ctx, model.User{Username: strings.TrimSpace(req.Username), Email: strings.TrimSpace(req.Email), PasswordHash: hash, Role: "user"})
	if err != nil {
		if errors.Is(err, repository.ErrUsernameOrEmailTaken) {
			return helper.Fail(c, fiber.StatusConflict, err.Error())
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat user")
	}
	return helper.Created(c, "user berhasil dibuat", user, "/api/v1/users/"+strconv.Itoa(user.ID))
}

func (s *UserService) Get(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, err.Error())
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data user")
	}
	if !CanAccessUser(current, user.ID, s.perms, "user:read:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengakses data user lain")
	}
	return helper.OK(c, "data user ditemukan", user)
}

func (s *UserService) Replace(c *fiber.Ctx) error {
	return s.update(c, true)
}

func (s *UserService) Patch(c *fiber.Ctx) error {
	return s.update(c, false)
}

func (s *UserService) update(c *fiber.Ctx, replace bool) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if replace && (strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Email) == "") {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "username dan email wajib diisi")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return s.userError(c, err)
	}
	if !CanAccessUser(current, user.ID, s.perms, "user:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data user lain")
	}
	if strings.TrimSpace(req.Username) == "" {
		req.Username = user.Username
	}
	if strings.TrimSpace(req.Email) == "" {
		req.Email = user.Email
	}
	updated, err := s.repo.Update(ctx, model.User{ID: id, Username: strings.TrimSpace(req.Username), Email: strings.TrimSpace(req.Email)})
	if err != nil {
		return s.userError(c, err)
	}
	return helper.OK(c, "data user berhasil diperbarui", updated)
}

func (s *UserService) AssignRole(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if errs := ValidateAssignRole(current, id, req, s.perms); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		return s.userError(c, err)
	}
	return helper.OK(c, "role user berhasil diperbarui", user)
}

func (s *UserService) Delete(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	if current.UserID == id {
		return helper.Fail(c, fiber.StatusForbidden, "tidak boleh menghapus akun sendiri")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, err.Error())
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus user")
	}
	return helper.NoContent(c)
}

func (s *UserService) userError(c *fiber.Ctx, err error) error {
	if errors.Is(err, repository.ErrUserNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, err.Error())
	}
	if errors.Is(err, repository.ErrUsernameOrEmailTaken) {
		return helper.Fail(c, fiber.StatusConflict, err.Error())
	}
	return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses user")
}
