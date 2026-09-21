package service

import (
	"errors"
	"strings"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"

	"github.com/gofiber/fiber/v2"
)

type PrestasiService struct {
	repo repository.PrestasiRepository
}

func NewPrestasiService(
	repo repository.PrestasiRepository,
) *PrestasiService {
	return &PrestasiService{repo: repo}
}

func (s *PrestasiService) Create(c *fiber.Ctx) error {
	var req model.CreatePrestasiRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	req.IDPrestasi = strings.TrimSpace(req.IDPrestasi)
	req.NamaPrestasi = strings.TrimSpace(req.NamaPrestasi)
	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)
		req.NIM = &nim
	}

	if req.IDPrestasi == "" || req.NamaPrestasi == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "id_prestasi dan nama_prestasi wajib diisi")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	prestasi, err := s.repo.Create(ctx, req)
	if err != nil {
		return helper.Fail(c, fiber.StatusConflict, "id prestasi atau data relasi sudah digunakan")
	}

	return helper.Created(c, "prestasi berhasil ditambahkan", prestasi, "/api/v1/prestasi/"+prestasi.IDPrestasi)
}

func (s *PrestasiService) Get(c *fiber.Ctx) error {
	idPrestasi := strings.TrimSpace(c.Params("id_prestasi"))

	if idPrestasi == "" {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id prestasi wajib diisi",
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	prestasi, err := s.repo.FindByID(ctx, idPrestasi)
	if err != nil {
		if errors.Is(err, repository.ErrPrestasiNotFound) {
			return helper.Fail(
				c,
				fiber.StatusNotFound,
				"prestasi tidak ditemukan",
			)
		}

		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data prestasi",
		)
	}

	return helper.OK(
		c,
		"data prestasi berhasil diambil",
		prestasi,
	)
}

func (s *PrestasiService) ListByNIM(c *fiber.Ctx) error {
	nim := strings.TrimSpace(c.Params("nim"))
	if nim == "" {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"nim wajib diisi",
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	prestasiList, err := s.repo.FindByNIM(ctx, nim)
	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data prestasi",
		)
	}

	return helper.OK(
		c,
		"data prestasi berhasil diambil",
		prestasiList,
	)
}
